package goutil

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Map is a concurrent map with loads, stores, and deletes.
// It is safe for multiple goroutines to call a Map's methods concurrently.
type Map interface {
	// Load returns the value stored in the map for a key, or nil if no
	// value is present.
	// The ok result indicates whether value was found in the map.
	Load(key interface{}) (value interface{}, ok bool)
	// Store sets the value for a key.
	Store(key, value interface{})
	// LoadOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored.
	LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
	// Range calls f sequentially for each key and value present in the map.
	// If f returns false, range stops the iteration.
	Range(f func(key, value interface{}) bool)
	// Random returns a pair kv randomly.
	// If exist=false, no kv data is exist.
	Random() (key, value interface{}, exist bool)
	// Delete deletes the value for a key.
	Delete(key interface{})
	// Clear clears all current data in the map.
	Clear()
	// Len returns the length of the map.
	Len() int
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RwMap creates a new concurrent safe map with sync.RWMutex.
// The normal Map is high-performance mapping under low concurrency conditions.
func RwMap(capacity ...int) Map {
	var cap int
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	return &rwMap{
		data: make(map[interface{}]interface{}, cap),
	}
}

// rwMap concurrent secure data storage,
// which is high-performance mapping under low concurrency conditions.
type rwMap struct {
	data map[interface{}]interface{}
	rwmu sync.RWMutex
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *rwMap) Load(key interface{}) (value interface{}, ok bool) {
	m.rwmu.RLock()
	value, ok = m.data[key]
	m.rwmu.RUnlock()
	return value, ok
}

// Store sets the value for a key.
func (m *rwMap) Store(key, value interface{}) {
	m.rwmu.Lock()
	m.data[key] = value
	m.rwmu.Unlock()
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *rwMap) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	m.rwmu.Lock()
	actual, loaded = m.data[key]
	if !loaded {
		m.data[key] = value
		actual = value
	}
	m.rwmu.Unlock()
	return actual, loaded
}

// Delete deletes the value for a key.
func (m *rwMap) Delete(key interface{}) {
	m.rwmu.Lock()
	delete(m.data, key)
	m.rwmu.Unlock()
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *rwMap) Range(f func(key, value interface{}) bool) {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}

// Clear clears all current data in the map.
func (m *rwMap) Clear() {
	m.rwmu.Lock()
	for k := range m.data {
		delete(m.data, k)
	}
	m.rwmu.Unlock()
}

// Random returns a pair kv randomly.
// If exist=false, no kv data is exist.
func (m *rwMap) Random() (key, value interface{}, exist bool) {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()
	length := len(m.data)
	if length == 0 {
		return
	}
	i := rand.Intn(length)
	for key, value = range m.data {
		if i == 0 {
			exist = true
			return
		}
		i--
	}
	return
}

// Len returns the length of the map.
// Note: the count is accurate.
func (m *rwMap) Len() int {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()
	return len(m.data)
}

// AtomicMap creates a concurrent map with amortized-constant-time loads, stores, and deletes.
// It is safe for multiple goroutines to call a atomicMap's methods concurrently.
// From go v1.9 sync.Map.
func AtomicMap() Map {
	return new(atomicMap)
}

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// atomicMap is a concurrent map with amortized-constant-time loads, stores, and deletes.
// It is safe for multiple goroutines to call a atomicMap's methods concurrently.
//
// It is optimized for use in concurrent loops with keys that are
// stable over time, and either few steady-state stores, or stores
// localized to one goroutine per key.
//
// For use cases that do not share these attributes, it will likely have
// comparable or worse performance and worse type safety than an ordinary
// map paired with a read-write mutex.
//
// The zero atomicMap is valid and empty.
//
// A atomicMap must not be copied after first use.
type atomicMap struct {
	mu sync.Mutex

	// read contains the portion of the map's contents that are safe for
	// concurrent access (with or without mu held).
	//
	// The read field itself is always safe to load, but must only be stored with
	// mu held.
	//
	// Entries stored in read may be updated concurrently without mu, but updating
	// a previously-expunged entry requires that the entry be copied to the dirty
	// map and unexpunged with mu held.
	read atomic.Value // readOnly

	// dirty contains the portion of the map's contents that require mu to be
	// held. To ensure that the dirty map can be promoted to the read map quickly,
	// it also includes all of the non-expunged entries in the read map.
	//
	// Expunged entries are not stored in the dirty map. An expunged entry in the
	// clean map must be unexpunged and added to the dirty map before a new value
	// can be stored to it.
	//
	// If the dirty map is nil, the next write to the map will initialize it by
	// making a shallow copy of the clean map, omitting stale entries.
	dirty map[interface{}]*entry

	// misses counts the number of loads since the read map was last updated that
	// needed to lock mu to determine whether the key was present.
	//
	// Once enough misses have occurred to cover the cost of copying the dirty
	// map, the dirty map will be promoted to the read map (in the unamended
	// state) and the next store to the map will make a new dirty copy.
	misses int

	// @added by henrylee2cn 2017/11/17
	length int32
}

