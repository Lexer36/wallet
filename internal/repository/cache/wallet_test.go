package cache_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"wallet/internal/repository/cache"
)

type MockLRUCache struct {
	data map[string]int64
}

func (m *MockLRUCache) Add(key interface{}, value interface{}) (evicted bool) {
	if m.data == nil {
		m.data = make(map[string]int64)
	}
	m.data[key.(string)] = value.(int64)
	return false
}

func (m *MockLRUCache) Get(key interface{}) (value interface{}, ok bool) {
	val, ok := m.data[key.(string)]
	return val, ok
}

func (m *MockLRUCache) Remove(key interface{}) (present bool) {
	if _, ok := m.data[key.(string)]; ok {
		delete(m.data, key.(string))
		return true
	}
	return false
}

func TestCache(t *testing.T) {
	t.Parallel()

	mockCache := &MockLRUCache{}
	cache := cache.New(mockCache)

	cache.Set(t.Context(), "wallet1", 1000)
	value, ok := mockCache.Get("wallet1")
	require.True(t, ok)
	require.Equal(t, int64(1000), value)

	// Тест для метода Get
	balance, ok := cache.Get(t.Context(), "wallet1")
	require.True(t, ok)
	require.Equal(t, int64(1000), balance)

	// Тест для метода Delete
	cache.Delete(t.Context(), "wallet1")
	_, ok = mockCache.Get("wallet1")
	require.False(t, ok)

	// Проверка, что после удаления значение больше не доступно
	_, ok = cache.Get(t.Context(), "wallet1")
	require.False(t, ok)
}
