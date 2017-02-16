package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheWithoutDefaultExpiration(t *testing.T) {
	C := New(0, 0)

	item, found := C.Get("test")
	assert.Equal(t, found, false)
	assert.Equal(t, item, nil)

	C.Set("test", 123, 50*time.Millisecond)
	item, found = C.Get("test")
	assert.Equal(t, found, true)
	assert.Equal(t, item.(int), 123)

	<-time.After(100 * time.Millisecond)

	item, found = C.Get("test")
	assert.Equal(t, found, false)
	assert.Equal(t, item, nil)
}

func TestCacheWithDefaultExpiration(t *testing.T) {
	C := New(500*time.Millisecond, 200*time.Millisecond)

	item, found := C.Get("test")
	assert.Equal(t, found, false)
	assert.Equal(t, item, nil)

	C.Set("test", 123, 0)
	item, found = C.Get("test")
	assert.Equal(t, found, true)
	assert.Equal(t, item.(int), 123)

	<-time.After(100 * time.Millisecond)
	item, found = C.Get("test")
	assert.Equal(t, found, true)
	assert.Equal(t, item.(int), 123)

	<-time.After(1000 * time.Millisecond)
	item, found = C.Get("test")
	assert.Equal(t, found, false)
	assert.Equal(t, item, nil)
}

func TestCacheDelete(t *testing.T) {
	C := New(0, 0)

	item, found := C.Get("test")
	assert.Equal(t, found, false)
	assert.Equal(t, item, nil)

	C.Set("test", 123, 0)
	item, found = C.Get("test")
	assert.Equal(t, found, true)
	assert.Equal(t, item.(int), 123)

	C.Delete("test")
	item, found = C.Get("test")
	assert.Equal(t, found, false)
	assert.Equal(t, item, nil)

}

func BenchmarkCache(b *testing.B) {
	benchmarkSet(b, 0)
}

func BenchmarkCacheWithExpire(b *testing.B) {
	benchmarkSet(b, 50*time.Millisecond)
}

func benchmarkSet(b *testing.B, d time.Duration) {
	b.StopTimer()
	C := New(d, d)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		C.Set("key", "value", 500*time.Millisecond)
	}

}
