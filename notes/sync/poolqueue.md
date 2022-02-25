# poolqueue 笔记

```
// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

import (
    "sync/atomic"
    "unsafe"
)

// poolDequeue is a lock-free fixed-size single-producer,
// poolDequeue 是一种无锁 但是容量不变的单生产者，多消费者的队列
// multi-consumer queue. The single producer can both push and pop
// 单一的生产者 能够从头节点 pop 和push 但是消费者只能在后面pop
// from the head, and consumers can pop from the tail.
//
// 还有一个新的特征，能够消灭没有使用的槽 来避免保持不必要的对象
// It has the added feature that it nils out unused slots to avoid
// unnecessary retention of objects. This is important for sync.Pool,
// 这对于sync.Pool来说非常重要，但是在一般的文献中不是常规考虑。
// but not typically a property considered in the literature.

type poolDequeue struct {
    // headTail 打包了32字节 头索引 和 32 字节 尾节点。
    // headTail packs together a 32-bit head index and a 32-bit
    // tail index. Both are indexes into vals modulo len(vals)-1.
    //
    // tail 在队列最老的对象
    // tail = index of oldest data in queue
    // head 下一个被填充的槽
    // head = index of next slot to fill
    //
    // 槽在范围内都是被消费者拥有
    // Slots in the range [tail, head) are owned by consumers.
    // 一个消费者 持续拥有一个槽直到不使用  
    // A consumer continues to own a slot outside this range until
    // it nils the slot, at which point ownership passes to the
    // producer.
    //
    // The head index is stored in the most-significant bits so
    // that we can atomically add to it and the overflow is
    // harmless.
    headTail uint64

    // vals is a ring buffer of interface{} values stored in this
    // dequeue. The size of this must be a power of 2.
    //
    // vals[i].typ is nil if the slot is empty and non-nil
    // otherwise. A slot is still in use until *both* the tail
    // index has moved beyond it and typ has been set to nil. This
    // is set to nil atomically by the consumer and read
    // atomically by the producer.
    vals []eface
}

type eface struct {
    typ, val unsafe.Pointer
}

const dequeueBits = 32

// dequeueLimit is the maximum size of a poolDequeue.
// dequeueLimit 是最大的size
//
// 
// This must be at most (1<<dequeueBits)/2 because detecting fullness
// depends on wrapping around the ring buffer without wrapping around
// the index. We divide by 4 so this fits in an int on 32-bit.
const dequeueLimit = (1 << dequeueBits) / 4

// dequeneNil 
// dequeueNil is used in poolDequeue to represent interface{}(nil).
// Since we use nil to represent empty slots, we need a sentinel value
// to represent nil.
type dequeueNil *struct{}

// unpack 
func (d *poolDequeue) unpack(ptrs uint64) (head, tail uint32) {
    // 感觉就是前半部分存头 后32存尾
    const mask = 1<<dequeueBits - 1
    head = uint32((ptrs >> dequeueBits) & mask)
    tail = uint32(ptrs & mask)
    return
}

func (d *poolDequeue) pack(head, tail uint32) uint64 {
    // 
    const mask = 1<<dequeueBits - 1
    return (uint64(head) << dequeueBits) |
        uint64(tail&mask)
}

// pushHead adds val at the head of the queue. It returns false if the
// queue is full. It must only be called by a single producer.
// pushHead 添加一个新的值在队列前面。如果满了 会返回false。必须单生产者使用
func (d *poolDequeue) pushHead(val interface{}) bool {
    // 获取头尾节点
    ptrs := atomic.LoadUint64(&d.headTail)
    head, tail := d.unpack(ptrs)
    // 如果已经满了
    if (tail+uint32(len(d.vals)))&(1<<dequeueBits-1) == head {
        // Queue is full.
        return false
    }
    // 就炫技 就是新的槽
    slot := &d.vals[head&uint32(len(d.vals)-1)]

    // Check if the head slot has been released by popTail.
    // 看是否已经被释放了
    typ := atomic.LoadPointer(&slot.typ)
    if typ != nil {
        // Another goroutine is still cleaning up the tail, so
        // the queue is actually still full.
        // 如果另外一个还在尾巴上清扫 所以queue还是满的
        return false
    }

    // The head slot is free, so we own it
    if val == nil {
        val = dequeueNil(nil)
    }
    *(*interface{})(unsafe.Pointer(slot)) = val

    // Increment head. This passes ownership of slot to popTail
    // and acts as a store barrier for writing the slot.
    // 增加head，交流slot的所有权，
    atomic.AddUint64(&d.headTail, 1<<dequeueBits)
    return true
}

// popHead removes and returns the element at the head of the queue.
// It returns false if the queue is empty. It must only be called by a
// single producer.
// popHead 移除 并返回元素 头 
// 如果是空的 必须要 单生产者调用
func (d *poolDequeue) popHead() (interface{}, bool) {
    var slot *eface
    for {
        ptrs := atomic.LoadUint64(&d.headTail)
        head, tail := d.unpack(ptrs)
        if tail == head {
            // Queue is empty.
            return nil, false
        }

        // Confirm tail and decrement head. We do this before
        // reading the value to take back ownership of this
        // slot.
        // 确认尾节点并减少头，我们在读 值并移除所有权之前。
        head--
        ptrs2 := d.pack(head, tail)
        // 如果成功写了 
        if atomic.CompareAndSwapUint64(&d.headTail, ptrs, ptrs2) {
            // We successfully took back slot.
            slot = &d.vals[head&uint32(len(d.vals)-1)]
            break
        }
    }

    val := *(*interface{})(unsafe.Pointer(slot))
    if val == dequeueNil(nil) {
        val = nil
    }
    // Zero the slot. Unlike popTail, this isn't racing with
    // pushHead, so we don't need to be careful here.
    // SLOT 置0 因为这个是只有一个人用， 所以无所谓
    *slot = eface{}
    return val, true
}

// popTail removes and returns the element at the tail of the queue.
// It returns false if the queue is empty. It may be called by any
// number of consumers.
func (d *poolDequeue) popTail() (interface{}, bool) {
    var slot *eface
    for {
        ptrs := atomic.LoadUint64(&d.headTail)
        head, tail := d.unpack(ptrs)
        // 空队列
        if tail == head {
            // Queue is empty.
            return nil, false
        }

        // Confirm head and tail (for our speculative check
        // above) and increment tail. If this succeeds, then
        // we own the slot at tail.
        ptrs2 := d.pack(head, tail+1)
        // 应该是不相等然后写入
        if atomic.CompareAndSwapUint64(&d.headTail, ptrs, ptrs2) {
            // Success.
            slot = &d.vals[tail&uint32(len(d.vals)-1)]
            break
        }
    }

    // We now own slot.
    val := *(*interface{})(unsafe.Pointer(slot))
    if val == dequeueNil(nil) {
        val = nil
    }

    // Tell pushHead that we're done with this slot. Zeroing the
    // slot is also important so we don't leave behind references
    // that could keep this object live longer than necessary.
    //
    // We write to val first and then publish that we're done with
    // this slot by atomically writing to typ.
    slot.val = nil
    atomic.StorePointer(&slot.typ, nil)
    // At this point pushHead owns the slot.

    return val, true
}

// poolChain is a dynamically-sized version of poolDequeue.
// poolChain 是自动自适应尺寸 版本 的poolDequeue
//
// This is implemented as a doubly-linked list queue of poolDequeues
// where each dequeue is double the size of the previous one. Once a
// dequeue fills up, this allocates a new one and only ever pushes to
// the latest dequeue. Pops happen from the other end of the list and
// once a dequeue is exhausted, it gets removed from the list.
// 是一个双端链表 一旦队列满了 会申请一个新的 并加到最后，
type poolChain struct {
    // head is the poolDequeue to push to. This is only accessed
    // by the producer, so doesn't need to be synchronized.
    head *poolChainElt

    // tail is the poolDequeue to popTail from. This is accessed
    // by consumers, so reads and writes must be atomic.
    tail *poolChainElt
}

type poolChainElt struct {
    poolDequeue

    // next and prev link to the adjacent poolChainElts in this
    // poolChain.
    //
    // next is written atomically by the producer and read
    // atomically by the consumer. It only transitions from nil to
    // non-nil.
    //
    // prev is written atomically by the consumer and read
    // atomically by the producer. It only transitions from
    // non-nil to nil.
    next, prev *poolChainElt
}

func storePoolChainElt(pp **poolChainElt, v *poolChainElt) {
    atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(pp)), unsafe.Pointer(v))
}

func loadPoolChainElt(pp **poolChainElt) *poolChainElt {
    return (*poolChainElt)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(pp))))
}

// pushHeadn
func (c *poolChain) pushHead(val interface{}) {
    d := c.head
    if d == nil {
        // Initialize the chain.
        const initSize = 8 // Must be a power of 2
        d = new(poolChainElt)
        d.vals = make([]eface, initSize)
        c.head = d
        storePoolChainElt(&c.tail, d)
    }

    if d.pushHead(val) {
        return
    }

    // The current dequeue is full. Allocate a new one of twice
    // the size.
    newSize := len(d.vals) * 2
    if newSize >= dequeueLimit {
        // Can't make it any bigger.
        newSize = dequeueLimit
    }

    d2 := &poolChainElt{prev: d}
    d2.vals = make([]eface, newSize)
    c.head = d2
    storePoolChainElt(&d.next, d2)
    d2.pushHead(val)
}

func (c *poolChain) popHead() (interface{}, bool) {
    d := c.head
    for d != nil {
        if val, ok := d.popHead(); ok {
            return val, ok
        }
        // There may still be unconsumed elements in the
        // previous dequeue, so try backing up.
        d = loadPoolChainElt(&d.prev)
    }
    return nil, false
}

func (c *poolChain) popTail() (interface{}, bool) {
    d := loadPoolChainElt(&c.tail)
    if d == nil {
        return nil, false
    }

    for {
        // It's important that we load the next pointer
        // *before* popping the tail. In general, d may be
        // transiently empty, but if next is non-nil before
        // the pop and the pop fails, then d is permanently
        // empty, which is the only condition under which it's
        // safe to drop d from the chain.
        d2 := loadPoolChainElt(&d.next)

        if val, ok := d.popTail(); ok {
            return val, ok
        }

        if d2 == nil {
            // This is the only dequeue. It's empty right
            // now, but could be pushed to in the future.
            return nil, false
        }

        // The tail of the chain has been drained, so move on
        // to the next dequeue. Try to drop it from the chain
        // so the next pop doesn't have to look at the empty
        // dequeue again.
        if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&c.tail)), unsafe.Pointer(d), unsafe.Pointer(d2)) {
            // We won the race. Clear the prev pointer so
            // the garbage collector can collect the empty
            // dequeue and so popHead doesn't back up
            // further than necessary.
            storePoolChainElt(&d2.prev, nil)
        }
        d = d2
    }
}
```

实际上就是一种无锁队列  一生产者 多消费者  生产者可以在头随便push pop 消费者只能尾巴上pop

