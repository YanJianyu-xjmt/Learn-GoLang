# 

GMP 模型

- 模型

- Goroutine = Golang + Coroutine 使用户级别线程
  
  -     相比线程，其启动代价很小，以很小栈空间启动2kb 左右
  
  - 能够动态的伸缩栈的大小，最大可以支持到GB级别
  
  - 工作在用户态，切换成本很小
  
  - 与线程关系是n：m。可以在n个系统线程上调度多个 m 个goroutine

- 内核调度实体 kse 用户和  kernel scheduling entity 之间的对应关系
  
  - 内核级线程模型
  
  - 用户级线程模型
  
  - 两级线程模型，也称混合型线程模型

- M 直接关联KSE 

- M 是铁打的，然后 P（逻辑处理器） 有一个本地队列，最多可以存放256个G，

- 然后个本地队列

- 然后还有个schedt

## 调度

- 线程M想运行任务就要获得P，即与之关联

- 然后从本地队列（LPQ） 获取G

- 如果LPQ 没有可运行的G

- 如果全局列表也未找到可运行G时，M会尝试从全局队列 CRQ中拿一批G放到本地P

- 如果全局也没有 work steal ，从边上偷 一半过来

- 拿到可以运行的G之后，M运行G，G执行之后，M会从P中获取下一个G，不断重复

# 调度的生命周期

- 先创建最初的m0 然后每个 g 有个g0

- g0：是每个m最初的g0，主要用于调度 g

- 调度初始化-》 设置M的最大数量，2 初始化 启动m0，m 绑定P

## m0 是启动程序之后 编号为 0 的主线程、

GMP 模型 G的数量是基本没限制的

P 

M 默认最带限制10000 基本不够

## 调度过程中的阻塞

- GMP模型的阻塞可能发生的情况
  
  - I/O select 
  
  - block on syscall 
  
  - channel
  
  - 等待锁
  
  - runtime.Gosched()
  
  - goroutine 因为channel 或者network 而被阻塞的时候，（实际上golang 已经使用netpoll 实现了 goroutine 网络I/o 阻塞只会阻塞g，m不会被阻塞），被阻塞的g会被放到某个队列里面，比如channel 的waiting list， g的状态 从 _Gruning 变成 —Gwaiting 等待，m 会跳过改g 获取并执行下一个g， 如果此时没有runable 的G供M运行，那么m和p解绑 ，进入sleep  。当这个G被下一个G 唤醒的时候，标记runable
  
  - 系统调用阻塞
    
    - 当G 被系统调用阻塞 —Gsyscall 状态，m 也会处于 block on syscall 
    
    - 
  
  - g 的状态
    
    - Gidle 刚分配 没有初始化
    
    - grunable 已经在对立中，还没有执行用户代码
    
    - gruning 不在运行队列里，可以执行用户代码 此时分配了m 和p
    
    - gsyscall  正在执行系统调用，此时分配了M
    
    - gwaiting 在运行时候被组织，没有执行用户代码，也不在运行队列中，此时它正处在某处阻塞等待中，比如channel
    
    - gmoribund—unsed 尚未使用，但是在gdb 进行了硬编码
    
    - gdead 尚未使用 
    
    - genqueue—unused 尚未使用
    
    - gccopystack 正在复制堆栈，并没有执行用户代码，也不再运行队列中
  
  - g的机构
    
    -  stack 就是g 自己的栈
    
    - m *m 隶属于哪个栈
    
    - schedt 保存了g的现场，gorutine 切换时通过它来恢复
    
    - atomicstatus G的运行状态
    
    - goid  id
    
    - schedlink  guintptr 下一个g g链表
    
    - preempt 抢占的标记
    
    - lockedm 锁定的m g 中断恢复指定的m执行
    
    - gopc 创建该goroutine的指令地址
    
    - startpc goroutine 函数的指令地址
  
  - m的结构
    
    - g0 每个m都有一个g
    
    - curg *g 当前的g
    
    - p puintptr 隶属于哪个p
    
    - nextp 当m被唤醒的时候，首先拥有整个p
    
    - id int64
    
    - spinning 是否处于自旋
    
    - park onte 
    
    - alllink *m on allm
    
    - schedlink muintptr 下一个m m链表
    
    - mache *mache 内存分配
    
    - lockedg guintptr 和G的lockedm 对应
    
    - freelink *m  on sched.FREEm 在sched free这里面时候
  
  - p 的内部结构
    
    - id 
    
    - status p的状态
    
    - link puintptr 下一个p p链表
    
    - m muiptr 拥有这个ep的m
    
    - mcache *mcahe
    
    - p本地 runable 状态G队列 无锁访问
      
      - runqhead uint32
      
      - runqtail uint32
      
      - runq 【256】guintptr
    
    - runnext  guintptr 一个比runq 优先级更高的runable G 
    
    - GfREE 状态dead 的链表、在获取G时候 从这里面获取 
      
      - dglist
      
      - n int32
    
    - gcBgmarkWorker guintptr
    
    - gcw gcWork 
  
  - P 的几种状态
    
    - Pidle 刚被分配，还没进行初始化
    
    - Pruning m与P 绑定嗲用acquirep 时候，p的状态改变为 pruning
    
    - psyscall 正在执行系统调用
    
    - pgcstop 暂挺运行 此时系统正在进行gc 知道gc结束后才会转变到下一个状态阶段
    
    - pdead 废弃 不在使用
  
  - 调度器内部结构
    
    - lock mutex 
    
    - midle  muintptr 空闲m链表
    
    - nmidle int32 空闲m数量
    
    - nmidlelocked  int32被锁住的m的数量
    
    - mnext int6 已经创建m的总数
    
    - maxmcount 允许最大的m的属灵
    
    - nmsys 不计入思索的m的数量
    
    - nmfreed 累计释放m的数量
    
    - pidle 空闲p链表
    
    - npidle 空闲p数量太
    
    - runq gQueue 全局runnable 队列
    
    - runqsize 全部runnable 的数量
    
    - gFree 
      
      - lock mutex
      
      - stack Gs with stacks
      
      - noStack 
      
      - n
    
    - freem *m 

