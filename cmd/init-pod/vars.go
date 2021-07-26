package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vorteil/direktiv/pkg/flow"
	"google.golang.org/protobuf/types/known/emptypb"
)

type isolateFiles struct {
	Key   string `json:"key"`
	As    string `json:"as"`
	Scope string `json:"scope"`
	Type  string `json:"type"`
}

func loadFiles(r *http.Request) error {

	hdr := "Direktiv-Files"
	strs := r.Header.Values(hdr)

	var ifiles []*isolateFiles

	for i, s := range strs {

		data, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return fmt.Errorf("invalid %s [%d]: %v", hdr, i, err)
		}

		files := new(isolateFiles)
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		err = dec.Decode(files)
		if err != nil {
			return fmt.Errorf("invalid %s [%d]: %v", hdr, i, err)
		}

		// TODO: extra validation

		ifiles = append(ifiles, files)

	}

	err := prepIsolateFiles(r.Context(), ifiles)
	if err != nil {
		return err
	}

	return nil

}

func prepIsolateFiles(ctx context.Context, ifiles []*isolateFiles) error {

	dir := isolateDir()

	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}

	for i, f := range ifiles {
		err = prepOneIsolateFiles(ctx, f)
		if err != nil {
			return fmt.Errorf("failed to prepare isolate files %d: %v", i, err)
		}
	}

	subDirs := []string{"namespace", "workflow", "instance"}
	for _, d := range subDirs {
		err := os.MkdirAll(path.Join(dir, fmt.Sprintf("out/%s", d)), 0750)
		if err != nil {
			return fmt.Errorf("failed to prepare isolate output dirs: %v", err)
		}
	}

	return nil

}

func prepOneIsolateFiles(ctx context.Context, f *isolateFiles) error {

	pr, pw := io.Pipe()

	go func() {
		err := fileReader(ctx, f, pw)
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			_ = pw.Close()
		}
	}()

	err := fileWriter(ctx, f, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return err
	}

	_ = pr.Close()

	return nil

}

func untarFile(tr *tar.Reader, path string) error {

	pdir, _ := filepath.Split(path)

	/* #nosec */
	err := os.MkdirAll(pdir, 0750)
	if err != nil {
		return err
	}

	/* #nosec */
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	/* #nosec */
	defer f.Close()

	_, err = io.Copy(f, tr)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil

}

func untar(dst string, r io.Reader) error {

	/* #nosec */
	err := os.MkdirAll(dst, 0750)
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)

	for {
		/* #nosec */
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		hdr.Name = filepath.Clean(hdr.Name)
		if strings.Contains(hdr.Name, "..") {
			return errors.New("zip-slip")
		}

		/* #nosec */
		path := filepath.Join(dst, hdr.Name)

		if hdr.Typeflag == tar.TypeReg {
			err = untarFile(tr, path)
			if err != nil {
				return err
			}
		} else if hdr.Typeflag == tar.TypeDir {
			/* #nosec */
			err = os.MkdirAll(path, 0750)
			if err != nil {
				return err
			}
		} else {
			return errors.New("unsupported tar archive contents")
		}

	}

	return nil

}

func writeFile(dst string, r io.Reader) error {

	/* #nosec */
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	/* #nosec */
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil

}

func writeAnyFile(ftype, dst string, pr io.Reader) error {

	// TODO: const the types

	var err error

	switch ftype {

	case "":
		fallthrough

	case "plain":
		err = writeFile(dst, pr)
		if err != nil {
			return err
		}

	case "base64":
		r := base64.NewDecoder(base64.StdEncoding, pr)
		err = writeFile(dst, r)
		if err != nil {
			return err
		}

	case "tar":
		err = untar(dst, pr)
		if err != nil {
			return err
		}

	case "tar.gz":
		gr, err := gzip.NewReader(pr)
		if err != nil {
			return err
		}

		err = untar(dst, gr)
		if err != nil {
			return err
		}

		err = gr.Close()
		if err != nil {
			return err
		}

	default:
		panic(ftype)
	}

	return nil

}

