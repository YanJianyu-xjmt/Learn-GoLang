# map 笔记

## 介绍

- sync map 适用的地方，大部分还得mutex 加普通的map
  
  - 一是只有一个g 写 写少，读多的情况
  
  - 二是不同的g写的 条目不一样不干扰

- sync map 是[INTERFALCE{}] INTERFACE{}的

## 原理

- 有一个 read map 只读map 

- 计一个missed 计数 计在read map中找不到的次数

- 有一个dirty map 

- 有一个 read.amended 记录是否dirty map 中有和read map 不一样

- 有一个mutex



## 读过程

- 先在 read读如果有，直接返回

- 上锁

- 在read查一次，如果有返回

- 如果read 没有，amended为true 表明dirty 可能有查了 missed + 1，返回

- 如果missed 比len（dirty）大，dirty 变成新的read ，dirty 为nil amended为false，missed 0

- 这里问题是啥呢，就是read中完全为nil了



## 写过程

- 先read map查 如果有 尝试改，如果设置为删除会失败，如果成功返回

- 上锁

- 先read map查 如果有 尝试改，如果设置为删除会失败，如果成功返回

- 如果read map里面有 且设置为删除。在dirty map 写入

- 如果 M.Dirty 中直接改

- 如果m。dirty 没有 写入一个 设置amended为ture