- goDEBUG trace方式
  
  - GODEBUG 变量可以控制运行时 调参变量  参数用逗号分隔。 name =val 
  
  - 观察GMP 用下面两个参数
    
    - schedtrace 设值schedtrace = X 参数可以  使得运行X毫秒 输出一行调度器的宅摘要信息 到标准err的输出中
    
    - scheddetail 色湖之 schedtrace = X 和 scheddetail = 1 可以使运行在X 毫秒输出一次详细的多行信息，包含调度 处理器 os线程  和 gorotine 的状态
    
    - package main（）
      
      比如：SCHED 0ms： gomaxprocs=1 idleprocs = 1 thread=4 spiningthreads=0 idlethread=1 runqueue = 0 【0】
    
    - gomaxproces p的数量 等于当前cpu 核心数
    
    - idleprocs 空闲的p的数量 与 gomaxprocs的差值， 即运行中p的属灵

- go tool trace
  
  - 怎么用呢  先create 文件
  
  - 然后 trace.Start(f)  defer trace.Stop F.CLOSE()

GMP 高效的策略

- 内存分配站在mache ，在P上绑定，每个P里面

- M是可以复用的 不需要反复创建与销毁，当没有Goroutine 处于自旋

- Work Stealing 和 Hand Off 策略保证M的高效利用

- 内存分配状态处于mcahe P ，G可以跨M调度，不再存在M调度局部性差的问题

- M 从关联的P获取G 不需要使用锁  是lock free的。这个很好理解，因为M和P 只会一个一个结合

## GMP模型为什么要有P

- GM模型  存在全局单一mutex 降低效率

- Goroutine 传递的问题，G和工作线程交接麻烦

- 每个m

- 每个M都要做内存缓存

- 频繁的线程阻塞/解阻塞  在syscalls 的情况下，线程经常被阻塞和解阻塞  这增加了很多性能开销

## 为什么要P

- M的数量多余P ，在GO 中，M的数量默认100000，P默认只有CPu的核心数，由于M的属性，如果存在了系统阻塞调用，阻塞了M 又不够用的情况下，M不断增加

## 源码

- 

```
https://blog.csdn.net/qq_42956653/article/details/121198941
```

可执文件

_rt0_amd64_linux 

osinit

new main goroutine

runtime.main

call main.main