func fileReader(ctx context.Context, f *isolateFiles, pw *io.PipeWriter) error {

	err := getVar(ctx, pw, nil, f.Scope, f.Key)
	if err != nil {
		return err
	}

	return nil

}

func fileWriter(ctx context.Context, f *isolateFiles, pr *io.PipeReader) error {

	// TODO: validate f.Type earlier so that the switch cannot get unexpected data here

	dir := isolateDir()
	dst := f.Key
	if f.As != "" {
		dst = f.As
	}
	dst = filepath.Join(dir, dst)
	dir, _ = filepath.Split(dst)

	/* #nosec */
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}

	err = writeAnyFile(f.Type, dst, pr)
	if err != nil {
		return err
	}

	return nil

}

func isolateDir() string {
	return "/direktiv-data/vars"
}

func getVar(ctx context.Context, w io.Writer, setTotalSize func(x int64), scope, key string) error {

	client, recv, err := requestVar(ctx, scope, key)
	if err != nil {
		return err
	}

	var received int64
	var noEOF = true
	for noEOF {
		msg, err := recv()
		if err == io.EOF {
			noEOF = false
		} else if err != nil {
			return err
		}

		if msg == nil {
			continue
		}

		totalSize := msg.GetTotalSize()

		if setTotalSize != nil {
			setTotalSize(totalSize)
			setTotalSize = nil
		}

		data := msg.GetValue()
		received += int64(len(data))

		if received > totalSize {
			return errors.New("variable returned too many bytes")
		}

		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			return err
		}

		if totalSize == received {
			break
		}
	}

	err = client.CloseSend()
	if err != nil && err != io.EOF {
		return err
	}

	return nil

}

type varClient interface {
	CloseSend() error
}

type varClientMsg interface {
	GetTotalSize() int64
	GetChunkSize() int64
	GetValue() []byte
}

func requestVar(ctx context.Context, scope, key string) (client varClient, recv func() (varClientMsg, error), err error) {

	// TODO: const the scopes
	// TODO: validate scope earlier so that the switch cannot get unexpected data here
	// TODO: log missing files but proceed anyway

	switch scope {

	case "namespace":
		var nvClient flow.DirektivFlow_GetNamespaceVariableClient
		nvClient, err = flowClient.GetNamespaceVariable(ctx, &flow.GetNamespaceVariableRequest{
			InstanceId: &instanceId,
			Key:        &key,
		})
		client = nvClient
		recv = func() (varClientMsg, error) {
			return nvClient.Recv()
		}

	case "workflow":
		var wvClient flow.DirektivFlow_GetWorkflowVariableClient
		wvClient, err = flowClient.GetWorkflowVariable(ctx, &flow.GetWorkflowVariableRequest{
			InstanceId: &instanceId,
			Key:        &key,
		})
		client = wvClient
		recv = func() (varClientMsg, error) {
			return wvClient.Recv()
		}

	case "":
		fallthrough

	case "instance":
		var ivClient flow.DirektivFlow_GetInstanceVariableClient
		ivClient, err = flowClient.GetInstanceVariable(ctx, &flow.GetInstanceVariableRequest{
			InstanceId: &instanceId,
			Key:        &key,
		})
		client = ivClient
		recv = func() (varClientMsg, error) {
			return ivClient.Recv()
		}

	default:
		panic(scope)
	}

	return

}

func setOutVariables(ctx context.Context) error {

	subDirs := []string{"namespace", "workflow", "instance"}
	for _, d := range subDirs {

		out := path.Join(isolateDir(), "out", d)

		files, err := ioutil.ReadDir(out)
		if err != nil {
			return fmt.Errorf("can not read out folder: %v", err)
		}

		for _, f := range files {

			fp := path.Join(isolateDir(), "out", d, f.Name())

			switch mode := f.Mode(); {
			case mode.IsDir():

				tf, err := ioutil.TempFile("", "outtar")
				if err != nil {
					return err
				}

				err = tarGzDir(fp, tf)
				if err != nil {
					return err
				}
				defer os.Remove(tf.Name())

				var end int64
				end, err = tf.Seek(0, io.SeekEnd)
				if err != nil {
					return err
				}

				_, err = tf.Seek(0, io.SeekStart)
				if err != nil {
					return err
				}

				err = setVar(ctx, end, tf, d, f.Name())
				if err != nil {
					return err
				}
			case mode.IsRegular():

				/* #nosec */
				v, err := os.Open(fp)
				if err != nil {
					return err
				}

				err = setVar(ctx, f.Size(), v, d, f.Name())
				if err != nil {
					_ = v.Close()
					return err
				}

				err = v.Close()
				if err != nil {
					return err
				}

			}

		}

	}

	return nil
}

