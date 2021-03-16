package direktiv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codeclysm/extract"
	"github.com/vorteil/vorteil/pkg/elog"
	"github.com/vorteil/vorteil/pkg/vcfg"
	"github.com/vorteil/vorteil/pkg/vdisk"
	"github.com/vorteil/vorteil/pkg/vpkg"
	"github.com/vorteil/vorteil/pkg/vproj"

	"github.com/containers/image/manifest"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
)

const (
	fullMask     = "255.255.255.255"
	direktivDir  = "/direktiv-data"
	direktivData = "data.in"
)

var (
	errorLog = filepath.Join(direktivDir, "error.log")
	logLog   = filepath.Join(direktivDir, "log.log")

	defaultVorteilProjectConf = `ignore = [".vorteilproject"]
[[target]]
	name = "default"
	vcfgs = ["default.vcfg"]
`
)

func extractContainer(image string, img v1.Image) (string, error) {

	f, err := ioutil.TempFile("", "imgdl*.tar")
	if err != nil {
		return "", err
	}

	log.Debugf("exporting image %v to %s", image, f.Name())
	fs := mutate.Extract(img)
	_, err = io.Copy(f, fs)
	if err != nil {
		return "", err
	}

	d, err := ioutil.TempDir("", "imgdir")
	if err != nil {
		return "", err
	}

	log.Debugf("untar image %v to %s", image, d)
	f.Seek(0, io.SeekStart)
	err = extract.Tar(context.Background(), f, d, nil)
	if err != nil {
		return "", err
	}
	log.Debugf("untar image %v finished", image)

	return d, nil
}

func buildImageDisk(image, cmd, kernel, path string,
	registries map[string]string) (string, error) {

	if len(image) == 0 {
		return "", fmt.Errorf("image can not be empty")
	}

	// we have parsed it already, no error
	var err error
	var ref name.Reference

	// this function can panic...
	func() {
		defer func() {
			r := recover()
			if r != nil && err == nil {
				err = errors.New("unable to parse image reference")
			}
		}()

		ref, err = name.ParseReference(image)
	}()

	if err != nil {
		return "", err
	}

	// authenticate if there are registries for this images
	opts := findAuthForRegistry(image, registries)
	img, err := remote.Image(ref, opts...)
	if err != nil {
		return "", err
	}

	d, err := extractContainer(image, img)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(d)

	config, err := getContainerConfig(img)
	if err != nil {
		return "", err
	}

	vcfg, err := writeVCFG(d, cmd, kernel, config)
	if err != nil {
		return "", err
	}

	// build raw disk
	return buildRaw(d, filepath.Join(cacheDir, hashImg(image, cmd)), vcfg)

}

func baseVCFG() *vcfg.VCFG {

	vcfgFile := new(vcfg.VCFG)
	vcfgFile.Programs = make([]vcfg.Program, 1)
	vcfgFile.Networks = make([]vcfg.NetworkInterface, 1)
	vcfgFile.Programs[0].Stdout = logLog
	vcfgFile.Programs[0].Stderr = errorLog

	vcfgFile.Networks[0].IP = fullMask
	vcfgFile.Networks[0].Mask = fullMask
	vcfgFile.Networks[0].Gateway = fullMask

	ram, _ := vcfg.ParseBytes("64 MiB")
	vcfgFile.VM.RAM = ram

	ds, _ := vcfg.ParseBytes("+64 MiB")
	vcfgFile.VM.DiskSize = ds

	// just a dummy to start the ntp process in vinitd
	vcfgFile.System.NTP = []string{"0.au.pool.ntp.org"}

	return vcfgFile

}

