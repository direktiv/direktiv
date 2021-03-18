package direktiv

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

type contextIsolateKey string

const (
	direktivBucket = "direktiv"

	kernelFolder = "/home/vorteil"

	isolateCtxID contextIsolateKey = "isolateCtxID"

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

type isolateServer struct {
	isolate.UnimplementedDirektivIsolateServer

	config         *Config
	minioClient    *minio.Client
	grpc           *grpc.Server
	fileCache      *fileCache
	cni            gocni.CNI
	instanceLogger *dlog.Log
	dbManager      *dbManager

	actx map[string]*ctxs

	flowClient flow.DirektivFlowClient
	grpcConn   *grpc.ClientConn

	mtx sync.Mutex
}

type isolateWorkflow struct {
	InstanceID string
	Namespace  string
	State      string
	Step       int
	Timeout    int
}

type isolateContainer struct {
	Image, Cmd string
	Size       int32
	Data       []byte
	Registries map[string]string
}

type isolateRequest struct {
	ActionID string

	Workflow  isolateWorkflow
	Container isolateContainer
}

// IsolateError is the struct returned from isolates if there is an error
type IsolateError struct {
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

func newIsolateManager(config *Config, dbManager *dbManager, l *dlog.Log) (*isolateServer, error) {

	log.Debugf("isolate flow endpoint: %v", config.FlowAPI.Endpoint)

	is := &isolateServer{
		config:         config,
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

	is.fileCache, err = newFileCache(is)
	if err != nil {
		return nil, err
	}

	// check CNI networking
	is.cni, err = is.prepareNetwork()

	// check the timeouts for firecracker sdk. they are very low for high load systems
	if len(os.Getenv("FIRECRACKER_GO_SDK_REQUEST_TIMEOUT_MILLISECONDS")) == 0 {
		log.Debugf("setting firecracker request timeout to 5000ms")
		os.Setenv("FIRECRACKER_GO_SDK_REQUEST_TIMEOUT_MILLISECONDS", "5000")
	}

	if len(os.Getenv("FIRECRACKER_GO_SDK_INIT_TIMEOUT_SECONDS")) == 0 {
		log.Debugf("setting firecracker sdk init to 5s")
		os.Setenv("FIRECRACKER_GO_SDK_INIT_TIMEOUT_SECONDS", "5")
	}

	return is, err

}

func (is *isolateServer) grpcStart(s *WorkflowServer) error {
	return s.grpcStart(&is.grpc, "isolate", s.config.IsolateAPI.Bind, func(srv *grpc.Server) {
		isolate.RegisterDirektivIsolateServer(srv, is)
	})
}

func (is *isolateServer) stop() {

	if is.grpc != nil {
		is.grpc.GracefulStop()
	}

	if is.grpcConn != nil {
		is.grpcConn.Close()
	}

}

func (is *isolateServer) name() string {
	return "isolate"
}

func (is *isolateServer) start(s *WorkflowServer) error {
	log.Infof("starting isolate runner")

	err := is.grpcStart(s)
	if err != nil {
		return err
	}

	var transport *http.Transport
	if is.config.Minio.Secure == 0 {
		log.Debugf("minio client insecureSkipVerify")
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	useSSL := true
	if is.config.Minio.SSL == 0 {
		log.Debugf("minio client using SSL")
		useSSL = false
	}

	minioClient, err := minio.New(is.config.Minio.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(is.config.Minio.User, is.config.Minio.Password, ""),
		Secure:    useSSL,
		Transport: transport,
	})

	if err != nil {
		log.Errorf("can not create minio client: %v", err)
		return err
	}

	found, err := minioClient.BucketExists(context.Background(), direktivBucket)
	if !found && err == nil {
		// create default bucket
		err = minioClient.MakeBucket(context.Background(), direktivBucket,
			minio.MakeBucketOptions{Region: is.config.Minio.Region})
		if err != nil {
			log.Errorf("can not create bucket for direktiv: %v", err)
			return err
		}
	}

	if err != nil {
		log.Errorf("can not connect to minio %v: %v", is.config.Minio.Endpoint, err)
		return err
	}

	is.minioClient = minioClient

	// get flow client
	conn, err := getEndpointTLS(is.config, flowComponent, is.config.FlowAPI.Endpoint)
	if err != nil {
		return err
	}

	is.grpcConn = conn
	is.flowClient = flow.NewDirektivFlowClient(conn)

	log.Infof("isolate runner started")

	return nil
}

func hashImg(img, cmd string) string {
	h := sha256.New()

	h.Write([]byte(fmt.Sprintf("%v", img)))
	h.Write([]byte(fmt.Sprintf("%v", cmd)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (is *isolateServer) addCtx(timeout *int64, isolateID string) *ctxs {

	var to int64
	if timeout != nil {
		to = *timeout
	}

	// create context for firecracker, it is max 15 minutes / 1800 seconds for a VM
	c := context.Background()
	ctx := context.WithValue(c, isolateCtxID, isolateID)
	if to == 0 || to > maxWaitSeconds {
		to = maxWaitSeconds
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(to)*time.Second)
	ctxs := &ctxs{
		cancel: cancel,
		ctx:    ctx,
	}

	is.mtx.Lock()
	defer is.mtx.Unlock()
	is.actx[isolateID] = ctxs

	return ctxs

}

func (is *isolateServer) finishCancelIsolate(isolateID string) {

	is.mtx.Lock()
	defer is.mtx.Unlock()
	if ctx, ok := is.actx[isolateID]; ok {
		ctx.cancel()
		delete(is.actx, isolateID)
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

func (is *isolateServer) runAsFirecracker(img, cmd, isolateID string, in *isolate.RunIsolateRequest, log15log dlog.Logger) {

	var data, din []byte

	serr := func(err error, errCode string) *IsolateError {
		ae := IsolateError{
			ErrorMessage: err.Error(),
			ErrorCode:    errCode,
		}
		return &ae
	}

	disk, err := is.fileCache.getImage(img, cmd, in.Registries)
	if err != nil {
		is.respondToAction(serr(err, errorImage), data, in)
		return
	}

	// prepare cni networking
	nws, err := is.setupNetworkForVM(isolateID)
	if err != nil {
		is.respondToAction(serr(err, errorNetwork), data, in)
		return
	}

	defer is.deleteNetworkForVM(isolateID)

	// build data disk to attach
	dataDisk, err := is.buildDataDisk(isolateID, in.Data, nws)
	if err != nil {
		is.respondToAction(serr(err, errorIO), data, in)
		return
	}
	defer os.Remove(dataDisk)

	ctxs := is.addCtx(in.Timeout, isolateID)

	defer is.finishCancelIsolate(isolateID)

	err = is.runFirecracker(ctxs.ctx, isolateID, disk, dataDisk, in.GetSize())
	if err != nil {
		is.respondToAction(serr(err, errorInternal), data, in)
		return
	}

	log.Debugf("firecracker finished")

	// successful, so get the logs & results
	dimg, err := vdecompiler.Open(dataDisk)
	if err != nil {
		is.respondToAction(serr(err, errorIO), data, in)
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
		is.respondToAction(serr(err, errorIO), data, in)
		return
	}

	if len(din) > 0 {

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
	err = readFileFromDisk(dimg, "/data.out", &data)
	if err != nil {
		log15log.Error(fmt.Sprintf("error parsing error file: %v", err))
		is.respondToAction(serr(err, errorIO), data, in)
		return
	}

	go func() {

		maxlen := math.Min(256, float64(len(data)))
		log.Debugf("responding to isolate caller: %v", string(data[0:int(maxlen)]))
		is.respondToAction(nil, data, in)

	}()

}

func (is *isolateServer) runAction(in *isolate.RunIsolateRequest) {

	var (
		ns, instID, isolateID string
		img, cmd              string
	)

	ns = in.GetNamespace()
	isolateID = in.GetActionId()

	log.Debugf("isolate action id: %v", isolateID)

	img = in.GetImage()
	cmd = in.GetCommand()
	instID = in.GetInstanceId()

	log15log, err := (*is.instanceLogger).LoggerFunc(ns, instID)
	if err != nil {
		log.Errorf("can not create logger for isolate: %v", err)
		return
	}
	defer log15log.Close()

	log.Debugf("isolation level: %v", is.config.IsolateAPI.Isolation)
	if is.config.IsolateAPI.Isolation == "container" {
		is.runAsContainer(img, cmd, isolateID, in, log15log)
	} else {
		is.runAsFirecracker(img, cmd, isolateID, in, log15log)
	}

}

// RunIsolate rus container images in firecracker VMs
func (is *isolateServer) RunIsolate(ctx context.Context, in *isolate.RunIsolateRequest) (*emptypb.Empty, error) {

	var resp emptypb.Empty

	if len(in.GetNamespace()) == 0 || len(in.GetImage()) == 0 {
		log.Errorf("namespace or image not provided")
		return &resp, fmt.Errorf("no namespace or image")
	}

	if len(in.GetActionId()) == 0 {
		log.Errorf("isolateID not provided")
		return &resp, fmt.Errorf("isolateID empty")
	}

	go is.runAction(in)

	return &resp, nil

}

func (is *isolateServer) respondToAction(ae *IsolateError, data []byte, in *isolate.RunIsolateRequest) {

	log.Debugf("isolate responding")

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

	_, err := is.flowClient.ReportActionResults(context.Background(), r)

	if err != nil {
		log.Errorf("error reporting isolate results: %v", err)
	}

}

func (is *isolateServer) retrieveImageS3(img, cmd, path string) error {

	h := hashImg(img, cmd)

	// only encrypt if SSL
	receiveOptions := minio.GetObjectOptions{}
	if is.config.Minio.SSL > 0 {
		encryption := encrypt.DefaultPBKDF([]byte(is.config.Minio.Encrypt), []byte(direktivBucket+h))
		receiveOptions.ServerSideEncryption = encryption
	}

	return is.minioClient.FGetObject(context.Background(), direktivBucket, h, path, receiveOptions)

}

func (is *isolateServer) storeImageS3(img, cmd, disk string) error {

	h := hashImg(img, cmd)

	// only encrypt if SSL
	storeOptions := minio.PutObjectOptions{}
	if is.config.Minio.SSL > 0 {
		encryption := encrypt.DefaultPBKDF([]byte(is.config.Minio.Encrypt), []byte(direktivBucket+h))
		storeOptions.ServerSideEncryption = encryption
	}

	_, err := is.minioClient.FPutObject(context.Background(), direktivBucket, h, disk, storeOptions)

	t := time.Now().Add((7 * 24) * time.Hour)
	is.minioClient.PutObjectRetention(context.Background(), direktivBucket, h, minio.PutObjectRetentionOptions{
		RetainUntilDate: &t,
	})

	return err

}

func (is *isolateServer) removeImageS3(img, cmd string) error {

	h := hashImg(img, cmd)

	return is.minioClient.RemoveObject(context.Background(), direktivBucket, h, minio.RemoveObjectOptions{})

}
