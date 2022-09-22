// cachelruttl is a naive implementation of a cache to use with awssecretmanager
//
// LRU: the size is fixed, it will start removing old entries when full.
//
// TTL: it will not return old entries (they will still be in the cache, but filtered out). Time precision: second.
package cachelruttl

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	c := New(2, time.Hour)
	c.Add("key1", "val1.A")
	val, ok := c.Get("key1")
	if !ok || val.(string) != "val1.A" {
		t.Fatal()
	}

	//replace
	c.Add("key1", "val1.B")
	val, ok = c.Get("key1")
	if !ok || val.(string) != "val1.B" {
		t.Fatal()
	}

	//check LRU
	c.Add("key2", "val2")
	if _, ok := c.Get("key2"); !ok {
		t.Fatal()
	}
	c.Add("key3", "val3")
	if _, ok := c.Get("key3"); !ok {
		t.Fatal()
	}
	//should have been removed
	if _, ok := c.Get("key1"); ok {
		t.Fatal()
	}

	//expired should not be available
	c.addWithExpired("key4", "val4", 10)
	if _, ok := c.Get("key4"); ok {
		t.Fatal()
	}
}
