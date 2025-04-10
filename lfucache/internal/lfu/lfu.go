package lfu

import (
	"errors"
	"iter"
)

var ErrKeyNotFound = errors.New("key not found")

type Node[K comparable] struct {
	Key  K
	Prev *Node[K]
	Next *Node[K]
}

// In Linkedlist we will have a map which helps to find fastly the proper node by key in that list
type LinkedList[K comparable] struct {
	head   *Node[K]
	tail   *Node[K]
	length int
	nodes  map[K]*Node[K]
}

// Constructor for creating dummy Linkedlist instance
func NewLinkedList[K comparable]() *LinkedList[K] {
	head := &Node[K]{}
	tail := &Node[K]{}
	head.Next = tail
	tail.Prev = head
	return &LinkedList[K]{
		head:  head,
		tail:  tail,
		nodes: make(map[K]*Node[K]),
	}
}

type LinkedListInterface[K comparable] interface {
	PushLeft(key K)
	PushRight(key K)
	PushNextTo(givenKey K, newKey K)
	PopLeft() (K, error)
	Pop(key K) error
	Len() int
}

// Inserting a new node into the linked list between prev and next
func (ll *LinkedList[K]) addNode(key K, prev, next *Node[K]) {
	// Checking if the key already exists
	if _, ok := ll.nodes[key]; ok {
		return
	}

	// Creating a new node for the key
	node := &Node[K]{Key: key}

	// Inserting the new node between prev and next
	node.Prev = prev
	node.Next = next
	prev.Next = node
	next.Prev = node

	// Updating the length and map
	ll.length++
	ll.nodes[key] = node
}

// Pushing a key to the front of the linkedlist
func (ll *LinkedList[K]) PushLeft(key K) {
	ll.addNode(key, ll.head, ll.head.Next)
}

// Pushing a key to the end of the linkedlist
func (ll *LinkedList[K]) PushRight(key K) {
	ll.addNode(key, ll.tail.Prev, ll.tail)
}

// Pushing a key right after the given node for fast adding to the linkedlist
func (ll *LinkedList[K]) PushNextTo(givenKey K, newKey K) {
	// Checking if the existing key is in the list
	if givenKey, ok := ll.nodes[givenKey]; !ok {
		ll.PushRight(newKey)
		return
	} else {
		ll.addNode(newKey, givenKey, givenKey.Next)
	}
	// Else inserting new key after the given key
}

// removeNode safely removes a node from the linked list, adjusting the neighbors and map
func (ll *LinkedList[K]) removeNode(node *Node[K]) {
	if node == nil {
		return
	}
	// Updating the neighbours
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev

	// Removing the node from the map and decrementing length
	delete(ll.nodes, node.Key)
	ll.length--
}

// Popping the key from the head of the list
func (ll *LinkedList[K]) PopLeft() (K, error) {
	// Checking if the list is empty
	if ll.head.Next == ll.tail {
		return *new(K), ErrKeyNotFound
	}

	// Updating the right neighbour
	node := ll.head.Next
	ll.removeNode(node)
	return node.Key, nil
}

// Pop removes a specific key from the linked list
func (ll *LinkedList[K]) Pop(key K) error {
	// Find the node in the map
	if node, ok := ll.nodes[key]; ok {
		ll.removeNode(node)
		return nil
	}
	return ErrKeyNotFound
}

func (ll *LinkedList[K]) Len() int {
	return ll.length
}

const DefaultCapacity = 5

// Cache
// O(capacity) memory
type Cache[K comparable, V any] interface {
	// Get returns the value of the key if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1)
	Get(key K) (V, error)

	// Put updates the value of the key if present, or inserts the key if not already present.
	//
	// When the cache reaches its capacity, it should invalidate and remove the least frequently used key
	// before inserting a new item. For this problem, when there is a tie
	// (i.e., two or more keys with the same frequency), the least recently used key would be invalidated.
	//
	// O(1)
	Put(key K, value V)

	// All returns the iterator in descending order of frequency.
	// If two or more keys have the same frequency, the most recently used key will be listed first.
	//
	// O(capacity)
	All() iter.Seq2[K, V]

	// Size returns the cache size.
	//
	// O(1)
	Size() int

	// Capacity returns the cache capacity.
	//
	// O(1)
	Capacity() int

	// GetKeyFrequency returns the element's frequency if the key exists in the cache,
	// otherwise, returns ErrKeyNotFound.
	//
	// O(1)
	GetKeyFrequency(key K) (int, error)
}

// cacheImpl represents LFU cache implementation
type cacheImpl[K comparable, V any] struct {
	capacity         int
	size             int
	minFreq          int
	values           map[K]V
	frequencies      map[K]int
	frequenciesOrder *LinkedList[int]
	lists            map[int]*LinkedList[K]
}

