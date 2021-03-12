package direktiv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/appc/spec/pkg/device"
	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/vorteil/pkg/vimg"
	"gopkg.in/freddierice/go-losetup.v1"
)

const (
	uid = 1000
	gid = 1000

	fc        = "/usr/local/bin/firecracker"
	jailerDir = "/srv/jailer"

	dataDiskSize = "+16 MiB"

	downloadPath = "https://downloads.vorteil.io/firecracker-vmlinux"
)

// mutex for loop devices
var loopMtx sync.Mutex

func (is *isolateServer) buildDataDisk(name string, data []byte, nws networkSetting) (string, error) {

	fpath := filepath.Join(os.TempDir(), fmt.Sprintf("data%s.raw", name))
	log.Debugf("building data disk %s", fpath)

	// directory to build files in
	dir, err := ioutil.TempDir("", "direktivdata")
	if err != nil {
		return fpath, err
	}
	defer os.RemoveAll(dir)

	// the actual disk
	file, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fpath, err
	}
	defer file.Close()

	// create data.in
	ioutil.WriteFile(filepath.Join(dir, direktivData), data, 0755)

	b, err := json.Marshal(nws)
	if err != nil {
		return fpath, err
	}
	ioutil.WriteFile(filepath.Join(dir, "network.in"), b, 0755)

	err = compileDataDisk(dir, fpath, dataDiskSize)
	if err != nil {
		return fpath, err
	}

	file.Sync()
	err = os.Chown(fpath, uid, gid)
	if err != nil {
		return fpath, err
	}

	return fpath, nil

}

type cowDisk struct {
	devRoot, devCow    losetup.Device
	cowDisk, finalDisk string
}

func cleanupCOW(name string, cd cowDisk) {

	cleanFile := func(path string) {
		if len(path) > 0 {
			os.Remove(path)
		}
	}

	detachFile := func(dev losetup.Device) {
		if len(dev.Path()) > 0 {
			dev.Detach()
		}
	}

	detachFile(cd.devRoot)
	detachFile(cd.devCow)

	cmd := exec.Command("dmsetup", "remove", "-f", name)
	cmd.Run()

	cleanFile(cd.finalDisk)
	cleanFile(cd.cowDisk)

}

// this takes raw disk and cow disk and does losetup & dmsetup
// firecracker jailer can not link it so we do a mknod in /tmp and use
// this in jailer
func createCOWDisk(name, disk string) (cowDisk, error) {

	var (
		cd  cowDisk
		err error
		c   *os.File
	)

	// mutex loop devices
	loopMtx.Lock()
	defer loopMtx.Unlock()

	// we need to cleanup if something fails
	defer func() {
		if err != nil {
			log.Errorf("error building COW disk: %v", err)
			cleanupCOW(name, cd)
		}
	}()

	cd.cowDisk = filepath.Join(os.TempDir(), fmt.Sprintf("%s.cow", name))

	log.Debugf("create cow disk: %v, %v", name, disk)

	// create empty file
	c, err = os.Create(cd.cowDisk)
	if err != nil {
		return cd, err
	}

	fi, err := os.Stat(disk)
	if err != nil {
		return cd, err
	}

	cowDiskSize := fi.Size() + (1024 * 1024 * 64)

	if err := c.Truncate(int64(cowDiskSize)); err != nil {
		return cd, err
	}
	os.Chmod(cd.cowDisk, 0777)

	// attach losetup
	cd.devRoot, err = losetup.Attach(disk, 0, true)
	if err != nil {
		log.Errorf("error loop disk: %v", err)
		return cd, err
	}
	log.Debugf("loop attached %v", cd.devRoot.Path())

	_, err = os.Stat(cd.cowDisk)
	if err != nil {
		return cd, err
	}

	cd.devCow, err = losetup.Attach(cd.cowDisk, 0, false)
	if err != nil {
		log.Errorf("error loop cow: %v", err)
		return cd, err
	}
	log.Debugf("cow attached %v", cd.devRoot.Path())

	log.Debugf("disks %s: %s, %s", name, cd.devRoot, cd.devCow)

	log.Debugf("create devmapper %s", name)
	cmd := exec.Command("dmsetup", "create", name, "--table", fmt.Sprintf("0 %d snapshot %s %s p 16", fi.Size()/512, cd.devRoot.Path(), cd.devCow.Path()))
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return cd, err
	}

	cow := fmt.Sprintf("/dev/mapper/%s", name)
	gg, err := os.Stat(cow)
	if err != nil {
		return cd, err
	}

	cd.finalDisk = fmt.Sprintf("/tmp/%s.raw", name)

	sys, ok := gg.Sys().(*syscall.Stat_t)
	if ok {

		dev := device.Makedev(device.Major(sys.Rdev), device.Minor(sys.Rdev))
		mode := uint32(fi.Mode()) | syscall.S_IFBLK
		if err := syscall.Mknod(cd.finalDisk, mode, int(dev)); err != nil {
			return cd, err
		}

		os.Chmod(cd.finalDisk, 0777)
	}

	return cd, err
}

