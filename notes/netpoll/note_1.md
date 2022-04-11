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

