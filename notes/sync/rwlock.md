# rwlock

## - 结构体

```
type RWMutex struct {
    w           Mutex  // held if there are pending writers
    writerSem   uint32 // semaphore for writers to wait for completing readers
    readerSem   uint32 // semaphore for readers to wait for completing writers
    readerCount int32  // number of pending readers
    readerWait  int32  // number of departing readers
}
```

这里哟一个mutex 有两个条件变量 writerSem 和 readerSem 还有两个条件变量，然后有一个reader count 还有个readerWait

```
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
    "internal/race"
    "sync/atomic"
    "unsafe"
)

// There is a modified copy of this file in runtime/rwmutex.go.
// If you make any changes here, see if you should make them there.

// A RWMutex is a reader/writer mutual exclusion lock.
// The lock can be held by an arbitrary number of readers or a single writer.
// The zero value for a RWMutex is an unlocked mutex.
//
// A RWMutex must not be copied after first use.
//
// If a goroutine holds a RWMutex for reading and another goroutine might
// call Lock, no goroutine should expect to be able to acquire a read lock
// until the initial read lock is released. In particular, this prohibits
// recursive read locking. This is to ensure that the lock eventually becomes
// available; a blocked Lock call excludes new readers from acquiring the
// lock.
type RWMutex struct {
    w           Mutex  // held if there are pending writers
    writerSem   uint32 // semaphore for writers to wait for completing readers
    readerSem   uint32 // semaphore for readers to wait for completing writers
    readerCount int32  // number of pending readers
    readerWait  int32  // number of departing readers
}

const rwmutexMaxReaders = 1 << 30
// 如果是以下的操作
// Happens-before relationships are indicated to the race detector via:
// - Unlock  -> Lock:  readerSem
// - Unlock  -> RLock: readerSem
// - RUnlock -> Lock:  writerSem
//
// The methods below temporarily disable handling of race synchronization
// events in order to provide the more precise model above to the race
// detector.
//
// For example, atomic.AddInt32 in RLock should not appear to provide
// acquire-release semantics, which would incorrectly synchronize racing
// readers, thus potentially missing races.

// RLock locks rw for reading.
//
// It should not be used for recursive read locking; a blocked Lock
// call excludes new readers from acquiring the lock. See the
// documentation on the RWMutex type.
// RLock 
func (rw *RWMutex) RLock() {
    // 如果是enable
    if race.Enabled {
        _ = rw.w.state
        race.Disable()
    }
    // 如果增加一个rw.readerCount计数
    if atomic.AddInt32(&rw.readerCount, 1) < 0 {
        // 如果写锁被人拿走
        // A writer is pending, wait for it.
        // 沉睡··
        runtime_SemacquireMutex(&rw.readerSem, false, 0)
    }
    // 
    if race.Enabled {
        race.Enable()
        race.Acquire(unsafe.Pointer(&rw.readerSem))
    }
}

// RUnlock undoes a single RLock call;
// it does not affect other simultaneous readers.
// It is a run-time error if rw is not locked for reading
// on entry to RUnlock.
// runlock
// 
func (rw *RWMutex) RUnlock() {
    if race.Enabled {
        _ = rw.w.state
        race.ReleaseMerge(unsafe.Pointer(&rw.writerSem))
        race.Disable()
    }
    // 减少一个计数，如果没有reader了 
    if r := atomic.AddInt32(&rw.readerCount, -1); r < 0 {
        // 如果减少计数
        // Outlined slow-path to allow the fast-path to be inlined
        rw.rUnlockSlow(r)
    }
    if race.Enabled {
        race.Enable()
    }
}

func (rw *RWMutex) rUnlockSlow(r int32) {
    // 如果没有reader了 如果到了最大的reder
    if r+1 == 0 || r+1 == -rwmutexMaxReaders {
        race.Enable()
        throw("sync: RUnlock of unlocked RWMutex")
    }
    // A writer is pending.
    // 减少一个等待的wait的数目 如果readwait没有了 会唤醒写锁
    if atomic.AddInt32(&rw.readerWait, -1) == 0 {
        // The last reader unblocks the writer.
        runtime_Semrelease(&rw.writerSem, false, 1)
    }
}

// Lock locks rw for writing.
// If the lock is already locked for reading or writing,
// Lock blocks until the lock is available.
// 如果mutex会等所有读锁
func (rw *RWMutex) Lock() {
    if race.Enabled {
        _ = rw.w.state
        race.Disable()
    }
    // First, resolve competition with other writers.
    rw.w.Lock()
    // Announce to readers there is a pending writer.
    // 通过减少变成负的 告诉reader 有写锁
    r := atomic.AddInt32(&rw.readerCount, -rwmutexMaxReaders) + rwmutexMaxReaders
    // Wait for active readers.
    // 如果当时readcount 不是0 且就是在很特殊情况下 （刚好计算r时候readcount都在减少）等待的数目和readcount加起来不为0
    if r != 0 && atomic.AddInt32(&rw.readerWait, r) != 0 {
        runtime_SemacquireMutex(&rw.writerSem, false, 0)
    }
    if race.Enabled {
        race.Enable()
        race.Acquire(unsafe.Pointer(&rw.readerSem))
        race.Acquire(unsafe.Pointer(&rw.writerSem))
    }
}

// Unlock unlocks rw for writing. It is a run-time error if rw is
// not locked for writing on entry to Unlock.
//
// As with Mutexes, a locked RWMutex is not associated with a particular
// goroutine. One goroutine may RLock (Lock) a RWMutex and then
// arrange for another goroutine to RUnlock (Unlock) it.
func (rw *RWMutex) Unlock() {
    if race.Enabled {
        _ = rw.w.state
        race.Release(unsafe.Pointer(&rw.readerSem))
        race.Disable()
    }

    // Announce to readers there is no active writer.
    // 告诉reader 没有写锁了
    r := atomic.AddInt32(&rw.readerCount, rwmutexMaxReaders)
    if r >= rwmutexMaxReaders {
        race.Enable()
        throw("sync: Unlock of unlocked RWMutex")
    }
    // Unblock blocked readers, if any.
    // 唤醒每一个读锁
    for i := 0; i < int(r); i++ {
        runtime_Semrelease(&rw.readerSem, false, 0)
    }
    // Allow other writers to proceed.
    rw.w.Unlock()
    if race.Enabled {
        race.Enable()
    }
}

// RLocker returns a Locker interface that implements
// the Lock and Unlock methods by calling rw.RLock and rw.RUnlock.
func (rw *RWMutex) RLocker() Locker {
    return (*rlocker)(rw)
}

type rlocker RWMutex

func (r *rlocker) Lock()   { (*RWMutex)(r).RLock() }
func (r *rlocker) Unlock() { (*RWMutex)(r).RUnlock() }
```

## 总结

- 总的来说golang 的rwlock是 以读为先的，在读锁的占用的情况下，要等所有的读锁没了才写，读锁可以一直续

- 在写锁情况下，如果有新的写请求，会先给读锁，然后unlock，写锁又等在后面了
