package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasicSetGetCache(t *testing.T) {
	c, err := NewCache(nil, Config{
		Topic: "basic",
	})
	require.NoError(t, err)

	err = c.Set("hello1", []byte("world"))
	require.NoError(t, err)

	val1, err := c.Get("hello1")
	require.NoError(t, err)
	require.NotEmpty(t, val1)
	require.Equal(t, val1, []byte("world"))

	c.Invalidate("hello1")
	_, err = c.Get("hello1")
	require.Error(t, err)
}

func TestInvalidateAllCache(t *testing.T) {
	c, err := NewCache(nil, Config{
		Topic: "invalidate",
	})
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		err = c.Set(fmt.Sprintf("hello%d", i), []byte("world"))
		require.NoError(t, err)
	}

	c.InvalidateAll()

	for i := 0; i < 5; i++ {
		_, err := c.Get(fmt.Sprintf("hello%d", i))
		require.Error(t, err, fmt.Sprintf("fetching %s", fmt.Sprintf("hello%d", i)))
	}
}

func TestFunctionCache(t *testing.T) {
	c, err := NewCache(nil, Config{
		Topic: "function",
		TTL:   time.Second,
	})
	require.NoError(t, err)

	value, err := c.GetFunction("hello1", fetcherOk)
	require.NoError(t, err)
	require.Equal(t, value, []byte("world"))

	// normal get should get it now
	value, err = c.Get("hello1")
	require.NoError(t, err)
	require.Equal(t, value, []byte("world"))

	_, err = c.GetFunction("hello2", fetcherFail)
	require.Error(t, err)
	_, err = c.Get("hello2")
	require.Error(t, err)

	value, err = c.GetFunction("hello3", fetcherOk)
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
