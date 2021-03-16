package direktiv

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	parser "github.com/novln/docker-parser"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/isolate"

	log "github.com/sirupsen/logrus"
)

func loginIfRequired(img string, registries map[string]string) string {

	// login if required
	r, _ := parser.Parse(img)

	if val, ok := registries[r.Registry()]; ok {

		file, err := ioutil.TempFile("", "auth")
		if err != nil {
			log.Errorf("can not create auth file: %v", err)
			return ""
		}

		ss := strings.SplitAfterN(val, "!", 2)
		if len(ss) != 2 {
			log.Errorf("authentication for registry invalid")
			return ""
		}

		// remove ! from username
		user := ss[0][0 : len(ss[0])-1]
		token := ss[1]

		encAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, token)))

		d := fmt.Sprintf(`{
			"auths": {
				"%s": {
					"auth": "%s"
				}
			}
			}`, r.Registry(), encAuth)

		file.WriteString(d)

		return file.Name()
	}

	return ""
}

func (is *isolateServer) runAsContainer(img, cmd, isolateID string, in *isolate.RunIsolateRequest, log15log dlog.Logger) {

	var data, din []byte

	log.Debugf("run action %s as container (%s)", in.GetInstanceId(), in.GetActionId())

	serr := func(err error, errCode string) *IsolateError {
		ae := IsolateError{
			ErrorMessage: err.Error(),
			ErrorCode:    errCode,
		}
		return &ae
	}

	var authFile string

	// login
	if in.GetRegistries() != nil {
		authFile = loginIfRequired(img, in.GetRegistries())
	}

	// create data-dir
	dir, err := ioutil.TempDir(os.TempDir(), in.GetActionId())
	if err != nil {
		is.respondToAction(serr(err, errorImage), data, in)
		return
	}

	log.Debugf("container data-dir: %v", dir)

	stdout, err := ioutil.TempFile("", "stdout")
	if err != nil {
		is.respondToAction(serr(err, errorInternal), data, in)
		return
	}
	stderr, err := ioutil.TempFile("", "stderr")
	if err != nil {
		is.respondToAction(serr(err, errorInternal), data, in)
		return
	}

	defer func() {
		os.RemoveAll(dir)
		if len(authFile) > 0 {
			os.Remove(authFile)
		}
		os.Remove(stdout.Name())
		os.Remove(stderr.Name())
	}()

	// write file to data dir
	err = ioutil.WriteFile(filepath.Join(dir, direktivData), in.GetData(), 0755)
	if err != nil {
		log.Errorf("can not write direktiv data for container: %v", err)
		is.respondToAction(serr(err, errorInternal), data, in)
		return
	}

	ctxs := is.addCtx(in.Timeout, isolateID)
	defer is.finishCancelIsolate(isolateID)

	args := []string{
		"run",
		"--volume",
		fmt.Sprintf("%s:%s", dir, direktivDir),
		"--storage-driver=vfs",
	}

	if len(authFile) > 0 {
		args = append(args, fmt.Sprintf("--authfile=%s", authFile))
	}

	switch in.GetSize() {
	case 1:
		args = append(args, "-m=512m")
		args = append(args, "--cpus=1")
	case 2:
		args = append(args, "-m=1024m")
		args = append(args, "--cpus=2")
	default:
		args = append(args, "-m=256m")
		args = append(args, "--cpus=1")
	}

	// img and command at the end
	args = append(args, img)
	if len(cmd) > 0 {
		args = append(args, cmd)
	}

	log.Debugf("run container %v with command %v", img, cmd)

	podman := exec.CommandContext(ctxs.ctx, "podman", args...)
	podman.Stdout = io.MultiWriter(os.Stdout, stdout)
	podman.Stderr = io.MultiWriter(os.Stdout, stderr)

	log.Debugf("podman cmd: %v", podman)

	err = podman.Run()
	if err != nil {
		log.Errorf("error executing container: %v", err)
		is.respondToAction(serr(err, errorInternal), data, in)
		return
	}

	// log output
	stdo, _ := ioutil.ReadFile(stdout.Name())
	stde, _ := ioutil.ReadFile(stderr.Name())

	if len(stdo) > 0 {
		log15log.Info(string(stdo))
	}
	if len(stde) > 0 {
		log15log.Error(string(stde))
	}

	log.Debugf("stdout: %v", string(stdo))
	log.Debugf("stderr: %v", string(stde))

	// if error.json, we use this and report error
	if _, err := os.Stat(filepath.Join(dir, "error.json")); !os.IsNotExist(err) {
		var ae IsolateError
		err := json.Unmarshal(din, &ae)
		if err != nil {
			log15log.Error(fmt.Sprintf("error parsing error file: %v", err))
			is.respondToAction(serr(fmt.Errorf("%w; %s", err, string(din)), errorIO), data, in)
			return
		}

		log15log.Error(ae.ErrorMessage)
		is.respondToAction(&ae, data, in)
		return
	}

	// can not do much if that fails, print to logs, otherwise we return the data
	data, err = ioutil.ReadFile(filepath.Join(dir, "data.out"))
	if err != nil {
		log15log.Error(fmt.Sprintf("error parsing data file: %v", err))
		is.respondToAction(serr(err, errorIO), data, in)
		return
	}

	go func() {
		maxlen := math.Min(256, float64(len(data)))
		log.Debugf("responding to isolate caller: %v", string(data[0:int(maxlen)]))
		is.respondToAction(nil, data, in)
	}()

}
