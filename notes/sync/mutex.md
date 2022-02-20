# 互斥锁
# mutex 解读

## 结构体

```
type Mutex struct {
    state int32
    sema  uint32
}
```

结构体如下 有两个32字节的变量 



- mutex是公平锁

- mutex 有两个模式 normal starvation

- 在normal 状态fifo  

- 如果是饥饿模式 直接放到队列前面

- sema 是条件变量

- state 0（可以用） 1（被锁）2-31等待队列计数

- 



Lock 比赛 记录

```
ffunc throw(string) // provided by runtime

// A Mutex is a mutual exclusion lock.
// The zero value for a Mutex is an unlocked mutex.
//
// A Mutex must not be copied after first use.
type Mutex struct {
    state int32
    sema  uint32
}

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
    Lock()
    Unlock()
}

const (
    mutexLocked = 1 << iota // mutex is locked
    mutexWoken
    mutexStarving
    mutexWaiterShift = iota

    // 公平锁
    //
    // Mutex can be in 2 modes of operations: normal and starvation.
    // mutex 有饥饿和normal模式
    // 在normal 模式 等待的的waters会在fifo中队列
    // 但是一个新woke的g 不会拥有muterx 而是要计算新的g
    // 新到达到的g有一个好处 就是现在就是跑在cpu上所以可以有很多 所以新woke的g可能会losing
    // 在一个case 下，队头的g可能等待超过1ms 这时候就会变成starving 模式
    //
    // 在饥饿模式直接放到队列里面，直接调用放到队列后面，就算是刚好锁可以用也不会给

    // 如果一个waiter收到所有权  他是队列最后一个 或者 小于1ms 会变成normal模式
    //
    // 常规模式性能更好 但是饥饿模式更加看重长尾latency

    starvationThresholdNs = 1e6
)

// Lock locks m.
// If the lock is already in use, the calling goroutine
// blocks until the mutex is available.
func (m *Mutex) Lock() {
    // Fast path: grab unlocked mutex.
    if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
        if race.Enabled {
            race.Acquire(unsafe.Pointer(m))
        }
        return
    }
    // Slow path (outlined so that the fast path can be inlined)
    m.lockSlow()
}

func (m *Mutex) lockSlow() {
    var waitStartTime int64
    starving := false
    awoke := false
    iter := 0
    old := m.state
    for {
        // 如果不在饥饿模式（饥饿模式直接）
        if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
            // 激活自旋语义 设置mutexWoken flag标记 用于不唤醒阻塞的g
            if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
                atom
