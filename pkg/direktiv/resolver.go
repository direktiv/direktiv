package direktiv

import (
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/resolver"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// KubeResolver ...
type KubeResolver struct {
}

// KubeResolverBuilder ...
type KubeResolverBuilder struct {
}

// Build ...
func (b *KubeResolverBuilder) Build(target resolver.Target,
	cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	r := NewResolver()
	host, port, err := net.SplitHostPort(target.Endpoint)
	if err != nil {
		log.Errorf("can not connect to %v", target.Endpoint)
	}

	go watchEndpoints(host, port, cc)

	return r, nil
}

// Scheme returns the default scheme for this reoslver
func (b *KubeResolverBuilder) Scheme() string {
	return "direktiv-kube"
}

// NewResolver returns a resolve for kubernetes
func NewResolver() *KubeResolver {
	return &KubeResolver{}
}

// Close closes the resoolver
func (r *KubeResolver) Close() {

}

// ResolveNow does nothing in our case
func (r *KubeResolver) ResolveNow(o resolver.ResolveNowOptions) {
}

func watchEndpoints(svc, port string, conn resolver.ClientConn) error {

	clientset, kns, err := getClientSet()
	if err != nil {
		log.Errorf("can not get client set for kuberneets resolver: %v", err)
		return err
	}

	// sometimes the namespace is part of the request, e.g. myservice.default
	ns := strings.SplitN(svc, ".", 2)
	if len(ns) == 2 {
		svc = ns[0]
		kns = ns[1]
	}

	// get initial backends
	var opt metav1.GetOptions
	ep, err := clientset.CoreV1().Endpoints(kns).Get(svc, opt)
	if err != nil {
		log.Errorf("can not get client set for kubernetes resolver: %v", err)
		return err
	}

	addrs := getAddrFromEnpoint(ep, port)
	log.Infof("initial addresses: %v", addrs)
	conn.UpdateState(resolver.State{Addresses: addrs})

	resyncPeriod := 30 * time.Minute
	si := kubeinformers.NewSharedInformerFactoryWithOptions(clientset, resyncPeriod,
		kubeinformers.WithNamespace(kns))

	si.Core().V1().Endpoints().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				s := obj.(*v1.Endpoints)
				if s.Name == svc {
					addrs := getAddrFromEnpoint(s, port)
					log.Infof("new addresses: %v", addrs)
					conn.UpdateState(resolver.State{Addresses: addrs})
				}
			},
			DeleteFunc: func(obj interface{}) {
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				s := newObj.(*v1.Endpoints)
				if s.Name == svc {
					addrs := getAddrFromEnpoint(s, port)
					log.Infof("updated addresses: %v", addrs)
					conn.UpdateState(resolver.State{Addresses: addrs})
				}
			},
		},
	)

	stop := make(chan struct{})
	si.WaitForCacheSync(stop)
	si.Start(stop)

	return nil

}

func getAddrFromEnpoint(ep *v1.Endpoints, port string) []resolver.Address {

	var ips []resolver.Address

	for _, ss := range ep.Subsets {
		for _, a := range ss.Addresses {
			addr := resolver.Address{
				Addr: fmt.Sprintf("%s:%s", a.IP, port),
			}
			ips = append(ips, addr)
		}
	}

	return ips
}
