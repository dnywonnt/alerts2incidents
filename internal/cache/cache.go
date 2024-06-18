package cache // dnywonnt.me/alerts2incidents/internal/cache

import (
	"container/list"
	"sync"

	log "github.com/sirupsen/logrus"
)

// CacheItem represents a single item in the cache, holding a key and its associated value.
type CacheItem struct {
	Key   string
	Value interface{}
}

// Cache defines the structure of the cache, including a map for quick lookup, a doubly linked list to maintain order,
// a mutex for concurrent access, a size limit, and a tag for logging.
type Cache struct {
	itemsMap map[string]*list.Element // Map for O(1) lookup of cache items.
	itemList *list.List               // Doubly linked list to maintain the order of items for eviction purposes.
	mu       sync.RWMutex             // Mutex for safe concurrent access.
	size     int                      // Maximum number of items the cache can hold.
	tag      string                   // An optional tag for identifying the cache in logs.
}

// NewCache initializes a new cache with the specified size and tag, and logs the initialization event.
func NewCache(size int, tag string) *Cache {
	log.WithFields(log.Fields{
		"size": size,
		"tag":  tag,
	}).Debug("Initializing a new cache")

	return &Cache{
		itemsMap: make(map[string]*list.Element),
		itemList: list.New(),
		size:     size,
		tag:      tag,
	}
}

// SetItem adds a new item to the cache or updates an existing one, while managing the size limit and eviction policy.
func (c *Cache) SetItem(key string, value interface{}) {
	c.mu.Lock() // Ensuring thread-safe access to the cache.
	defer c.mu.Unlock()

	// If the cache has a size limit and is full, remove the least recently used (LRU) item.
	if c.size > 0 && c.itemList.Len() == c.size {
		lastItem := c.itemList.Back()
		if lastItem != nil {
			// Log the removal of the last item.
			log.WithFields(log.Fields{
				"key": lastItem.Value.(*CacheItem).Key,
				"tag": c.tag,
			}).Debug("Removing the last item from cache")

			c.itemList.Remove(lastItem)
			delete(c.itemsMap, lastItem.Value.(*CacheItem).Key)
		}
	}

	// If the item already exists, update it and move it to the front of the list to mark it as recently used.
	if elem, found := c.itemsMap[key]; found {
		log.WithFields(log.Fields{
			"key": key,
			"tag": c.tag,
		}).Debug("Key found; updating existing item in the cache")
		c.itemList.MoveToFront(elem)
		elem.Value = &CacheItem{Key: key, Value: value}
	} else {
		// If the item is new, add it to the front of the list and to the map.
		log.WithFields(log.Fields{
			"key": key,
			"tag": c.tag,
		}).Debug("Key not found; adding new item to the cache")
		elem := c.itemList.PushFront(&CacheItem{Key: key, Value: value})
		c.itemsMap[key] = elem
	}
}

// GetItem retrieves an item from the cache based on its key, marking the item as recently used if found.
func (c *Cache) GetItem(key string) (*CacheItem, bool) {
	c.mu.Lock() // Locking for thread-safe write access because we'll modify the list order.
	defer c.mu.Unlock()

	log.WithFields(log.Fields{
		"key": key,
		"tag": c.tag,
	}).Debug("Retrieving item from cache")

	// Return the item if found. Move it to the front of the list to mark as recently used.
	if elem, found := c.itemsMap[key]; found {
		// Move the accessed item to the front of the list to indicate recent use.
		c.itemList.MoveToFront(elem)
		return elem.Value.(*CacheItem), true
	}

	return nil, false
}

// DeleteItem removes an item from the cache based on its key.
func (c *Cache) DeleteItem(key string) {
	c.mu.Lock() // Ensuring thread-safe write access to the cache.
	defer c.mu.Unlock()

	log.WithFields(log.Fields{
		"key": key,
		"tag": c.tag,
	}).Debug("Deleting item from cache")

	// If the item is found, remove it from both the map and the list.
	if elem, found := c.itemsMap[key]; found {
		c.itemList.Remove(elem)
		delete(c.itemsMap, key)
	}
}

// GetAllItems returns a slice containing all items in the cache.
func (c *Cache) GetAllItems() []*CacheItem {
	c.mu.RLock() // Ensuring thread-safe read access.
	defer c.mu.RUnlock()

	log.WithFields(log.Fields{
		"tag": c.tag,
	}).Debug("Retrieving all items from cache")

	items := []*CacheItem{}

	// Iterate over the list and add each item to the slice.
	for element := c.itemList.Front(); element != nil; element = element.Next() {
		items = append(items, element.Value.(*CacheItem))
	}

	return items
}

// GetItems returns a slice of CacheItems from the cache, based on the specified pagination parameters.
func (c *Cache) GetItems(pageNum int, pageSize int) []*CacheItem {
	c.mu.RLock() // Ensuring thread-safe read access.
	defer c.mu.RUnlock()

	log.WithFields(log.Fields{
		"pageNum":  pageNum,
		"pageSize": pageSize,
		"tag":      c.tag,
	}).Debug("Retrieving items from cache")

	items := []*CacheItem{}
	startIndex := (pageNum - 1) * pageSize
	endIndex := startIndex + pageSize
	currentIndex := 0

	// Iterate over the list and add items within the requested page to the slice.
	for element := c.itemList.Front(); element != nil; element = element.Next() {
		if currentIndex >= startIndex && currentIndex < endIndex {
			items = append(items, element.Value.(*CacheItem))
		}
		if currentIndex >= endIndex {
			break
		}
		currentIndex++
	}

	return items
}

// GetTotalItems returns the total number of items currently in the cache.
func (c *Cache) GetTotalItems() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalItems := len(c.itemsMap)

	log.WithFields(log.Fields{
		"totalItems": totalItems,
		"tag":        c.tag,
	}).Debug("Counting total items in cache")

	return totalItems
}

// Clear removes all items from the cache.
func (c *Cache) Clear() {
	c.mu.Lock() // Ensuring thread-safe write access to the cache.
	defer c.mu.Unlock()

	log.WithFields(log.Fields{
		"tag": c.tag,
	}).Debug("Clearing the cache")

	// Remove all items from the map and list.
	c.itemsMap = make(map[string]*list.Element)
	c.itemList.Init()
}

// GetMaxSize returns the maximum size of the cache.
func (c *Cache) GetMaxSize() int {
	c.mu.RLock() // Ensuring thread-safe read access.
	defer c.mu.RUnlock()

	log.WithFields(log.Fields{
		"tag": c.tag,
	}).Debug("Retrieving the maximum size of the cache")

	return c.size
}