func (is *isolateServer) runFirecracker(ctx context.Context, name, disk, dataDisk string, size int32) error {

	log.Debugf("run firecracker vm with %s, %s", disk, dataDisk)

	d, err := createCOWDisk(name, disk)
	if err != nil {
		return err
	}

	// cleanup
	defer func() {

		loopMtx.Lock()
		cleanupCOW(name, d)
		loopMtx.Unlock()

		// remove jailer files
		jailerFiles := filepath.Join("/srv/jailer/firecracker/", name)
		os.RemoveAll(jailerFiles)
	}()

	// boot disk
	rootDrive := models.Drive{
		DriveID:      firecracker.String("1"),
		PathOnHost:   &d.finalDisk,
		IsRootDevice: firecracker.Bool(true),
		IsReadOnly:   firecracker.Bool(false),
		Partuuid:     vimg.Part2UUIDString,
	}

	// data disk
	secondDrive := models.Drive{
		DriveID:      firecracker.String("2"),
		PathOnHost:   &dataDisk,
		IsRootDevice: firecracker.Bool(false),
		IsReadOnly:   firecracker.Bool(false),
	}

	devices := []models.Drive{
		rootDrive, secondDrive,
	}

	networkIfaces := []firecracker.NetworkInterface{{
		StaticConfiguration: &firecracker.StaticNetworkConfiguration{
			HostDevName: "tap0",
		},
	}}

	kf := fmt.Sprintf("/tmp/kernel-%s", is.config.Kernel.Linux)

	// download kernel
	if _, err := os.Stat(kf); os.IsNotExist(err) {

		log.Debugf("downloading %s", fmt.Sprintf("%s/firecracker-%s", downloadPath, is.config.Kernel.Linux))
		resp, err := http.Get(fmt.Sprintf("%s/firecracker-%s", downloadPath, is.config.Kernel.Linux))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		out, err := os.Create(kf)
		if err != nil {
			return err
		}
		defer out.Close()
		io.Copy(out, resp.Body)

	}

	var (
		cpu, mem *int64
	)

	switch size {
	case 1:
		cpu = firecracker.Int64(1)
		mem = firecracker.Int64(512)
	case 2:
		cpu = firecracker.Int64(2)
		mem = firecracker.Int64(1024)
	default:
		cpu = firecracker.Int64(1)
		mem = firecracker.Int64(256)
	}

	log.Debugf("firecracker using %d cpu, %d ram", *cpu, *mem)

	fcConf := firecracker.Config{
		SocketPath:      fmt.Sprintf("fcsock%v.sock", name),
		KernelImagePath: kf,
		KernelArgs:      fmt.Sprintf("init=/vorteil/vinitd rw console=ttyS0 loglevel=2 reboot=k panic=1 pci=off i8042.noaux i8042.nomux i8042.nopnp i8042.dumbkbd vt.color=0x00 random.trust_cpu=on root=PARTUUID=%s direktiv", vimg.Part2UUIDString),
		Drives:          devices,
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  cpu,
			HtEnabled:  firecracker.Bool(false),
			MemSizeMib: mem,
		},
		LogLevel:          "warn",
		ForwardSignals:    []os.Signal{},
		NetNS:             fmt.Sprintf("/var/run/netns/%s", name),
		NetworkInterfaces: networkIfaces,
		JailerCfg: &firecracker.JailerConfig{
			ChrootBaseDir:  jailerDir,
			ID:             name,
			UID:            firecracker.Int(0),
			GID:            firecracker.Int(0),
			Stdout:         os.Stdout,
			Stderr:         os.Stdout,
			NumaNode:       firecracker.Int(0),
			ChrootStrategy: firecracker.NewNaiveChrootStrategy("linux"),
			ExecFile:       fc,
		},
	}

	fclog := log.New()
	fclog.SetLevel(log.WarnLevel)

	machine, err := firecracker.NewMachine(ctx, fcConf, firecracker.WithLogger(fclog.WithField("fc", name)))
	if err != nil {
		return err
	}

	if err := machine.Start(ctx); err != nil {
		return err
	}

	return machine.Wait(ctx)

}
