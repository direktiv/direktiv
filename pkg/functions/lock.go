package functions

import (
	"fmt"
	"os"
	"time"

	"github.com/vorteil/direktiv/pkg/util"
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

	logger.Debugf("lock for cm %s in namespace %s",
		os.Getenv("DIREKTIV_LOCK_CM"), os.Getenv(util.DirektivNamespace))

	kubernetesLock = distributed_locker.NewKubernetesLocker(
		dc, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		}, os.Getenv("DIREKTIV_LOCK_CM"), os.Getenv(util.DirektivNamespace),
	)

	logger.Infof("kubernetes lock created")

	return nil

}

func kubeLock(key string, blocking bool) (lockgate.LockHandle, error) {

	logger.Debugf("locking %s", key)

	acquired, lock, err := kubernetesLock.Acquire(key,
		lockgate.AcquireOptions{Shared: false, NonBlocking: blocking,
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

	logger.Debugf("unlocking %s", lock.LockName)

	err := kubernetesLock.Release(lock)
	if err != nil {
		logger.Errorf("can not unlock %v: %v", lock.LockName, err)
	}

}