// readOnly is an immutable struct stored atomically in the atomicMap.read field.
type readOnly struct {
	m       map[interface{}]*entry
	amended bool // true if the dirty map contains some key not in m.
}

// expunged is an arbitrary pointer that marks entries which have been deleted
// from the dirty map.
var expunged = unsafe.Pointer(new(interface{}))

// An entry is a slot in the map corresponding to a particular key.
type entry struct {
	// p points to the interface{} value stored for the entry.
	//
	// If p == nil, the entry has been deleted and m.dirty == nil.
	//
	// If p == expunged, the entry has been deleted, m.dirty != nil, and the entry
	// is missing from m.dirty.
	//
	// Otherwise, the entry is valid and recorded in m.read.m[key] and, if m.dirty
	// != nil, in m.dirty[key].
	//
	// An entry can be deleted by atomic replacement with nil: when m.dirty is
	// next created, it will atomically replace nil with expunged and leave
	// m.dirty[key] unset.
	//
	// An entry's associated value can be updated by atomic replacement, provided
	// p != expunged. If p == expunged, an entry's associated value can be updated
	// only after first setting m.dirty[key] = e so that lookups using the dirty
	// map find the entry.
	p unsafe.Pointer // *interface{}
}

func newEntry(i interface{}) *entry {
	return &entry{p: unsafe.Pointer(&i)}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *atomicMap) Load(key interface{}) (value interface{}, ok bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.load()
}

func (e *entry) load() (value interface{}, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expunged {
		return nil, false
	}
	return *(*interface{})(p), true
}

// Store sets the value for a key.
func (m *atomicMap) Store(key, value interface{}) {
	read, _ := m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		switch e.tryStore(&value) {
		case 1:
			return
		case 2:
			// @added by henrylee2cn 2017/11/17
			atomic.AddInt32(&m.length, 1)
			return
		}
	}

	m.mu.Lock()
	read, _ = m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		switch e.tryStore(&value) {
		case 1:
			m.mu.Unlock()
			return
		case 2:
			// @added by henrylee2cn 2017/11/17
			atomic.AddInt32(&m.length, 1)
			m.mu.Unlock()
			return
		case 0:
			if e.unexpungeLocked() {
				// The entry was previously expunged, which implies that there is a
				// non-nil dirty map and this entry is not in it.
				m.dirty[key] = e
				// @added by henrylee2cn 2017/11/17
				atomic.AddInt32(&m.length, 1)
			}
			e.storeLocked(&value)
		}

	} else if e, ok := m.dirty[key]; ok {
		e.storeLocked(&value)
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
		atomic.AddInt32(&m.length, 1)
	}
	m.mu.Unlock()
}

// tryStore stores a value if the entry has not been expunged.
//
// If the entry is expunged, tryStore returns 0 and leaves the entry
// unchanged.
func (e *entry) tryStore(i *interface{}) int8 {
	p := atomic.LoadPointer(&e.p)
	if p == expunged {
		return 0
	}
	for {
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			if p == nil {
				return 2
			}
			return 1
		}
		p = atomic.LoadPointer(&e.p)
		if p == expunged {
			return 0
		}
	}
}

// unexpungeLocked ensures that the entry is not marked as expunged.
//
// If the entry was previously expunged, it must be added to the dirty map
// before m.mu is unlocked.
func (e *entry) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&e.p, expunged, nil)
}

