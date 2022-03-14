# Context.md

## context interface

context 是个interface 有这几个就行

```
type Context interface{
    Deadline (dealline time.Time, ok bool)
     Done () <-chan struct{}
    Err() error
    Value(key interface{}) interface{}    

}
```

使用的方法如下：

流式 DoSOmeting 产生 并送到 out这个channel里面，再ctx。Done 关闭的时候结束

```
func Stream(ctx context.Context, out chan<- Value) error {
      for {
              v, err := DoSomething(ctx)
              if err != nil {
                  return err
              }
              select {
              case <-ctx.Done():
                  return ctx.Err()
              case out <- v:
              }
          }
      }
```

withCancel 会这样，直接用 CancelFunc 直接阻塞，然后



## 使用方法

再内部调用维护一个调用树   通过调用树 再传递参数超时时候  或者退出通知，还能传递元数据 



首先最简单的emtyCtx

```
/ An emptyCtx is never canceled, has no values, and has no deadline. It is not
// struct{}, since vars of this type must have distinct addresses.
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}

func (e *emptyCtx) String() string {
	switch e {
	case background:
		return "context.Background"
	case todo:
		return "context.TODO"
	}
	return "unknown empty Context"
}

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

// Background returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline. It is typically used by the main function,
// initialization, and tests, and as the top-level Context for incoming
// requests.
func Background() Context {
	return background
}

// TODO returns a non-nil, empty Context. Code should use context.TODO when
// it's unclear which Context to use or it is not yet available (because the
// surrounding function has not yet been extended to accept a Context
// parameter).

func TODO() Context {
	return todo
}

```

TODO emptyCtx 返回的都是nil 说明是没啥用的nil 主要 用于做 contex.BACKGROUND 又Background 和TODO 凉饿函数，感觉就是构建



  cancelCtx context

```

type cancelCtx struct {
	Context

	mu       sync.Mutex            // protects following fields
	done     chan struct{}         // created lazily, closed by first cancel call
	children map[canceler]struct{} // set to nil by the first cancel call
	err      error                 // set to non-nil by the first cancel call
}

func (c *cancelCtx) Done() <-chan struct{} {
	c.mu.Lock()
	if c.done == nil {
		c.done = make(chan struct{})
	}
	d := c.done
	c.mu.Unlock()
	return d
}

func (c *cancelCtx) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

func (c *cancelCtx) String() string {
	return fmt.Sprintf("%v.WithCancel", c.Context)
}

// cancel closes c.done, cancels each of c's children, and, if
// removeFromParent is true, removes c from its parent's children.
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err
	if c.done == nil {
		c.done = closedchan
	} else {
		close(c.done)
	}
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		removeChild(c.Context, c)
	}
}
```



## 主要是两种功能，一个是传递value，一个就是控制协程。



withCancel可以通过 cancelFunc 随时取消协程

withDeadline withTimeOut ，然后





- 使用源码，然后

```
ackage main
 
import (
	"context"
	"errors"
	"fmt"
	"time"
)
 
var c = 1
 
func doSome(i int) error {
	c++
	fmt.Println(c)
	if c > 3 {
		return errors.New("err occur")
	}
	return nil
}
 
func speakMemo(ctx context.Context, cancelFunc context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("ctx.Done")
			return
		default:
			fmt.Println("exec default func")
			err := doSome(3)
			if err != nil {
				fmt.Printf("cancelFunc()")
				cancelFunc()
			}
		}
	}
}
 
func main() {
	rootContext := context.Background()
	ctx, cancelFunc := context.WithCancel(rootContext)
	go speakMemo(ctx, cancelFunc)
	time.Sleep(time.Second * 5)
}

```