调用的方法 显示 new main  goroutine  newproc

mstart ---》 schedule

沟通难过 关键字go 创建协程，然后编译器转化为new proc 然后main 

[GolangGMP模型 GMP(二):goroutine的创建，运行与恢复_清风-CSDN博客](https://blog.csdn.net/qq_42956653/article/details/121213577)

[深入理解GMP模型_清风-CSDN博客_gmp 模型](https://blog.csdn.net/qq_42956653/article/details/121234816)

深入GMP 模型

- m0 M0 是启动成迅速之后 编号为0 主线程   这个M对应的实例会在全局变量runtime.m0中，M0负责执⾏初始化操作和启动第⼀个G， 在之后M0就和其他的M⼀样了。

- g0: G0是每次启动⼀个M都会第⼀个创建的gourtine，G0仅⽤于负责调度的G，G0不指向任何可执⾏的函数,每个M都会有⼀个⾃⼰的G0。在调度或系统调⽤时会使⽤G0的栈空间, 全局变量的G0是M0的G0。

- allgs 记录所有的G

- allm 记录所有的M

- allp 记录所有P

- sched sched是调度器，这里记录所有空闲的m，空闲的P 全局队列 runq

## newproc 创建协程G

- go关键字创建协程，会被编译器转换为newproc函数调用。
  
  - newproc函数这里主要做的就是切换到g0栈去调用newproc1函数
    newproc1创建一个新G
  
  - 调用runqput把这个G放到P的本地队列中，如满则放全局队列
    如果当前有空闲的p，而且没有处于spinning状态(线程自旋)的M，也就是说所有M都在忙，同时主协程以及开始执行了，那么就调用wakep函数，启动一个m并把它置为spinning状态。spinning状态的M启动后，忙不迭的执行调度循环寻找任务，从本地runq，到全局runq，再到其他p的runq，只为找到一个待执行的G。

## runtime。gopark 挂起G

timer channel io 等都会调用gopark 挂起G 让出cpu 时间片 （主动让出）

- 态从_GRUNING 变成gwaiting

- 调用mcall（park—m） 切换到g0  执行parik_M 它主要负责保存当前协程的执行现场

- park—m 会根据g0找到当前的m，把m。curg 设置为nil

- 调用schedule（） 寻找下一个待执行的G

## runtime。goready 唤醒 g

- 切换到g0栈 并执行 runtime。ready

- 把_GWAITING 修改 —grunable

- 放到当前P的本地队列，如果犯了 放到全局队列

- 同协程创建时候 接下来会检查 是否又空闲的P，并没有spininig 状态的M 是的话也会wakep 启动新的M

## sysmon 监控线程

线程刚开始，M0 切换到main goroutine 执行入口是runtime。main 

- 监控main goroutine 创建，监控线程独立gmp 之外，会重复执行一些列任务，只不过会视情况调整自己的休眠时间

- 监控线程检测接下来又的timer 执行，不仅会按需要调整休眠时间，还会再恐怖出M创建新的工作线程，保证timer 顺利执行

- 获取就绪的IO时间需要主动轮询，所以为了降低IO延迟，时不时轮询也就是执行netpoll

- 强制执行GC

## Schedult 调度

- schedule 这里会给M 找到一个等待执行的G，首先要确定当前的M是否和当前G来运行，如果绑定了。档期那M不能执行G，所以阻塞当前的M，等调度G时候，自动把M唤醒

- 如果没有绑定 先康康GC是不是再等待执行

- 当前变量schedt 这里有个gcwaiting 标识，如果gc再等待执行，那么执行gc 再回来继续执行调度程序

- 接下来还会检查一些有没有要执行的timer，调度程序会有一定几率去全局runq 获取一部分G到本地runq中

[深入理解GMP模型_清风-CSDN博客_gmp 模型](https://blog.csdn.net/qq_42956653/article/details/121234816)

## execte

- 如果没有绑定的M 调用excute 函数再当前M上执行这个M 伤者G    exectue 函数这里建立当前M和这个G的关联关系，并把G的状态从 grunable 修改为 gruning 如果不继承上一个执行中协程的中间。九八P这里的调度计数+1 最后调用gogo函数，从 g。sched从这里恢复协程栈指针，指令指针等，接着继续协程的执行
