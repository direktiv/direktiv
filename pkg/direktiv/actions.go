package direktiv

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	gocni "github.com/containerd/go-cni"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	parser "github.com/novln/docker-parser"
	log "github.com/sirupsen/logrus"
	"github.com/vorteil/direktiv/pkg/dlog"
	"github.com/vorteil/direktiv/pkg/flow"
	"github.com/vorteil/direktiv/pkg/isolate"
	"github.com/vorteil/vorteil/pkg/elog"
	"github.com/vorteil/vorteil/pkg/imagetools"
	"github.com/vorteil/vorteil/pkg/vdecompiler"
	"github.com/vorteil/vorteil/pkg/vimg"
	"github.com/vorteil/vorteil/pkg/vkern"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type contextActionKey string

const (
	actionSubscription = "direktivaction"

	actionRespSubscription = "direktivactionresp"

	queueActionGroup = "daction"

	direktivBucket = "direktiv"

	kernelFolder = "/home/vorteil"

	actionCtxID contextActionKey = "actionCtxID"

	maxWaitSeconds = 1800
)

const (
	noError       = ""
	errorInternal = "au.com.direktiv.error.internal"
	errorImage    = "au.com.direktiv.error.image"
	errorNetwork  = "au.com.direktiv.error.network"
	errorIO       = "au.com.direktiv.error.io"
)

const (
	vmSmall = iota
	vmMedium
	vmLarge
)

type ctxs struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type actionManager struct {
	isolate.UnimplementedDirektivIsolateServer

	config         *Config
	minioClient    *minio.Client
	grpcIsolate    *grpc.Server
	grpcFlow       flow.DirektivFlowClient
	fileCache      *fileCache
	cni            gocni.CNI
	instanceLogger *dlog.Log
	dbManager      *dbManager

	actx map[string]*ctxs
	// actionCtxs []*ctxs
}

type actionWorkflow struct {
	InstanceID string
	Namespace  string
	State      string
	Step       int
	Timeout    int
}

type actionContainer struct {
	Image, Cmd string
	Size       int32
	Data       []byte
	Registries map[string]string
}

type actionRequest struct {
	ActionID string

	Workflow  actionWorkflow
	Container actionContainer
}

// ActionError is the struct returned from actions if there is an error
type ActionError struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// ContainerAuth implements authn.Authorize. Used for token authentication
type ContainerAuth struct {
	user, token string
}

// Authorization interface impl
func (f *ContainerAuth) Authorization() (*authn.AuthConfig, error) {

	ac := &authn.AuthConfig{
		Username: f.user,
		Password: f.token,
	}

	return ac, nil
}

func authorizationForRegistry(a string) *ContainerAuth {

	ac := &ContainerAuth{}

	ss := strings.SplitAfterN(a, "!", 2)
	if len(ss) != 2 {
		log.Errorf("authentication for registry invalid")
		return ac
	}

	// remove ! from username
	ac.user = ss[0][0 : len(ss[0])-1]
	ac.token = ss[1]

	return ac

}

func newActionManager(config *Config, dbManager *dbManager, l *dlog.Log) (*actionManager, error) {

	log.Debugf("action flow endpoint: %v", config.FlowAPI.Endpoint)

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(config.FlowAPI.Endpoint, opts...)
	if err != nil {
		return nil, err
	}
	grpcFlow := flow.NewDirektivFlowClient(conn)

	if config == nil || grpcFlow == nil {
		return nil, fmt.Errorf("config, grpc client and dbManager required for action manager")
	}

	am := &actionManager{
		config:         config,
		grpcFlow:       grpcFlow,
		instanceLogger: l,
		dbManager:      dbManager,
		actx:           make(map[string]*ctxs),
	}

	if len(config.Minio.User) == 0 || len(config.Minio.Password) == 0 {
		return nil, fmt.Errorf("minio username or password not set")
	}

	if len(config.Minio.Endpoint) == 0 {
		return nil, fmt.Errorf("minio endpoint not set")
	}

	vorteild := kernelFolder
	kernels := filepath.Join(vorteild, "kernels")
	watch := filepath.Join(kernels, "watch")
	sources := []string{"https://downloads.vorteil.io/kernels"}

	ksrc, err := vkern.CLI(vkern.CLIArgs{
		Directory:          kernels,
		DropPath:           watch,
		RemoteRepositories: sources,
	}, &elog.CLI{
		DisableTTY: true,
	})

	if err != nil {
		return nil, err
	}

	vkern.Global = ksrc
	vimg.GetKernel = ksrc.Get
	vimg.GetLatestKernel = vkern.ConstructGetLastestKernelsFunc(&ksrc)

	am.fileCache, err = newFileCache(am)
	if err != nil {
		return nil, err
	}

	// check CNI networking
	am.cni, err = am.prepareNetwork()

	return am, err

}