func writeVCFG(dir, cmd, kernel string, config map[string]interface{}) (*vcfg.VCFG, error) {

	// vcfgFile := new(vcfg.VCFG)
	// vcfgFile.Programs = make([]vcfg.Program, 1)
	// vcfgFile.Networks = make([]vcfg.NetworkInterface, 1)

	vcfgFile := baseVCFG()

	// get configuration
	var (
		ci manifest.Schema2Config
	)

	sb, err := json.Marshal(config["config"])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(sb, &ci)
	if err != nil {
		return nil, err
	}

	vcfgFile.VM.Kernel = kernel

	vcfgFile.Programs[0].Cwd = ci.WorkingDir
	vcfgFile.Programs[0].Env = ci.Env

	vcfgFile.Programs[0].Env = append(vcfgFile.Programs[0].Env,
		"DIREKTIV_DIR=/direktiv_data")

	log.Debugf("env for app: %v", vcfgFile.Programs[0].Env)

	var p string
	if len(cmd) > 0 {
		log.Debugf("command provided: %v", cmd)
		s := strings.Split(p, " ")
		_, err := findBinary(s[0], ci.Env, ci.WorkingDir, dir)
		if err != nil {
			log.Errorf("can not find executable: %s", s[0])
			return nil, err
		}
	} else {
		log.Debugf("command not provided, searching in %s", dir)
		// if no command provided, we build it based on manifest
		p, err = buildCommand(fmt.Sprintf("%s", dir), ci)
		if err != nil {
			log.Errorf("can not find executable: %s", p)
			return nil, err
		}
	}

	log.Debugf("using %s command", p)
	vcfgFile.Programs[0].Args = p

	// generate vmdk and store it
	b, err := vcfgFile.Marshal()
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/default.vcfg", dir), b, 0644)
	if err != nil {
		return nil, err
	}

	// create chrony file. uses ptp0 instead of remote NTP
	os.Mkdir(fmt.Sprintf("%s/etc", dir), 0755)
	ntp := "refclock PHC /dev/ptp0 poll 3 dpoll -2 offset 0"
	err = ioutil.WriteFile(fmt.Sprintf("%s/etc/chrony.conf", dir), []byte(ntp), 0644)
	if err != nil {
		return nil, err
	}

	// this wil be used for mounting
	err = os.Mkdir(filepath.Join(dir, direktivDir), 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(filepath.Join(dir, ".vorteilproject"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.WriteString(defaultVorteilProjectConf)

	return vcfgFile, err

}

func getContainerConfig(img v1.Image) (map[string]interface{}, error) {

	var cjson map[string]interface{}
	cfg, err := img.RawConfigFile()
	if err != nil {
		return cjson, err
	}

	err = json.Unmarshal([]byte(cfg), &cjson)

	if err != nil {
		return cjson, err
	}

	return cjson, nil
}

func findBinary(name string, env []string,
	cwd string, targetDir string) (string, error) {

	if strings.HasPrefix(name, "./") {
		abs, err := filepath.Abs(name)
		if err != nil {
			return name, nil
		}
		cwd, err := os.Getwd()
		if err != nil {
			return name, nil
		}
		rel, err := filepath.Rel(cwd, abs)
		if err != nil {
			return name, nil
		}
		name = rel
	}

	// absolute
	if strings.HasPrefix(name, "/") {
		fp := filepath.Join(targetDir, name)
		if _, err := os.Lstat(fp); err == nil {
			return name, nil
		}
		return "", fmt.Errorf("can not find binary %s", name)
	}

	for _, e := range env {
		elems := strings.SplitN(e, "=", 2)
		if elems[0] == "PATH" {
			elems = strings.Split(elems[1], ":")
			for _, p := range elems {
				path := filepath.Join(targetDir, p, strings.ReplaceAll(name, "\"", ""))
				if _, err := os.Stat(path); err == nil {
					return filepath.Join(p, strings.ReplaceAll(name, "\"", "")), nil
				}
			}
		}
	}

	path := filepath.Join(targetDir, cwd, strings.ReplaceAll(name, "\"", ""))
	if _, err := os.Stat(path); err == nil {
		return filepath.Join(cwd, strings.ReplaceAll(name, "\"", "")), nil
	}

	return "", fmt.Errorf("can not find binary %s", name)
}

func buildCommand(dir string, ci manifest.Schema2Config) (string, error) {

	var finalCmd, args []string

	if len(ci.Entrypoint) > 0 {
		ss := []string(ci.Entrypoint)
		finalCmd = append(finalCmd, ss...)
	}

	if len(ci.Cmd) > 0 {
		finalCmd = append(finalCmd, ci.Cmd...)
	}

	bin, err := findBinary(finalCmd[0], ci.Env, ci.WorkingDir, dir)
	if err != nil {
		return "", err
	}
	args = append(args, bin)

	for _, arg := range finalCmd[1:] {
		if len(arg) == 1 {
			continue
		}
		if strings.Index(arg, " ") > 0 {
			args = append(args, fmt.Sprintf("'%s'", arg))
		} else {
			args = append(args, arg)
		}
	}

	argsString := strings.Join(args, " ")
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(argsString, " "), nil

}

func getBuilder(dir string, vcfg *vcfg.VCFG) (vpkg.Builder, error) {

	proj, err := vproj.LoadProject(dir)
	if err != nil {
		return nil, err
	}

	tgt, _ := proj.Target("")

	builder, err := tgt.NewBuilder()
	if err != nil {
		return nil, err
	}

	err = builder.MergeVCFG(vcfg)
	if err != nil {
		return nil, err
	}

	return builder, nil

}

func buildRaw(dir, target string, vcfg *vcfg.VCFG) (string, error) {

	log.Debugf("building raw from %s to %s", dir, target)

	builder, err := getBuilder(dir, vcfg)
	if err != nil {
		return "", err
	}
	defer builder.Close()

	reader, err := vpkg.ReaderFromBuilder(builder)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	err = vdisk.Build(context.Background(), f, &vdisk.BuildArgs{
		PackageReader: reader,
		Format:        vdisk.RAWFormat,
		Logger: &elog.CLI{
			DisableTTY: true,
		},
		WithVCFGDefaults: true,
	})

	return f.Name(), err

}
