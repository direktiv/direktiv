package direktiv

import (
	"context"
	"fmt"
	"runtime"

	gocni "github.com/containerd/go-cni"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

const (
	cniConf = `
  	{
    "name": "direktiv-net",
    "cniVersion": "0.4.0",
    "plugins": [
      {
        "type": "ptp",
  			"ipMasq": true,
  			"ipam": {
          "type": "host-local",
          "subnet": "10.77.0.0/16"
        }
      },
      {
        "type": "firewall"
      },
      {
        "type": "tc-redirect-tap"
      }
    ]
  }`
)

var (
	defaultMask = "255.255.255.255"
)

type networkSetting struct {
	IP      string `json:"ip"`
	Mask    string `json:"mask"`
	Gateway string `json:"gw"`
}

func (is *isolateServer) prepareNetwork() (gocni.CNI, error) {

	l, err := gocni.New(
		gocni.WithPluginDir([]string{"/opt/cni/bin"}),
		gocni.WithInterfacePrefix("eth"))

	if err != nil {
		log.Errorf("failed to initialize cni library: %v", err)
		return nil, err
	}

	// Load the cni configuration
	if err := l.Load(gocni.WithLoNetwork, gocni.WithConfListBytes([]byte(cniConf))); err != nil {
		log.Errorf("failed to load cni configuration: %v", err)
		return nil, err
	}

	// err = ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0755)
	// if err != nil {
	// 	log.Warnf("could not set ip forwarding")
	// }

	return l, nil

}

// DeleteNetwork cleans up netns networks
func (is *isolateServer) deleteNetworkForVM(name string) error {

	path := fmt.Sprintf("/var/run/netns/%s", name)

	log.Debugf("deleting network: %s", name)

	if err := is.cni.Remove(context.Background(), name, path); err != nil {
		log.Errorf("failed to delete network: %v", err)
		return err
	}

	return deleteNetns(name)

}

func (is *isolateServer) setupNetworkForVM(name string) (networkSetting, error) {

	var nws networkSetting

	err := createNetns(name)
	if err != nil {
		log.Errorf("error creating namespace: %v", err)
		return nws, err
	}

	// default netns path
	path := fmt.Sprintf("/var/run/netns/%s", name)
	result, err := is.cni.Setup(context.Background(), name, path, gocni.WithArgs("TC_REDIRECT_TAP_GID", "1000"),
		gocni.WithArgs("TC_REDIRECT_TAP_UID", "1000"), gocni.WithArgs("TC_REDIRECT_TAP_NAME", "tap0"), gocni.WithArgs("IgnoreUnknown", "true"))
	if err != nil {
		log.Errorf("error creating cni: %v", err)
		return nws, err
	}

	log.Debugf("ip for network (%s): %v", name, result.Interfaces["eth1"].IPConfigs[0].IP)

	nws.IP = result.Interfaces["eth1"].IPConfigs[0].IP.String()
	nws.Gateway = result.Interfaces["eth1"].IPConfigs[0].Gateway.String()
	nws.Mask = defaultMask

	return nws, nil
}

func createNetns(name string) error {

	log.Debugf("create netns %s", name)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	origns, err := netns.Get()
	if err != nil {
		return err
	}
	defer origns.Close()

	newns, err := netns.NewNamed(name)
	if err != nil {
		log.Errorf("can not create netns %s", name)
		return err
	}

	defer newns.Close()

	netns.Set(origns)
	return nil

}

func deleteNetns(name string) error {
	log.Debugf("delete netns %s", name)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	return netns.DeleteNamed(name)

}
