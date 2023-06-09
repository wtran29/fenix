package cache

import "testing"

func TestRedisCache_Exists(t *testing.T) {
	err := testRedisCache.Remove("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo in cache, should not be there")
	}

	err = testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	inCache, err = testRedisCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not in cache, should be there")
	}

}

func TestRedisCache_Get(t *testing.T) {
	err := testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	v, err := testRedisCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if v != "bar" {
		t.Error("incorrect value from cache")
	}
}

func TestRedisCache_Remove(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Remove("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found, should not be in cache")
	}
}

func TestRedisCache_Empty(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Empty()
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found, should not be in cache")
	}
}

func TestRedisCache_EmptyByMatch(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("alpha2", "cat")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.EmptyByMatch("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Exists("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found, should not be in cache")
	}

	inCache, err = testRedisCache.Exists("alpha2")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha2 found, should not be in cache")
	}

	inCache, err = testRedisCache.Exists("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found, should be in cache")
	}

}

func TestEncodeDecode(t *testing.T) {
	entry := Entry{}
	entry["foo"] = "bar"
	bytes, err := encode(entry)
	if err != nil {
		t.Error(err)
	}

	_, err = decode(string(bytes))
	if err != nil {
		t.Error(err)
	}
}
