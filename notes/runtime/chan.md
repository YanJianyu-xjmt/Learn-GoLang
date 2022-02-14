# Chan.go

## 结构体

- ```
  type hchan struct {
  	// 所有的data数目
  	qcount   uint           // total data in the queue
  	dataqsiz uint           // size of the circular queue
  	buf      unsafe.Pointer // points to an array of dataqsiz elements
  	elemsize uint16
  	closed   uint32
  	elemtype *_type // element type
  	sendx    uint   // send index
  	recvx    uint   // receive index
  	recvq    waitq  // list of recv waiters
  	sendq    waitq  // list of send waiters
  	lock     mutex
  }
  ```

- 总结
  
  - 包含缓冲区 和 等待/发送 队列  里面挂着 协程
  
  - 在make 的时候会是释放buf 如果 是size 为0或者 不包含指针 直接分配  如果包含 会根据新的元素重新搞一个缓冲区

- 发送/接收过程
  
  - 如果等待队列里面有等待的g 直接发送
  
  - 如果没有放在缓冲区里面
  
  - 如果缓冲区还满了  会block住 把自己挂到等待队列里面

- channel 不是无锁的奥