// New initializes the cache with the given capacity.
// If no capacity is provided, the cache will use DefaultCapacity.
func New[K comparable, V any](capacity ...int) *cacheImpl[K, V] {
	// Setting default value of capacity as the cache limit
	cacheLimit := DefaultCapacity
	// Checking if non-default capacity is provided and setting that value
	if len(capacity) > 0 {
		cacheLimit = capacity[0]
	}
	// Ensuring if cacheLimit is valid
	if cacheLimit < 0 {
		panic("Capacity must be more than zero")
	}
	return &cacheImpl[K, V]{
		// Sets the maximum capacity of the cache
		capacity: cacheLimit,
		// Stores the key-value pairs in the cache for fast retrieve of values by key
		values: make(map[K]V),
		// Stores the key-value pairs in the cache for fast retrieve of frequency by key
		frequencies: make(map[K]int),
		// Orders frequencies for easier management of least frequently used elements
		frequenciesOrder: NewLinkedList[int](),
		// Maps a frequency to a list of keys that have that frequency, allowing efficient access and removal
		lists: make(map[int]*LinkedList[K]),
	}
}

func (l *cacheImpl[K, V]) Get(key K) (V, error) {
	// Checking if our key in cache
	// If found, the frequency is updated and the value is returned
	if value, ok := l.values[key]; ok {
		l.updateFrequency(key)
		return value, nil
	} else {
		return *new(V), ErrKeyNotFound
	}
	// If not found, the error is thrown
}

func (l *cacheImpl[K, V]) updateFrequency(key K) {
	// Retrieving the frequency which needed to be incremented
	curFreq := l.frequencies[key]
	// Popping the frequency from list to place it to incremented frequency list
	err := l.lists[curFreq].Pop(key)
	if err != nil {
		return
	}
	newFreq := curFreq + 1
	l.frequencies[key] = newFreq
	// Initialising the list of new frequency if didn't exist before
	if l.lists[newFreq] == nil {
		l.lists[newFreq] = NewLinkedList[K]()
		// Updating frequenciesOrder with a new frequency for iterator,
		// because we want to know all existing frequencies in the sorted order to properly iterate over it
		l.frequenciesOrder.PushNextTo(curFreq, newFreq)
	}
	// Placing the key in the new list after incrementing the frequency
	l.lists[newFreq].PushRight(key)
	// Handling the case when the previous list becomes empty, so we just delete it and updating frequenciesOrder
	if l.lists[curFreq].Len() == 0 {
		err := l.frequenciesOrder.Pop(curFreq)
		if err != nil {
			return
		}
		delete(l.lists, curFreq)
		// Handling the case when our minFreq is not valid anymore, and we need to update it
		// Basically, it happens when the minimum frequency list is deleted
		if l.minFreq == curFreq {
			l.minFreq++
		}
	}
}

func (l *cacheImpl[K, V]) Put(key K, value V) {
	if l.capacity == 0 {
		return
	}
	// Checking if the key does not exist and the cache reached the capacity
	if _, ok := l.values[key]; !ok && len(l.values) == l.capacity {
		// Retrieving the LFU list which is the list with minimum frequency
		minFreqList, ok := l.lists[l.minFreq]
		if ok && minFreqList.Len() > 0 {
			// Trying to evict the least frequently used elem
			if evictedKey, err := minFreqList.PopLeft(); err == nil {
				// Removing the evicted key from values and frequencies maps
				delete(l.values, evictedKey)
				delete(l.frequencies, evictedKey)
			}
		}
	}
	// Setting the new value for the given key in the cache
	l.values[key] = value
	// Checking if the key does not already exist in the frequency map
	if _, ok := l.frequencies[key]; !ok {
		// Add frequency 1 to the frequency order list
		l.frequenciesOrder.PushLeft(1)
		// As the key did not exist before its frequency is set to 1
		l.frequencies[key] = 1
		l.minFreq = 1
		// Incrementing the size after putting
		l.size++
		// Initialize a new linked list for frequency 1 if it does not exist
		if l.lists[1] == nil {
			l.lists[1] = NewLinkedList[K]()
		}
		// Add the key to linked list
		l.lists[1].PushRight(key)
	} else {
		// If the key already exists, update its frequency
		l.updateFrequency(key)
	}
}
func (l *cacheImpl[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		// Iterating over frequenciesOrder list from highest to lowest frequency
		for orderNode := l.frequenciesOrder.tail.Prev; orderNode != l.frequenciesOrder.head; orderNode = orderNode.Prev {
			freq := orderNode.Key
			// Checking if there are keys in the current frequency list
			if keysList, ok := l.lists[freq]; !ok || keysList.Len() == 0 {
				continue
			}
			// Yielding pairs for each key in the current frequency list
			if !l.yieldPairs(freq, yield) {
				return
			}
		}
	}
}

// Iterating over the keys in the frequency list
// The function returns false to stop iteration, otherwise, it returns true after completing the iteration.
func (l *cacheImpl[K, V]) yieldPairs(freq int, yield func(K, V) bool) bool {
	keysList := l.lists[freq]
	// As we need to iterate from the most recently used key we go from tail to head within one frequency
	for elem := keysList.tail.Prev; elem != keysList.head; elem = elem.Prev {
		key := elem.Key
		if value, ok := l.values[key]; ok {
			// Calling the yield function with the current key and value
			if !yield(key, value) {
				return false
			}
		}
	}
	return true
}

func (l *cacheImpl[K, V]) Size() int {
	return l.size
}

func (l *cacheImpl[K, V]) Capacity() int {
	return l.capacity
}

func (l *cacheImpl[K, V]) GetKeyFrequency(key K) (int, error) {
	// Checking if we have the needed key in cache and retrieving its value
	if freq, ok := l.frequencies[key]; ok {
		return freq, nil
	} else {
		return 0, ErrKeyNotFound
	}
}
