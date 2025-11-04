package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
	ttl   time.Duration
}

var (
	Instance *Cache
	once     sync.Once
)

// Init inicializa el sistema de caché global
func Init(defaultTTL time.Duration) *Cache {
	once.Do(func() {
		Instance = &Cache{
			items: make(map[string]CacheItem),
			ttl:   defaultTTL,
		}
		// Limpiar caché expirado cada 5 minutos
		go Instance.cleanupExpired()
	})
	return Instance
}

// Get obtiene la instancia global del caché
func Get() *Cache {
	if Instance == nil {
		return Init(5 * time.Minute)
	}
	return Instance
}

// Set guarda un valor en caché
func (c *Cache) Set(key string, value interface{}, ttl ...time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	duration := c.ttl
	if len(ttl) > 0 {
		duration = ttl[0]
	}

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// GetValue obtiene un valor del caché
func (c *Cache) GetValue(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Verificar si expiró
	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Delete elimina un valor del caché
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// DeleteByPrefix elimina todas las claves que empiecen con un prefijo
func (c *Cache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}

// Clear limpia todo el caché
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
}

// cleanupExpired limpia items expirados periódicamente
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size retorna el número de items en caché
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Marshal serializa y guarda en caché
func (c *Cache) Marshal(key string, value interface{}, ttl ...time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.Set(key, data, ttl...)
	return nil
}

// Unmarshal obtiene y deserializa del caché
func (c *Cache) Unmarshal(key string, target interface{}) (bool, error) {
	data, found := c.GetValue(key)
	if !found {
		return false, nil
	}

	bytes, ok := data.([]byte)
	if !ok {
		return false, nil
	}

	if err := json.Unmarshal(bytes, target); err != nil {
		return false, err
	}

	return true, nil
}