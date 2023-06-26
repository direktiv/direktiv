package cluster

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasicSetGetCache(t *testing.T) {

	config := DefaultCacheConfig()
	config.Prefix = "basic"

	c, err := NewCache(config)
	require.NoError(t, err)

	err = c.Set("hello1", []byte("world"), 0)
	require.NoError(t, err)
	err = c.Set("hello2", []byte("world"), time.Second)
	require.NoError(t, err)

	time.Sleep(time.Second)

	val1, err := c.Get("hello1")
	require.NoError(t, err)
	require.NotEmpty(t, val1)
	require.Equal(t, val1, []byte("world"))

	_, err = c.Get("hello2")
	require.Error(t, err)

	c.Invalidate("hello1")
	_, err = c.Get("hello1")
	require.Error(t, err)

}
func TestInvalidateAllCache(t *testing.T) {

	config := DefaultCacheConfig()
	config.Prefix = "invalidate"

	c, err := NewCache(config)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		err = c.Set(fmt.Sprintf("hello%d", i), []byte("world"), 0)
		require.NoError(t, err)
	}

	c.InvalidateAll()

	for i := 0; i < 5; i++ {
		_, err := c.Get(fmt.Sprintf("hello%d", i))
		require.Error(t, err, fmt.Sprintf("fetching %s", fmt.Sprintf("hello%d", i)))
	}

}

func TestFunctionCache(t *testing.T) {

	config := DefaultCacheConfig()
	config.Prefix = "function"

	c, err := NewCache(config)
	require.NoError(t, err)

	value, err := c.GetFunction("hello1", fetcherOk, 0)
	require.NoError(t, err)
	require.Equal(t, value, []byte("world"))

	// normal get should get it now
	value, err = c.Get("hello1")
	require.NoError(t, err)
	require.Equal(t, value, []byte("world"))

	_, err = c.GetFunction("hello2", fetcherFail, 0)
	require.Error(t, err)
	_, err = c.Get("hello2")
	require.Error(t, err)

	value, err = c.GetFunction("hello3", fetcherOk, time.Second)
	require.NoError(t, err)
	require.Equal(t, value, []byte("world"))
	time.Sleep(time.Second)

	// expired
	_, err = c.Get("hello3")
	require.Error(t, err)

}

func fetcherOk(key string) ([]byte, error) {
	return []byte("world"), nil
}

func fetcherFail(key string) ([]byte, error) {
	return []byte("world"), fmt.Errorf("failed")
}

func TestClusterCache(t *testing.T) {

	// start cluster
	count := 2
	nodes, err := createCluster(t, count, []string{"cache"}, true)
	require.NoError(t, err)

	for i := 0; i < count; i++ {
		defer nodes[i].Stop()
	}

	// check three node cluster
	require.Eventually(t, func() bool {
		return rightNumber(nodes)
	}, 10*time.Second, time.Second, "node count failed")

	config1 := DefaultCacheConfig()
	config1.Node = nodes[0]
	config1.Prefix = "cache1"

	c1, err := NewCache(config1)
	require.NoError(t, err)

	config2 := DefaultCacheConfig()
	config2.Node = nodes[1]
	config2.Prefix = "cache2"

	c2, err := NewCache(config2)
	require.NoError(t, err)

	fmt.Printf("jklhsjahd %v %v", c1, c2)

	c1.Set("hello", []byte("world"), 0)
	c2.Set("hello", []byte("world"), 0)

	val1, err := c1.Get("hello")
	require.NoError(t, err)
	require.Equal(t, val1, []byte("world"))
	val2, err := c2.Get("hello")
	require.NoError(t, err)
	require.Equal(t, val2, []byte("world"))

	// invalidate on one and wait for other server to invalidate it
	err = c1.Invalidate("hello")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		_, err = c1.Get("hello")
		if err == nil {
			return false
		}
		_, err = c1.Get("hello")
		if err == nil {
			return false
		}
		return true
	}, 10*time.Second, time.Second, "node count failed")

}