func (am *actionManager) grpcIsolateStart() error {

	bind := am.config.IsolateAPI.Bind
	log.Debugf("action endpoint starting at %s", bind)

	// TODO: save listener somewhere so that it can be shutdown
	// TODO: save grpc somewhere so that it can be shutdown
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	am.grpcIsolate = grpc.NewServer()

	isolate.RegisterDirektivIsolateServer(am.grpcIsolate, am)

	go am.grpcIsolate.Serve(listener)

	return nil

}

func (am *actionManager) start() error {
	log.Infof("starting action runner")

	insecure := true
	if am.config.Minio.Secure > 0 {
		insecure = false
	}

	ssl := false
	if am.config.Minio.SSL > 0 {
		ssl = true
	}

	minioClient, err := minio.New(am.config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(am.config.Minio.User, am.config.Minio.Password, ""),
		Secure: ssl,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: insecure},
		},
	})

	if err != nil {
		log.Errorf("can not create minio client: %v", err)
		return err
	}

	found, err := minioClient.BucketExists(context.Background(), direktivBucket)
	if !found && err == nil {
		// create default bucket
		err = minioClient.MakeBucket(context.Background(), direktivBucket,
			minio.MakeBucketOptions{Region: am.config.Minio.Region})
		if err != nil {
			log.Errorf("can not create bucket for direktiv: %v", err)
			return err
		}
	}

	if err != nil {
		log.Errorf("can not connect to minio %v: %v", am.config.Minio.Endpoint, err)
		return err
	}

	am.minioClient = minioClient

	return nil
}

func (am *actionManager) stop() {

	log.Infof("stopping action runner")
	am.grpcIsolate.GracefulStop()

}

