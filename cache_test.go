package go_cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCache_EmptyAtStartup(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	a, err := tc.Get("sampleKeyA")
	assert.Nil(t, a)
	assert.ErrorIs(t, err, ErrItemNotFound)

	b, err := tc.Get("sampleKeyB")
	assert.Nil(t, b)
	assert.ErrorIs(t, err, ErrItemNotFound)

	c, err := tc.Get("sampleKeyC")
	assert.Nil(t, c)
	assert.ErrorIs(t, err, ErrItemNotFound)

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

		a, err := tc.Get("aKey")
		assert.Equal(t, "aValue", a)
		assert.Nil(t, err)

		b, err := tc.Get("bKey")
		assert.Equal(t, "bValue", b)
		assert.Nil(t, err)

		c, err := tc.Get("cKey")
		assert.Equal(t, 3, c)
		assert.Nil(t, err)

		tc.Delete("aKey")

		a, err = tc.Get("aKey")
		assert.Nil(t, a)
		assert.ErrorIs(t, err, ErrItemNotFound)
	})

	t.Run("withExpirationTime", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 1*time.Millisecond)
		defer tc.Stop()

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", DefaultExpiration)
		tc.Set("cKey", 3, DefaultExpiration)

		<-time.After(25 * time.Millisecond)

		a, err := tc.Get("aKey")
		assert.Nil(t, a)
		assert.ErrorIs(t, err, ErrItemNotFound)

		b, err := tc.Get("bKey")
		assert.Nil(t, b)
		assert.ErrorIs(t, err, ErrItemNotFound)

		c, err := tc.Get("cKey")
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrItemNotFound)
	})

	t.Run("withSeveralExpirationTimes", func(t *testing.T) {
		tc := NewCache(20*time.Millisecond, 1*time.Millisecond)
		defer tc.Stop()

		tc.Set("aKey", "aValue", DefaultExpiration)
		tc.Set("bKey", "bValue", NoExpiration)
		tc.Set("cKey", 3, 50*time.Millisecond)

		<-time.After(25 * time.Millisecond)

		a, err := tc.Get("aKey")
		assert.Nil(t, a)
		assert.ErrorIs(t, err, ErrItemNotFound)

		b, err := tc.Get("bKey")
		assert.Equal(t, "bValue", b)
		assert.Nil(t, err)

		c, err := tc.Get("cKey")
		assert.Equal(t, 3, c)
		assert.Nil(t, err)

		<-time.After(30 * time.Millisecond)

		c, err = tc.Get("cKey")
		assert.Nil(t, c)
		assert.ErrorIs(t, err, ErrItemNotFound)
	})
}

func TestCache_StorePointer(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	type TestStruct struct {
		Field int
	}

	tc.Set("testStruct", &TestStruct{1}, DefaultExpiration)

	x, err := tc.Get("testStruct")
	ts := x.(*TestStruct)
	assert.Equal(t, 1, ts.Field)
	assert.Nil(t, err)

	ts.Field++

	x, err = tc.Get("testStruct")
	ts = x.(*TestStruct)
	assert.Equal(t, 2, ts.Field)
	assert.Nil(t, err)
}

func TestCache_Delete(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	tc.Delete("notExistingKey")

	tc.Set("aKey", "aValue", DefaultExpiration)
	tc.Set("bKey", "bValue", DefaultExpiration)

	a, err := tc.Get("aKey")
	assert.Equal(t, "aValue", a)
	assert.Nil(t, err)

	b, err := tc.Get("bKey")
	assert.Equal(t, "bValue", b)
	assert.Nil(t, err)

	tc.Delete("aKey")
	tc.Delete("bKey")

	a, err = tc.Get("aKey")
	assert.Nil(t, a)
	assert.ErrorIs(t, err, ErrItemNotFound)

	b, err = tc.Get("bKey")
	assert.Nil(t, b)
	assert.ErrorIs(t, err, ErrItemNotFound)
}

func TestCache_Flush(t *testing.T) {
	tc := NewCache(NoExpiration, 0)
	defer tc.Stop()

	tc.Set("aKey", "aValue", DefaultExpiration)
	tc.Set("bKey", "bValue", DefaultExpiration)

	a, err := tc.Get("aKey")
	assert.Equal(t, "aValue", a)
	assert.Nil(t, err)

	b, err := tc.Get("bKey")
	assert.Equal(t, "bValue", b)
	assert.Nil(t, err)

	tc.Flush()

	a, err = tc.Get("aKey")
	assert.Nil(t, a)
	assert.ErrorIs(t, err, ErrItemNotFound)

	b, err = tc.Get("bKey")
	assert.Nil(t, b)
	assert.ErrorIs(t, err, ErrItemNotFound)
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
}
