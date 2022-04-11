# Go NetPoller

## Blocking

在GO 所有的I/O 都是阻塞的，go 生态围绕着对抗阻塞并 通过channle 和 协程解决并发，而不是回调和future。

一个例子就是 net/http 包，当介绍了链接之后，会创建一个新的协程来处理链接上的事情。 感觉就是 connection-per-goroutine 

这样可以非常直接了当的处理。



go会使用os提供的异步接口解决问题，而不是使用阻塞的协程；因为过多的协程会消耗很多。

netpoller 转化异步io 为同步io  ，netpoller 使用。全都放在自己的线程里面。



当用Go接受了一个链接，文件描述符会设置为非阻塞模式，会描述为文件描述符没准备好，会返回错误码并说so。

{无论何时一个协程尝试读或者写一个链接，网络码会做这个操作直到收到一个error、最后调用netpoller，告诉协程已经好了去做io一次。这个协程会被调度出这个线程，并且另一个协程会替代。}就是会协程检查，会返回错误直到好了。



当netpoller 从os收到通知，能够展现io的文件描述符，。能够通过内部数据结构，看到多少协程被阻塞  并通知他们。  协程能够重复尝试io 操作，因为会被阻塞直到成功。



听起来和epoll差不多，因为就是。但是不是查找一个函数指针和结构体包含状态的变量，但是可以查找调度的协程。能够从各种状态中释放，重新加测你是否接受到了足够的数据，并判断函数指针 就像传统的 unix 网络io。

## 文档

https://morsmachine.dk/netpoller



## NetPoll 在Linux的具体实现

- src/runtime/netpoll_epoll.go
- src/runtime/netpoll.go