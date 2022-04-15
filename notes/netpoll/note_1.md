# NetPoller 

## 常用api

```
// 初始化，就调用一次
func netpollinit()
// arm 边缘触发fd通知，pd参数用于回传netpollredy当fd ready好了，返回一个error值
func netpollopen(fd uintptr, pd *pollDesc) int32
// netpollclose 关闭fd 返回一个error
func netpollclose(fd uintptr) int32
//     Disable notifications for fd. Return an errno value.
// poll 网络，如果delta<0 无限期阻塞
// delta = 0 没有阻塞
// 如果delta > 0, 阻塞 delta 纳秒 返回一个g的列表，调用netpollreadey
func netpoll(delta int64) gList
//     Poll the network. If delta < 0, block indefinitely. If delta == 0,
//     poll without blocking. If delta > 0, block for up to delta nanoseconds.
//     Return a list of goroutines built by calling netpollready.

// 唤醒network poller，假定在netpoll 中阻塞
func netpollBreak()

// 判断fd是否在poller中
func netpollIsPollDescriptor(fd uintptr) bool
//     Reports whether fd is a file descriptor used by the poller.
```

```
// pollDesc contains 2 binary semaphores, rg and wg, to park reader and writer
// goroutines respectively. The semaphore can be in the following states:
// pdReady - io readiness notification is pending;
//           a goroutine consumes the notification by changing the state to nil.
// pdWait - a goroutine prepares to park on the semaphore, but not yet parked;
//          the goroutine commits to park by changing the state to G pointer,
//          or, alternatively, concurrent io notification changes the state to pdReady,
//          or, alternatively, concurrent timeout/close changes the state to nil.
// G pointer - the goroutine is blocked on the semaphore;
//             io notification or timeout/close changes the state to pdReady or nil respectively
//             and unparks the goroutine.
// nil - none of the above.
const (
	pdReady uintptr = 1
	pdWait  uintptr = 2
)
// 两个状态码，一个是pdReady 读通知还没来， 靠改变状态nil，通知。两个信号量rg wg 用于挂起读和写协程
// pdwait 用于通过信号量挂起，但是还没有挂起
```
这里记录
暂时到这
