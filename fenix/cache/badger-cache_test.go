package cache

import (
	"testing"
)

func TestBadgerCache_Exists(t *testing.T) {
	err := testBadgerCache.Remove("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache, should not exist")
	}

	err = testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}
	inCache, err = testBadgerCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found in cache")
	}

	err = testBadgerCache.Remove("foo")
	if err != nil {
		t.Error(err)
	}
}

func TestBadgerCache_Get(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	val, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if val != "bar" {
		t.Error("did not get correct value from cache")
	}
}

func TestBadgerCache_Remove(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Remove("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache, should not exists")
	}
}

func TestBadgerCache_Empty(t *testing.T) {
	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Empty()
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, should not exist")
	}
}

func TestBadgerCache_EmptyByMatch(t *testing.T) {
	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("alpha2", "beta2")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.EmptyByMatch("a")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, should not exist")
	}

	inCache, err = testBadgerCache.Exists("alpha2")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache, should not exist")
	}

	inCache, err = testBadgerCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found in cache, should exist")
	}
}