```
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
	"sync/atomic"
	"unsafe"
)

// Map is like a Go map[interface{}]interface{} but is safe for concurrent use
// map 是像一种map【interface{}】interface{} 但是对于 并发是安全的是 

// by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
//
// The Map type is specialized. Most code should use a plain Go map instead,
// 对于Map type是特定的。大多数时候用 普通go map 并结合locking  
// with separate locking or coordination, for better type safety and to make it
// easier to maintain other invariants along with the map content.
//
// The Map type is optimized for two common use cases: (1) when the entry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a Map may significantly reduce lock
// contention compared to a Go map paired with a separate Mutex or RWMutex.
// 这个map 对于两种通用case 有优化 1）：当一条记录 指定的key 只写入一次 并读多次，作为缓存并只增长的时候，
// （2）当很多g读，并且写不同的条目
//
// The zero Map is empty and ready for use. A Map must not be copied after first use.
// 一个zero Map 是空的能源用。一个map在初次使用之后不能被拷贝
type Map struct {
	mu Mutex

	// read contains the portion of the map's contents that are safe for
	// concurrent access (with or without mu held).
	// read 包含portion of map的内容 对于并发进入是安全的
	// The read field itself is always safe to load, but must only be stored with
	// mu held.
	// 对于读取字段 本身是安全的load 但是一般mu持有的情况下 会被写入
	//
	// 记录在read 中可能会被 在灭有mu的情况下被并发更新  但是更新一个 之前（expunged）删除
	// 条目需要持有锁时被拷贝到脏map中
	// Entries stored in read may be updated concurrently without mu, but updating
	// a previously-expunged entry requires that the entry be copied to the dirty
	// map and unexpunged with mu held.
	read atomic.Value // readOnly

	// dirty contains the portion of the map's contents that require mu to be
	// dirty 持有保存着map 需要持有mu的部分。为了保证 dirty map能被促进的读map 快一点，包含很多没有删除记录在读map
	// held. To ensure that the dirty map can be promoted to the read map quickly,
	// 很多包含没有被删除的记录会被记录在read map 中
	// it also includes all of the non-expunged entries in the read map.
	//
	// 被删除的记录 不会被记录在dirty map中。一个被删除的记录必须被删除并加到diryt mpa
	// Expunged entries are not stored in the dirty map. An expunged entry in the
	// clean map must be unexpunged and added to the dirty map before a new value
	// can be stored to it.
	//
	// 如果dirty map 是nil，下一个写的会make map 
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
}

// readOnly is an immutable struct stored atomically in the Map.read field.
// readoly 是一个不变的结构体，存在map。read 字段中
type readOnly struct {
	m       map[interface{}]*entry
	amended bool // true if the dirty map contains some key not in m. 
	// true 如果是dirty map 包含一些key 在m中没有
}

// expunged is an arbitrary pointer that marks entries which have been deleted
// from the dirty map.
var expunged = unsafe.Pointer(new(interface{}))

// An entry is a slot in the map corresponding to a particular key.
// 一条entry 是一个槽对应一个特殊的key
type entry struct {
	// p points to the interface{} value stored for the entry.
	//
	// If p == nil, the entry has been deleted and m.dirty == nil.
	// 入宫p==nil，这条记录已经被删除 并且m。dirty == nil如果p == 删除 记录已经被删除 m.diryt != nil 
	// 并且记录在m 。dirty 中迷失
	// If p == expunged, the entry has been deleted, m.dirty != nil, and the entry
	// is missing from m.dirty.
	//
	// 除此之外，记录是合法的并且记录在m。read。m【key】 并且 如果 M.DIRTY ！=nil 在m。dirty 【key】
	// Otherwise, the entry is valid and recorded in m.read.m[key] and, if m.dirty
	// != nil, in m.dirty[key].
	//
	// 一条记录能被删除 且被原子替代，当一个M.DIRTY 会被是m。dirty 是下个打开的，会自动替代nil 删除并离开m。dirty【key】
	// An entry can be deleted by atomic replacement with nil: when m.dirty is
	// next created, it will atomically replace nil with expunged and leave
	// m.dirty[key] unset.
	//
	// 一条记录会和联系到
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
// 导入返回值存储在map中 或者是nil 如果没有值暂时
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	read, _ := m.read.Load().(readOnly)



	e, ok := read.m[key]
	
	// 如果没找到，且标记dirty map 中有read map中不存在的值
	if !ok && read.amended {
		m.mu.Lock() 
		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// 避免在上锁之前这段孔隙改变
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		// 如果没找到，
		if !ok && read.amended {
			// 在dirty 中找
			e, ok = m.dirty[key]
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			// 估计就是增加一个miss
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.load()
}

// e.load就是原子操作做
func (e *entry) load() (value interface{}, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expunged {
		return nil, false
	}
	return *(*interface{})(p), true
}

// Store sets the value for a key.
// 存储key value
func (m *Map) Store(key, value interface{}) {

	read, _ := m.read.Load().(readOnly)
	// 如果read 有这个key 那么store
	// 因为是找到entry 指针，entry改内容，所以是原子的没有问题
	// TRYSTORE 如果之前没有标记被删除
	if e, ok := read.m[key]; ok && e.tryStore(&value) {
		return
	}

	// 持有锁
	m.mu.Lock()
	// 
	read, _ = m.read.Load().(readOnly)
	// 如果现在在read 中
	if e, ok := read.m[key]; ok {
		// 如果之前被删除
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			// 如果之前记录被删除，显示non -nil dirty map 并且这个记录没有在里面
			m.dirty[key] = e
		}
		// 如果是这一段孔隙被写入的
		e.storeLocked(&value)
		// 如果在dirty 中
	} else if e, ok := m.dirty[key]; ok {
		// 改一下
		e.storeLocked(&value)
	} else {
		// 如果dirty 中没有和read中不一样的
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
	}
	m.mu.Unlock()
}

// tryStore stores a value if the entry has not been expunged.
//
// 如果之前被删除 了。返回false
// If the entry is expunged, tryStore returns false and leaves the entry
// unchanged.
func (e *entry) tryStore(i *interface{}) bool {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == expunged {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return true
		}
	}
}

// unexpungeLocked ensures that the entry is not marked as expunged.
// unexpungeLocked 确认这条记录 不被标记位 被删除
//
// 如果这条记录已经被删除，这条会被加到dirty map 在mu被释放之前
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
func (m *Map) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool) {
	// Avoid locking if it's a clean hit.
	read, _ := m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	m.mu.Lock()
	read, _ = m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			m.dirty[key] = e
		}
		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := m.dirty[key]; ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		m.missLocked()
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			// 我们添加头条新key 到dirty map
			// 确认被释放空间且 确认read map是完整的
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
		actual, loaded = value, false
	}
	m.mu.Unlock()

	return actual, loaded
}

// tryLoadOrStore atomically loads or stores a value if the entry is not
// expunged.
// 尝试
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

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
// LoadAndDlete 删除一个
func (m *Map) LoadAndDelete(key interface{}) (value interface{}, loaded bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	// 如果没有 且dirty 中可能有
	if !ok && read.amended {
		m.mu.Lock()

		read, _ = m.read.Load().(readOnly)
		
		e, ok = read.m[key]
		// 
		if !ok && read.amended {
			e, ok = m.dirty[key]
			delete(m.dirty, key) // 在dirty 中删除这key
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			// M.MISSEDlOCKED 增加一个 并判断是否要替换
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if ok {
		// 记录删除
		return e.delete()
	}
	return nil, false
}


// Delete deletes the value for a key.
func (m *Map) Delete(key interface{}) {
	m.LoadAndDelete(key)
}

func (e *entry) delete() (value interface{}, ok bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == expunged {
			return nil, false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return *(*interface{})(p), true
		}
	}
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently, Range may reflect any mapping for that key
// from any point during the Range call.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *Map) Range(f func(key, value interface{}) bool) {
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

// 把miss 计数加一个
func (m *Map) missLocked() {
	m.misses++
	// 如果miss 小于dirty 数目
	if m.misses < len(m.dirty) {
		return
	}
	// 把dirty 置成read ，dirty变成nil
	m.read.Store(readOnly{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

// 
func (m *Map) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnly)
	// 如果
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

```









