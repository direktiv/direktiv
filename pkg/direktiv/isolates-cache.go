package direktiv

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	hash "github.com/mitchellh/hashstructure/v2"
	parser "github.com/novln/docker-parser"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	baseDir = "/tmp"
)

var (
	cacheDir = fmt.Sprintf("%s/diskcache", baseDir)
)

type cacheItem struct {
	lastAccessed time.Time
	lastChanged  time.Time
	size         int64
}

type fileCache struct {
	items         map[string]*cacheItem
	spaceLeft     int64
	mtx           sync.Mutex
	isolateServer *isolateServer
}

func newFileCache(is *isolateServer) (*fileCache, error) {

	var stat unix.Statfs_t
	err := unix.Statfs(baseDir, &stat)
	if err != nil {
		log.Errorf("can not get available disk size")
		return nil, err
	}

	// create cache dir if it does not exist
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.Mkdir(cacheDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	// delete all existing cache files
	err = os.RemoveAll(cacheDir)
	if err != nil {
		return nil, err
	}

	// 70% percent of disk space is cache for images
	fc := &fileCache{
		spaceLeft:     int64(float64(stat.Bavail*uint64(stat.Bsize)) * 0.7),
		items:         make(map[string]*cacheItem),
		isolateServer: is,
	}

	log.Infof("cache size: %s", bytefmt.ByteSize(uint64(fc.spaceLeft)))

	return fc, nil
}

func getLastChanged(image string, registries map[string]string) (time.Time, error) {

	t := time.Now()

	ref, err := name.ParseReference(image)
	if err != nil {
		return time.Time{}, err
	}

	opts := findAuthForRegistry(image, registries)

	// img, err := remote.Image(ref, remote.WithAuth(&FluxAuth{}))
	img, err := remote.Image(ref, opts...)
	if err != nil {
		return time.Time{}, err
	}

	config, err := getContainerConfig(img)
	if err != nil {
		return t, err
	}

	if c, ok := config["created"]; ok {
		t, err = time.Parse(time.RFC3339Nano, c.(string))
		if err != nil {
			t = time.Now()
			return t, err
		}
	}

	return t, err

}

func needsUpdate(item *cacheItem, image string, registries map[string]string) (bool, error) {

	ref, err := parser.Parse(image)

	// only check if latest
	if err == nil && ref.Tag() == "latest" {
		lc, err := getLastChanged(image, registries)
		if err != nil {
			return false, err
		}
		log.Debugf("compare last changed %v = %v", lc, item.lastChanged)
		if !lc.Equal(item.lastChanged) {
			return true, nil
		}
	}

	return false, nil
}

func (fc *fileCache) getImage(img, cmd string, registries map[string]string) (string, error) {

	h := hashImg(img, cmd)
	disk := filepath.Join(cacheDir, h)

	// create lock for building, 180 seconds for building it
	lockHash, _ := hash.Hash(fmt.Sprintf("%s-%s", img, cmd), hash.FormatV2, nil)
	log.Debugf("building disk for hash: %v", lockHash)
	conn, err := fc.isolateServer.dbManager.lockDB(lockHash, 180)
	if err != nil {
		return "", err
	}
	defer fc.isolateServer.dbManager.unlockDB(lockHash, conn)

	log.Debugf("getting img %s (%s)", img, h[:8])

	// get local first
	if i, ok := fc.items[h]; ok {
		log.Debugf("item %s in cache", h[:8])
		upd, err := needsUpdate(i, img, registries)
		if err != nil {
			return "", err
		}
		if !upd {
			return disk, nil
		} else {
			delete(fc.items, h)
			os.Remove(disk)
			fc.isolateServer.removeImageS3(img, cmd)
		}
	}

	err = fc.isolateServer.retrieveImageS3(img, cmd, disk)
	if err != nil {
		// not local and not remote, we need to build the disk
		log.Debugf("disk not found on s3: %v", err)

		disk, err = buildImageDisk(img, cmd,
			fc.isolateServer.config.Kernel.Runtime, cacheDir, registries)
		if err != nil {
			log.Errorf("image build error: %v", err)
			return "", err
		}

		// we can ignore errors here, worst case we build every time the disk is
		// requested and not in cache
		err = fc.isolateServer.storeImageS3(img, cmd, disk)
		if err != nil {
			log.Errorf("image build error: %v", err)
		}

	}

	// add to cache and return path
	fi, err := os.Stat(disk)
	if err != nil {
		return "", err
	}

	lc, err := getLastChanged(img, registries)
	if err != nil {
		return "", err
	}

	err = fc.addItem(h, fi.Size(), lc)
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, h), nil
}

func (fc *fileCache) removeItem(key string) {

	if i, ok := fc.items[key]; ok {
		fc.spaceLeft += i.size
		delete(fc.items, key)
		os.Remove(filepath.Join(cacheDir, key))
	}

}

func (fc *fileCache) addItem(key string, sz int64, t time.Time) error {

	fc.mtx.Lock()
	defer fc.mtx.Unlock()

	err := fc.checkCacheSize(sz)
	if err != nil {
		return err
	}

	fc.spaceLeft -= sz
	fc.items[key] = &cacheItem{
		lastAccessed: time.Now(),
		lastChanged:  t,
		size:         sz,
	}

	return nil

}

func (fc *fileCache) checkCacheSize(sz int64) error {

	if fc.spaceLeft > sz {
		fc.spaceLeft -= sz
		return nil
	}

	counter := len(fc.items)

	// delete the oldest
	for fc.spaceLeft < sz && counter >= 0 {
		o := time.Now()
		n := ""

		for name, itm := range fc.items {
			if n == "" {
				o = itm.lastAccessed
				n = name
			} else if itm.lastAccessed.Before(o) {
				o = itm.lastAccessed
				n = name
			}
		}

		// remove oldest
		if len(n) > 0 {
			fc.removeItem(n)
		}

		counter--
	}

	if fc.spaceLeft < sz {
		return fmt.Errorf("not enough space in cache")
	}

	return nil

}