func hashImg(img, cmd string) string {
	h := sha256.New()

	h.Write([]byte(fmt.Sprintf("%v", img)))
	h.Write([]byte(fmt.Sprintf("%v", cmd)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (am *actionManager) addCtx(timeout *int64, actionID string) *ctxs {

	var to int64
	if timeout != nil {
		to = *timeout
	}

	// create context for firecracker, it is max 15 minutes / 1800 seconds for a VM
	c := context.Background()
	ctx := context.WithValue(c, actionCtxID, actionID)
	if to == 0 || to > maxWaitSeconds {
		to = maxWaitSeconds
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(to)*time.Second)
	ctxs := &ctxs{
		cancel: cancel,
		ctx:    ctx,
	}
	am.actx[actionID] = ctxs

	return ctxs

}

func (am *actionManager) finishCancelIsolate(actionID string) {

	if ctx, ok := am.actx[actionID]; ok {
		ctx.cancel()
		delete(am.actx, actionID)
	}

}

func findAuthForRegistry(img string, registries map[string]string) []remote.Option {
	// authenticate if there are registries for this images
	opts := []remote.Option{}
	r, _ := parser.Parse(img)

	if val, ok := registries[r.Registry()]; ok {
		log.Debugf("found auth for %s", r.Registry())
		auth := authorizationForRegistry(val)
		opts = append(opts, remote.WithAuth(auth))
	}

	return opts

}

func (am *actionManager) runAction(in *isolate.RunIsolateRequest) {

	var (
		ns, instID, actionID string
		img, cmd             string

		data, din []byte
	)

	ns = in.GetNamespace()
	actionID = in.GetActionId()

	img = in.GetImage()
	cmd = in.GetCommand()
	instID = in.GetInstanceId()

	log15log, err := (*am.instanceLogger).LoggerFunc(ns, instID)
	if err != nil {
		log.Errorf("can not create logger for isolate: %v", err)
		return
	}
	defer log15log.Close()

	serr := func(err error, errCode string) *ActionError {
		ae := ActionError{
			ErrorMessage: err.Error(),
			ErrorCode:    errCode,
		}
		return &ae
	}

	disk, err := am.fileCache.getImage(img, cmd, in.Registries)
	if err != nil {
		am.respondToAction(serr(err, errorImage), data, in)
		return
	}

	// prepare cni networking
	nws, err := am.setupNetworkForVM(actionID)
	if err != nil {
		am.respondToAction(serr(err, errorNetwork), data, in)
		return
	}

	defer am.deleteNetworkForVM(actionID)

	// build data disk to attach
	dataDisk, err := am.buildDataDisk(actionID, in.Data, nws)
	if err != nil {
		am.respondToAction(serr(err, errorIO), data, in)
		return
	}
	defer os.Remove(dataDisk)

	ctxs := am.addCtx(in.Timeout, actionID)

	defer am.finishCancelIsolate(actionID)

	err = am.runFirecracker(ctxs.ctx, actionID, disk, dataDisk, in.GetSize())
	if err != nil {
		am.respondToAction(serr(err, errorInternal), data, in)
		return
	}

	log.Debugf("firecracker finished")

	// successful, so get the logs & results
	dimg, err := vdecompiler.Open(dataDisk)
	if err != nil {
		am.respondToAction(serr(err, errorIO), data, in)
		return
	}

	readFileFromDisk := func(disk *vdecompiler.IO, file string, d *[]byte) error {

		var buf bytes.Buffer

		// we don't check the error here. files don't have to exist
		rdr, err := imagetools.CatImageFile(dimg, file, false)

		if err != nil {
			return nil
		}

		_, err = io.Copy(&buf, rdr)
		if err != nil {
			return err
		}

		*d = buf.Bytes()
		return nil

	}

	ll := func(file string) {
		var in []byte
		err = readFileFromDisk(dimg, file, &in)

		if err == nil && len(in) > 0 {
			log15log.Info(string(in))
		}
		if err != nil {
			log.Errorf("error reading %s: %v", file, err)
		}
	}

	ll("/log.log")
	ll("/error.log")

	err = readFileFromDisk(dimg, "/error.json", &din)
	if err != nil {
		am.respondToAction(serr(err, errorIO), data, in)
		return
	}

	if len(din) > 0 {

		var ae ActionError
		err := json.Unmarshal(din, &ae)
		if err != nil {
			log15log.Error(fmt.Sprintf("error parsing error file: %v", err))
			am.respondToAction(serr(fmt.Errorf("%w; %s", err, string(din)), errorIO), data, in)
			return
		}

		log15log.Error(ae.ErrorMessage)
		am.respondToAction(&ae, data, in)
		return

	}

	// can not do much if that fails, print to logs, otherwise we return the data
	err = readFileFromDisk(dimg, "/data.out", &data)
	if err != nil {
		log15log.Error(fmt.Sprintf("error parsing error file: %v", err))
		am.respondToAction(serr(err, errorIO), data, in)
		return
	}

	go func() {

		log.Debugf("responding to action caller")
		am.respondToAction(nil, data, in)

	}()

}

// RunIsolate rus container images in firecracker VMs
func (am *actionManager) RunIsolate(ctx context.Context, in *isolate.RunIsolateRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	if len(in.GetNamespace()) == 0 || len(in.GetImage()) == 0 {
		log.Errorf("namespace or image not provided")
		return &resp, fmt.Errorf("no namespace or image")
	}

	if len(in.GetActionId()) == 0 {
		log.Errorf("actionID not provided")
		return &resp, fmt.Errorf("actionID empty")
	}

	log.Debugf("running isolate %s", in.GetNamespace())

	go am.runAction(in)

	return &resp, nil

}

// handleAction is to receive the action to execute. is req/resp
func (am *actionManager) respondToAction(ae *ActionError, data []byte, in *isolate.RunIsolateRequest) {

	log.Debugf("action responding")

	r := &flow.ReportActionResultsRequest{
		InstanceId: in.InstanceId,
		Step:       in.Step,
		ActionId:   in.ActionId,
		Output:     data,
	}

	if ae != nil {
		log.Debugf("error reporting: %v", ae)
		r.ErrorCode = &ae.ErrorCode
		r.ErrorMessage = &ae.ErrorMessage
	}

	_, err := am.grpcFlow.ReportActionResults(context.Background(), r)

	if err != nil {
		log.Errorf("error reporting action results: %v", err)
	}

}

func (am *actionManager) retrieveImageS3(img, cmd, path string) error {

	h := hashImg(img, cmd)

	encryption := encrypt.DefaultPBKDF([]byte(am.config.Minio.Encrypt), []byte(direktivBucket+h))

	return am.minioClient.FGetObject(context.Background(), direktivBucket, h, path, minio.GetObjectOptions{
		ServerSideEncryption: encryption,
	})

}

func (am *actionManager) storeImageS3(img, cmd, disk string) error {

	h := hashImg(img, cmd)

	encryption := encrypt.DefaultPBKDF([]byte(am.config.Minio.Encrypt), []byte(direktivBucket+h))
	_, err := am.minioClient.FPutObject(context.Background(), direktivBucket, h, disk, minio.PutObjectOptions{
		ServerSideEncryption: encryption,
	})

	t := time.Now().Add((7 * 24) * time.Hour)
	am.minioClient.PutObjectRetention(context.Background(), direktivBucket, h, minio.PutObjectRetentionOptions{
		RetainUntilDate: &t,
	})

	return err

}

func (am *actionManager) removeImageS3(img, cmd string) error {

	h := hashImg(img, cmd)

	return am.minioClient.RemoveObject(context.Background(), direktivBucket, h, minio.RemoveObjectOptions{})

}
