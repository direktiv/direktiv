package isolates

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var kubernetesLock *distributed_locker.DistributedLocker

func initKubernetesLock() error {

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}

	kubernetesLock = distributed_locker.NewKubernetesLocker(
		dc, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, os.Getenv("DIREKTIV_LOCK_CM"), os.Getenv("DIREKTIV_KUBERNETES_NAMESPACE"),
	)

	log.Infof("kubernetes lock created")

	return nil

}

func kubeLock(key string) (lockgate.LockHandle, error) {

	log.Debugf("locking %s", key)

	acquired, lock, err := kubernetesLock.Acquire(key,
		lockgate.AcquireOptions{Shared: false,
			Timeout: 30 * time.Second})

	if err != nil {
		return lockgate.LockHandle{}, err
	}

	if !acquired {
		return lockgate.LockHandle{}, fmt.Errorf("lock %s not aquired", key)
	}

	return lock, nil

}

func kubeUnlock(lock lockgate.LockHandle) {

	log.Debugf("unlocking %s", lock.LockName)

	err := kubernetesLock.Release(lock)
	if err != nil {
		log.Errorf("can not unlock %v: %v", lock.LockName, err)
	}

}
