package direktiv

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"

	"github.com/vorteil/vorteil/pkg/elog"
	"github.com/vorteil/vorteil/pkg/vcfg"
	"github.com/vorteil/vorteil/pkg/vdisk"
	"github.com/vorteil/vorteil/pkg/vimg"
	"github.com/vorteil/vorteil/pkg/vio"
)

// PartitionUUID of the disk
var PartitionUUID = []byte{
	0x7d, 0x44, 0x48, 0x40,
	0x9d, 0xc0, 0x11, 0xd1,
	0xb2, 0x45, 0x5f, 0xfd,
	0xce, 0x74, 0xfa, 0xd3,
}

// PartitionName of the disk
var PartitionName = []byte{0x76, 0x0, 0x6f, 0x0, 0x72, 0x0, 0x74, 0x0, 0x65, 0x0, 0x69, 0x0,
	0x6c, 0x0, 0x2d, 0x0, 0x72, 0x0, 0x6f, 0x0, 0x6f, 0x0, 0x74, 0x0} // "vorteil-root" in utf16

// compileDataDisk builds the data disk
func compileDataDisk(dir, output, size string) error {

	imgSize, err := vcfg.ParseBytes(size)
	if err != nil {
		return err
	}

	log := &elog.CLI{
		DisableTTY: true,
	}

	fsTree, err := vio.FileTreeFromDirectory(dir)
	if err != nil {
		return err
	}

	fsCompiler, err := vdisk.NewFilesystemCompiler("ext2", log, fsTree, nil)
	if err != nil {
		return err
	}

	fsCompiler.SetMinimumInodes(1024)
	fsCompiler.SetMinimumInodesPer64MiB(256)
	fsCompiler.IncreaseMinimumInodes(1024)

	if imgSize.IsDelta() {
		delta := vcfg.Bytes(0)
		delta.ApplyDelta(imgSize)
		fsCompiler.IncreaseMinimumFreeSpace(int64(delta))
	} else {
		fsCompiler.IncreaseMinimumFreeSpace(16 * 0x100000) // 16 MiB
	}

	err = fsCompiler.Commit(context.Background())
	if err != nil {
		return err
	}

	minSize := fsCompiler.MinimumSize()
	fsSize := minSize
	if !imgSize.IsDelta() {
		gptSize := int64(vimg.P0FirstLBA+33) * vimg.SectorSize
		fsSize = int64(imgSize) - gptSize
	} else {
		imgSize = vcfg.Bytes(fsSize + int64(vimg.P0FirstLBA+33)*vimg.SectorSize)
	}

	err = fsCompiler.Precompile(context.Background(), fsSize)
	if err != nil {
		return err
	}

	gpt, err := GenerateGPT(imgSize)
	if err != nil {
		return err
	}

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	var ws io.WriteSeeker

	// write GPT stuff
	err = gpt.WritePrimary(f)
	if err != nil {
		goto cleanup
	}

	// write filesystem partition
	_, err = f.Seek(vimg.P0FirstLBA*vimg.SectorSize, io.SeekStart)
	if err != nil {
		goto cleanup
	}

	ws, err = vio.WriteSeeker(f)
	if err != nil {
		goto cleanup
	}

	err = fsCompiler.Compile(context.Background(), ws)
	if err != nil {
		goto cleanup
	}

	// write redundant GPT stuff
	_, err = f.Seek(int64(imgSize)-(33*vimg.SectorSize), io.SeekStart)
	if err != nil {
		goto cleanup
	}

	err = gpt.WriteBackup(f)
	if err != nil {
		goto cleanup
	}

	return nil

cleanup:
	_ = f.Close()
	_ = os.Remove(output)
	return err

}

type GPT struct {
	DiskSize               int64
	diskUID                []byte
	gptEntries             []byte
	gptEntriesCRC          uint32
	lastUsableLBA          uint64
	secondaryGPTHeaderLBA  uint64
	secondaryGPTEntriesLBA uint64
}

func GenerateGPT(size vcfg.Bytes) (*GPT, error) {

	x := size.Units(vcfg.Bytes(vimg.SectorSize))

	gpt := &GPT{
		DiskSize:               int64(size.Units(vcfg.Byte)),
		secondaryGPTHeaderLBA:  uint64(x - 1),
		secondaryGPTEntriesLBA: uint64(x - 33),
		lastUsableLBA:          uint64(x - 34),
	}

	err := gpt.generateGPTEntries()
	if err != nil {
		return nil, err
	}

	return gpt, nil

}

func (gpt *GPT) generateUID() ([]byte, error) {

	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	// NOTE: I wrote this a long time ago and cannot explain why I do these
	// bitwise operations...
	buf[6] = buf[6]&^0xf0 | 0x40
	buf[8] = buf[8]&^0xc0 | 0x80

	return buf, nil

}