// storeLocked unconditionally stores a value to the entry.
//
// The entry must be known not to be expunged.
func (e *entry) storeLocked(i *interface{}) {
	atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *atomicMap) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	// Avoid locking if it's a clean hit.
	read, _ := m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			// @added by henrylee2cn 2017/11/17
			if !loaded {
				atomic.AddInt32(&m.length, 1)
			}
			return actual, loaded
		}
	}

	m.mu.Lock()
	read, _ = m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			m.dirty[key] = e
		}
		actual, loaded, ok = e.tryLoadOrStore(value)
		// @added by henrylee2cn 2017/12/01
		if ok && !loaded {
			atomic.AddInt32(&m.length, 1)
		}
	} else if e, ok := m.dirty[key]; ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		m.missLocked()
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
		atomic.AddInt32(&m.length, 1)
		actual, loaded = value, false
	}
	m.mu.Unlock()

	return actual, loaded
}

// tryLoadOrStore atomically loads or stores a value if the entry is not
// expunged.
//
// If the entry is expunged, tryLoadOrStore leaves the entry unchanged and
// returns with ok==false.
func (e *entry) tryLoadOrStore(i interface{}) (actual interface{}, loaded, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == expunged {
		return nil, false, false
	}
	if p != nil {
		return *(*interface{})(p), true, true
	}

	// Copy the interface after the first load to make this method more amenable
	// to escape analysis: if we hit the "load" path or the entry is expunged, we
	// shouldn't bother heap-allocating.
	ic := i
	for {
		if atomic.CompareAndSwapPointer(&e.p, nil, unsafe.Pointer(&ic)) {
			return i, false, true
		}
		p = atomic.LoadPointer(&e.p)
		if p == expunged {
			return nil, false, false
		}
		if p != nil {
			return *(*interface{})(p), true, true
		}
	}
}

// Delete deletes the value for a key.
func (m *atomicMap) Delete(key interface{}) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			if _, ok = m.dirty[key]; ok {
				delete(m.dirty, key)
				atomic.AddInt32(&m.length, -1)
				m.mu.Unlock()
				return
			}
		}
		m.mu.Unlock()
	}
	if ok && e.delete() {
		atomic.AddInt32(&m.length, -1)
	}
}

func (e *entry) delete() (hadValue bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == expunged {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return true
		}
	}
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the atomicMap's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *atomicMap) Range(f func(key, value interface{}) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read, _ := m.read.Load().(readOnly)
	if read.amended {
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly)
		if read.amended {
			read = readOnly{m: m.dirty}
			m.read.Store(read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

// Clear clears all current data in the map.
func (m *atomicMap) Clear() {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read, _ := m.read.Load().(readOnly)
	if read.amended {
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly)
		if read.amended {
			read = readOnly{m: m.dirty}
			m.read.Store(read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for _, e := range read.m {
		_, ok := e.load()
		if !ok {
			continue
		}
		if e.delete() {
			atomic.AddInt32(&m.length, -1)
		}
	}
}

func (m *atomicMap) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(readOnly{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

func (m *atomicMap) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnly)
	m.dirty = make(map[interface{}]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

func (e *entry) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&e.p, nil, expunged) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
	}
	return p == expunged
}

// Len returns the length of the map.
// Note:
//  the length may be inaccurate.
// @added by henrylee2cn 2017/11/17
func (m *atomicMap) Len() int {
	return int(atomic.LoadInt32(&m.length))
}

// Random returns a pair kv randomly.
// If exist=false, no kv data is exist.
// @added by henrylee2cn 2017/08/10
func (m *atomicMap) Random() (key, value interface{}, exist bool) {
	var (
		length, i int
		read      readOnly
		e         *entry
	)
	for {
		read, _ = m.read.Load().(readOnly)
		if read.amended {
			m.mu.Lock()
			read, _ = m.read.Load().(readOnly)
			if read.amended {
				read = readOnly{m: m.dirty}
				m.read.Store(read)
				m.dirty = nil
				m.misses = 0
			}
			m.mu.Unlock()
		}
		length = m.Len()
		if length <= 0 {
			return nil, nil, false
		}
		i = rand.Intn(length)
		for key, e = range read.m {
			value, exist = e.load()
			if !exist {
				continue
			}
			if i > 0 {
				i--
				continue
			}
			return
		}
	}
}
