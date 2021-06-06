package lru_cache

import (
	"container/list"
	"fmt"
	"time"
)

type Entry struct {
	Key   uint32
	Value string
}

type LRUCache struct {
	capacity int
	ttl      int
	queue    *list.List // список всех записей в порядке от старых к новым
	storage  map[uint32]*list.Element
}

func NewCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      -1,
		storage:  make(map[uint32]*list.Element),
		queue:    list.New(),
	}
}

func NewCacheTTL(capacity int, ttl int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		storage:  make(map[uint32]*list.Element),
		queue:    list.New(),
	}
}

func NewCacheMemLimit(size int, ttl int) *LRUCache {
	/*
		Оценим приблизительно, сколько памяти может занимать кэш из N элементов.
		1. Каждый entry содержит 4 байта для ключа и 8 байт для указателя на строку, итого 12 байт.
		2. queue - двусвязный список, т.е. на каждый элемент приходится 2 указателя на соседние узлы и значение entry.
		   Считая, что размер указателя - 8 байт, получаем (8 * 2 + 12) * N байт памяти для N элементов.
		3. storage - хеш-таблица, в общем случае ее размер зависит от реализации и количества коллизий.
		   В качестве оценки снизу можно рассмотерть случай идеальной хеш-функции,
		   когда в таблице имеется ровно N бакетов по 1 элементу каждый. Тогда имеем:
		   (8 * N) - размер массива указателей на бакеты;
		   (8 * N) - суммарный размер всех бакетов (каждый бакет состоит из одного указателя на узел queue).
		   Итого (16 * N) байт для N элементов.
		4. Строки длиной 128 имеют размер 128 * 2 каждая, если считать, что каждый символ кодируется двумя байтами.
		Итого: size = (28 + 8 + 8 + 256) * N = 300 * N
	*/
	return &LRUCache{
		capacity: size / 300,
		ttl:      ttl,
		storage:  make(map[uint32]*list.Element),
		queue:    list.New(),
	}
}

func (cache LRUCache) RemoveOldest() (string, bool) {
	if cache.queue.Len() > 0 {
		oldest_node := cache.queue.Back()
		entry := oldest_node.Value.(Entry)
		delete(cache.storage, entry.Key)
		cache.queue.Remove(oldest_node)
		return entry.Value, true
	} else {
		return "", false
	}
}

func (cache LRUCache) Remove(key uint32) (string, bool) {
	node, exists := cache.storage[key]
	if exists {
		cache.queue.Remove(node)
		entry := node.Value.(Entry)
		delete(cache.storage, entry.Key)
		return entry.Value, true
	} else {
		return "", false
	}
}

func (cache LRUCache) Put(key uint32, value string) {
	node, exists := cache.storage[key]
	if exists {
		// Обновляем существующую запись
		cache.queue.MoveToFront(node)
		node.Value = Entry{Key: key, Value: value}
	} else {
		// Удаляем самую старую запись, если кэш переполнен
		if cache.queue.Len() == cache.capacity {
			cache.RemoveOldest()
		}
		new_entry := Entry{Key: key, Value: value}
		cache.storage[key] = cache.queue.PushFront(new_entry)
	}
	// Запускаем таймер удаления записи
	if cache.ttl != -1 {
		timer := time.NewTimer(time.Second * time.Duration(cache.ttl))
		go func() {
			<-timer.C
			cache.Remove(key)
			timer.Stop()
		}()
	}
}

func (cache LRUCache) Get(key uint32) (string, bool) {
	node, exists := cache.storage[key]
	if exists {
		// "Освежаем" запись, к которой было обращение
		cache.queue.MoveToFront(node)
		return node.Value.(Entry).Value, true
	} else {
		return "", false
	}
}

func (cache LRUCache) Print() {
	fmt.Print("{")
	for node := cache.queue.Front(); node != nil; node = node.Next() {
		entry := node.Value.(Entry)
		fmt.Printf("%d: %s", entry.Key, entry.Value)
		if node != cache.queue.Back() {
			fmt.Print(", ")
		}
	}
	fmt.Println("}")
}