func tarGzDir(src string, buf io.Writer) error {

	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		if !fi.Mode().IsDir() && !fi.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// use "subpath"
		header.Name = filepath.ToSlash(file[len(src):])

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			/* #nosec */
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zr.Close(); err != nil {
		return err
	}

	return nil
}

type varSetClient interface {
	CloseAndRecv() (*emptypb.Empty, error)
}

type varSetClientMsg struct {
	Key        *string
	InstanceId *string
	Value      []byte
	TotalSize  *int64
	ChunkSize  *int64
}

func setVar(ctx context.Context, totalSize int64, r io.Reader, scope, key string) error {

	// TODO: const the scopes
	// TODO: validate scope earlier so that the switch cannot get unexpected data here
	// TODO: log missing files but proceed anyway

	var err error
	var client varSetClient
	var send func(*varSetClientMsg) error

	switch scope {

	case "namespace":
		var nvClient flow.DirektivFlow_SetNamespaceVariableClient
		nvClient, err = flowClient.SetNamespaceVariable(ctx)
		client = nvClient
		send = func(x *varSetClientMsg) error {
			req := &flow.SetNamespaceVariableRequest{}
			req.Key = x.Key
			req.InstanceId = x.InstanceId
			req.TotalSize = x.TotalSize
			req.Value = x.Value
			req.ChunkSize = x.ChunkSize
			return nvClient.Send(req)
		}

	case "workflow":
		var wvClient flow.DirektivFlow_SetWorkflowVariableClient
		wvClient, err = flowClient.SetWorkflowVariable(ctx)
		client = wvClient
		send = func(x *varSetClientMsg) error {
			req := &flow.SetWorkflowVariableRequest{}
			req.Key = x.Key
			req.InstanceId = x.InstanceId
			req.TotalSize = x.TotalSize
			req.Value = x.Value
			req.ChunkSize = x.ChunkSize
			return wvClient.Send(req)
		}

	case "":
		fallthrough

	case "instance":
		var ivClient flow.DirektivFlow_SetInstanceVariableClient
		ivClient, err = flowClient.SetInstanceVariable(ctx)
		client = ivClient
		send = func(x *varSetClientMsg) error {
			req := &flow.SetInstanceVariableRequest{}
			req.Key = x.Key
			req.InstanceId = x.InstanceId
			req.TotalSize = x.TotalSize
			req.Value = x.Value
			req.ChunkSize = x.ChunkSize
			return ivClient.Send(req)
		}

	default:
		panic(scope)
	}

	chunkSize := int64(0x200000) // 2 MiB
	if totalSize <= 0 {
		buf := new(bytes.Buffer)
		_, err := io.CopyN(buf, r, chunkSize+1)
		if err == nil {
			return errors.New("large payload requires defined Content-Length")
		}
		if err != io.EOF {
			return err
		}

		data := buf.Bytes()
		r = bytes.NewReader(data)
		totalSize = int64(len(data))
	}

	var written int64
	for {
		chunk := chunkSize
		if totalSize-written < chunk {
			chunk = totalSize - written
		}

		buf := new(bytes.Buffer)
		k, err := io.CopyN(buf, r, chunk)
		if err != nil {
			return err
		}

		written += k

		err = send(&varSetClientMsg{
			TotalSize:  &totalSize,
			ChunkSize:  &chunkSize,
			Key:        &key,
			InstanceId: &instanceId,
			Value:      buf.Bytes(),
		})
		if err != nil {
			return err
		}

		if written == totalSize {
			break
		}
	}

	_, err = client.CloseAndRecv()
	if err != nil && err != io.EOF {
		return err
	}

	return nil

}