func (gpt *GPT) generateGPTEntries() error {

	// var err error
	// gpt.diskUID, err = gpt.generateUID()
	gpt.diskUID = PartitionUUID

	// if err != nil {
	// 	return err
	// }

	p0 := vimg.GPTEntry{
		TypeGUID: [16]byte{0xE3, 0xBC, 0x68, 0x4F, 0xCD, 0xE8,
			0xB1, 0x4D, 0x96, 0xE7, 0xFB, 0xCA, 0xF9, 0x84, 0xB7, 0x09}, // Linux x86-64 root filesystem partition
		FirstLBA: uint64(vimg.P0FirstLBA),
		LastLBA:  uint64(gpt.lastUsableLBA),
	}

	copy(p0.PartitionGUID[:], PartitionUUID)
	copy(p0.Name[:], PartitionName)

	entriesBuffer := new(bytes.Buffer)
	_ = binary.Write(entriesBuffer, binary.LittleEndian, p0)

	gpt.gptEntries = entriesBuffer.Bytes()

	crc := crc32.NewIEEE()
	_, _ = io.Copy(crc, bytes.NewReader(gpt.gptEntries))
	_, _ = io.CopyN(crc, vio.Zeroes, vimg.MaximumGPTEntries*vimg.GPTEntrySize-int64(len(gpt.gptEntries)))
	gpt.gptEntriesCRC = crc.Sum32()

	return nil
}

func (gpt *GPT) WritePrimary(w io.Writer) error {

	// mbr

	mbr := &vimg.ProtectiveMBR{
		Status:        0x7F,
		PartitionType: 0xEE,
		FirstLBA:      1,
		MagicNumber:   [2]byte{0x55, 0xAA},
		TotalSectors:  uint32(gpt.DiskSize/vimg.SectorSize) - 1,
	}

	err := binary.Write(w, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	// gpt header

	hdr := vimg.GPTHeader{
		Signature:      vimg.GPTSignature,
		Revision:       [4]byte{0, 0, 1, 0},
		HeaderSize:     vimg.GPTHeaderSize,
		CurrentLBA:     vimg.PrimaryGPTHeaderLBA,
		BackupLBA:      uint64(gpt.secondaryGPTHeaderLBA),
		FirstUsableLBA: vimg.P0FirstLBA,
		LastUsableLBA:  uint64(gpt.lastUsableLBA),
		StartLBAParts:  vimg.PrimaryGPTEntriesLBA,
		NoOfParts:      vimg.MaximumGPTEntries,
		SizePartEntry:  vimg.GPTEntrySize,
		CRCParts:       gpt.gptEntriesCRC,
	}

	copy(hdr.GUID[:], gpt.diskUID)

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, hdr)

	crc := crc32.NewIEEE()
	_, _ = io.CopyN(crc, bytes.NewReader(buf.Bytes()), vimg.GPTHeaderSize)

	hdr.CRC = crc.Sum32()
	err = binary.Write(w, binary.LittleEndian, hdr)
	if err != nil {
		return err
	}

	// entries

	k, err := io.Copy(w, bytes.NewReader(gpt.gptEntries))
	if err != nil {
		return err
	}

	_, err = io.CopyN(w, vio.Zeroes, vimg.SectorSize*vimg.GPTEntriesSectors-k)
	if err != nil {
		return err
	}

	return nil

}

func (gpt *GPT) WriteBackup(w io.Writer) error {

	// entries

	k, err := io.Copy(w, bytes.NewReader(gpt.gptEntries))
	if err != nil {
		return err
	}

	_, err = io.CopyN(w, vio.Zeroes, vimg.SectorSize*vimg.GPTEntriesSectors-k)
	if err != nil {
		return err
	}

	// header

	hdr := vimg.GPTHeader{
		Signature:      vimg.GPTSignature,
		Revision:       [4]byte{0, 0, 1, 0},
		HeaderSize:     vimg.GPTHeaderSize,
		CurrentLBA:     uint64(gpt.secondaryGPTHeaderLBA),
		BackupLBA:      vimg.PrimaryGPTHeaderLBA,
		FirstUsableLBA: vimg.P0FirstLBA,
		LastUsableLBA:  uint64(gpt.lastUsableLBA),
		StartLBAParts:  uint64(gpt.secondaryGPTEntriesLBA),
		NoOfParts:      vimg.MaximumGPTEntries,
		SizePartEntry:  vimg.GPTEntrySize,
		CRCParts:       gpt.gptEntriesCRC,
	}

	copy(hdr.GUID[:], gpt.diskUID)

	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, hdr)

	crc := crc32.NewIEEE()
	_, _ = io.CopyN(crc, bytes.NewReader(buf.Bytes()), vimg.GPTHeaderSize)

	hdr.CRC = crc.Sum32()
	err = binary.Write(w, binary.LittleEndian, hdr)
	if err != nil {
		return err
	}

	return nil

}
