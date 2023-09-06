package go_cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_EmptyAtStartup(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	a, found := tc.Get("sampleKeyA")
	assert.Nil(t, a)
	assert.False(t, found)

	b, found := tc.Get("sampleKeyB")
	assert.Nil(t, b)
	assert.False(t, found)

	c, found := tc.Get("sampleKeyC")
	assert.Nil(t, c)
	assert.False(t, found)

	ic := tc.ItemCount()
	assert.Equal(t, 0, ic)
}

func TestCache_SetAndGet(t *testing.T) {
	t.Run("withNoExpirationTime", func(t *testing.T) {
		tc := NewCache(NoExpiration, 0)
		defer tc.Stop()

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", DefaultExpiration)
		tc.Set("cKey", 3, DefaultExpiration)

		a, found := tc.Get("aKey")
		assert.Equal(t, "aValue", a)
		assert.True(t, found)

		b, found := tc.Get("bKey")
		assert.Equal(t, "bValue", b)
		assert.True(t, found)

		c, found := tc.Get("cKey")
		assert.Equal(t, 3, c)
		assert.True(t, found)

		tc.Delete("aKey")

		a, found = tc.Get("aKey")
		assert.Nil(t, a)
		assert.False(t, found)
	})

	t.Run("withExpirationTime", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 1*time.Millisecond)
		defer tc.Stop()

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", DefaultExpiration)
		tc.Set("cKey", 3, DefaultExpiration)

		<-time.After(25 * time.Millisecond)

		a, found := tc.Get("aKey")
		assert.Nil(t, a)
		assert.False(t, found)

		b, found := tc.Get("bKey")
		assert.Nil(t, b)
		assert.False(t, found)

		c, found := tc.Get("cKey")
		assert.Nil(t, c)
		assert.False(t, found)
	})

	t.Run("withSeveralExpirationTimes", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 1*time.Millisecond)
		defer tc.Stop()

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", NoExpiration)
		tc.Set("cKey", 3, 50*time.Millisecond)

		<-time.After(25 * time.Millisecond)

		a, False := tc.Get("aKey")
		assert.Nil(t, a)
		assert.False(t, False)

		b, False := tc.Get("bKey")
		assert.Equal(t, "bValue", b)
		assert.True(t, False)

		c, False := tc.Get("cKey")
		assert.Equal(t, 3, c)
		assert.True(t, False)

		<-time.After(30 * time.Millisecond)

		c, False = tc.Get("cKey")
		assert.Nil(t, c)
		assert.False(t, False)
	})

	t.Run("withSeveralExpirationTimesAndHigherCleanUpInterval", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 60*time.Millisecond)
		defer tc.Stop()

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", NoExpiration)
		tc.Set("cKey", 3, 50*time.Millisecond)

		<-time.After(25 * time.Millisecond)

		a, False := tc.Get("aKey")
		assert.Nil(t, a)
		assert.False(t, False)

		b, False := tc.Get("bKey")
		assert.Equal(t, "bValue", b)
		assert.True(t, False)

		c, False := tc.Get("cKey")
		assert.Equal(t, 3, c)
		assert.True(t, False)

		<-time.After(30 * time.Millisecond)

		c, False = tc.Get("cKey")
		assert.Nil(t, c)
		assert.False(t, False)
	})
}

func TestCache_StorePointer(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	type TestStruct struct {
		Field int
	}

	tc.Set("testStruct", &TestStruct{1}, DefaultExpiration)

	x, found := tc.Get("testStruct")
	ts := x.(*TestStruct)
	assert.Equal(t, 1, ts.Field)
	assert.True(t, found)

	ts.Field++

	x, found = tc.Get("testStruct")
	ts = x.(*TestStruct)
	assert.Equal(t, 2, ts.Field)
	assert.True(t, found)
}

func TestCache_Delete(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	tc.Delete("notExistingKey")

	tc.Set("aKey", "aValue", DefaultExpiration)
	tc.Set("bKey", "bValue", DefaultExpiration)

	a, found := tc.Get("aKey")
	assert.Equal(t, "aValue", a)
	assert.True(t, found)

	b, found := tc.Get("bKey")
	assert.Equal(t, "bValue", b)
	assert.True(t, found)

	tc.Delete("aKey")
	tc.Delete("bKey")

	a, found = tc.Get("aKey")
	assert.Nil(t, a)
	assert.False(t, found)

	b, found = tc.Get("bKey")
	assert.Nil(t, b)
	assert.False(t, found)
}

func TestCache_Flush(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	tc.Set("aKey", "aValue", DefaultExpiration)
	tc.Set("bKey", "bValue", DefaultExpiration)

	a, found := tc.Get("aKey")
	assert.Equal(t, "aValue", a)
	assert.True(t, found)

	b, found := tc.Get("bKey")
	assert.Equal(t, "bValue", b)
	assert.True(t, found)

	tc.Flush()

	a, found = tc.Get("aKey")
	assert.Nil(t, a)
	assert.False(t, found)

	b, found = tc.Get("bKey")
	assert.Nil(t, b)
	assert.False(t, found)
}

func TestCache_ItemCount(t *testing.T) {
	t.Run("withDelete", func(t *testing.T) {
		tc := NewCache(NoExpiration, 0)
		defer tc.Stop()

		ic := tc.ItemCount()
		assert.Equal(t, 0, ic)

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", DefaultExpiration)

		ic = tc.ItemCount()
		assert.Equal(t, 2, ic)

		tc.Delete("aKey")

		ic = tc.ItemCount()
		assert.Equal(t, 1, ic)

		tc.Delete("bKey")

		ic = tc.ItemCount()
		assert.Equal(t, 0, ic)
	})

	t.Run("withFlush", func(t *testing.T) {
		tc := NewCache(NoExpiration, 0)
		defer tc.Stop()

		ic := tc.ItemCount()
		assert.Equal(t, 0, ic)

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", DefaultExpiration)

		ic = tc.ItemCount()
		assert.Equal(t, 2, ic)

		tc.Flush()

		ic = tc.ItemCount()
		assert.Equal(t, 0, ic)
	})

	t.Run("withExpirationTime", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 1*time.Millisecond)
		defer tc.Stop()

		ic := tc.ItemCount()
		assert.Equal(t, 0, ic)

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", DefaultExpiration)

		ic = tc.ItemCount()
		assert.Equal(t, 2, ic)

		<-time.After(25 * time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 0, ic)
	})

	t.Run("withSeveralExpirationTimes", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 1*time.Millisecond)
		defer tc.Stop()

		ic := tc.ItemCount()
		assert.Equal(t, 0, ic)

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", NoExpiration)
		tc.Set("cKey", 3, 50*time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 3, ic)

		<-time.After(25 * time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 2, ic)

		<-time.After(30 * time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 1, ic)
	})

	t.Run("withSeveralExpirationTimesAndHigherCleanUpInterval", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 60*time.Millisecond)
		defer tc.Stop()

		ic := tc.ItemCount()
		assert.Equal(t, 0, ic)

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", NoExpiration)
		tc.Set("cKey", 3, 50*time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 3, ic)

		<-time.After(25 * time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 3, ic)

		<-time.After(30 * time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 3, ic)

		<-time.After(100 * time.Millisecond)

		ic = tc.ItemCount()
		assert.Equal(t, 1, ic)
	})
}
