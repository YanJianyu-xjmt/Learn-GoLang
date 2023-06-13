# context包

## 0 context define

context.Context is an interface

### 0.2 Deadline()

return the deadline of goroutine  

bool是设没设置deadline

## 0.3 Done

done return a read-only channel  

## 0.4 Value

获取keyvS

```
type Context interface {
    // Deadline returns the time when work done on behalf of this context
    // should be canceled. Deadline returns ok==false when no deadline is
    // set. Successive calls to Deadline return the same results.
    Deadline() (deadline time.Time, ok bool)

    // Done returns a channel that's closed when work done on behalf of this
    // context should be canceled. Done may return nil if this context can
    // never be canceled. Successive calls to Done return the same value.
    // The close of the Done channel may happen asynchronously,
    // after the cancel function returns.
    //
    // WithCancel arranges for Done to be closed when cancel is called;
    // WithDeadline arranges for Done to be closed when the deadline
    // expires; WithTimeout arranges for Done to be closed when the timeout
    // elapses.
    //
    // Done is provided for use in select statements:
    //
    //  // Stream generates values with DoSomething and sends them to out
    //  // until DoSomething returns an error or ctx.Done is closed.
    //  func Stream(ctx context.Context, out chan<- Value) error {
    //      for {
    //          v, err := DoSomething(ctx)
    //          if err != nil {
    //              return err
    //          }
    //          select {
    //          case <-ctx.Done():
    //              return ctx.Err()
    //          case out <- v:
    //          }
    //      }
    //  }
    //
    // See https://blog.golang.org/pipelines for more examples of how to use
    // a Done channel for cancellation.
    Done() <-chan struct{}

    // If Done is not yet closed, Err returns nil.
    // If Done is closed, Err returns a non-nil error explaining why:
    // Canceled if the context was canceled
    // or DeadlineExceeded if the context's deadline passed.
    // After Err returns a non-nil error, successive calls to Err return the same error.
    Err() error

    // Value returns the value associated with this context for key, or nil
    // if no value is associated with key. Successive calls to Value with
    // the same key returns the same result.
    //
    // Use context values only for request-scoped data that transits
    // processes and API boundaries, not for passing optional parameters to
    // functions.
    //
    // A key identifies a specific value in a Context. Functions that wish
    // to store values in Context typically allocate a key in a global
    // variable then use that key as the argument to context.WithValue and
    // Context.Value. A key can be any type that supports equality;
    // packages should define keys as an unexported type to avoid
    // collisions.
    //
    // Packages that define a Context key should provide type-safe accessors
    // for the values stored using that key:
    //
    //     // Package user defines a User type that's stored in Contexts.
    //     package user
    //
    //     import "context"
    //
    //     // User is the type of value stored in the Contexts.
    //     type User struct {...}
    //
    //     // key is an unexported type for keys defined in this package.
    //     // This prevents collisions with keys defined in other packages.
    //     type key int
    //
    //     // userKey is the key for user.User values in Contexts. It is
    //     // unexported; clients use user.NewContext and user.FromContext
    //     // instead of using this key directly.
    //     var userKey key
    //
    //     // NewContext returns a new Context that carries value u.
    //     func NewContext(ctx context.Context, u *User) context.Context {
    //         return context.WithValue(ctx, userKey, u)
    //     }
    //
    //     // FromContext returns the User value stored in ctx, if any.
    //     func FromContext(ctx context.Context) (*User, bool) {
    //         u, ok := ctx.Value(userKey).(*User)
    //         return u, ok
    //     }
    Value(key any) any
}
```

## 2 官方的包里面的函数

## 2.2 WithCancels

这里就是生成一个CancelFunc 

```
func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
```

返回以一个和parent拷贝的新done channel，当context的cancel被调用的时候，或者父context 的done的时候。

例子

```
package main

import (
    "context"
    "fmt"
)

func main() {
    // gen generates integers in a separate goroutine and
    // sends them to the returned channel.
    // The callers of gen need to cancel the context once
    // they are done consuming generated integers not to leak
    // the internal goroutine started by gen.
    gen := func(ctx context.Context) <-chan int {
        dst := make(chan int)
        n := 1
        go func() {
            for {
                select {
                case <-ctx.Done():
                    return // returning not to leak the goroutine
                case dst <- n:
                    n++
                }
            }
        }()
        return dst
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel() // cancel when we are finished consuming integers

    for n := range gen(ctx) {
        fmt.Println(n)
        if n == 5 {
            break
        }
    }
}
```

## 2.3 WithCancelCause

```
func WithCancelCause(parent Context) (ctx Context, cancel CancelCauseFunc)
```

和上面的类似

```
ctx, cancel := context.WithCancelCause(parent)
cancel(myError)
ctx.Err() // returns context.Canceled
context.Cause(ctx) // returns myError
```

主要是可以记录error 

## 2.4 WithDeadline

和cancel类似，但是不会超过d，这个时间，如果父亲的早，那会和父亲的一样；如果

```

```

## 2.5 Background

返回一个非空的的context ，永远不会被cancel deadline

## 2.6 todo

主要是用于还没决定用什么context的时候

## 2.7 WithValue

```
func WithValue(parent Context, key, val any) Context
```

写入keyval 

# 

# 3 源码解读

## 3.2 backGroud

```
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
// An emptyCtx is never canceled, has no values, and has no deadline. It is not
// struct{}, since vars of this type must have distinct addresses.
type emptyCtx int
```

返回的都是emptyCtx，emptyCtx 实际上就是个int

```
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
```

实际上简单实现非常简单

## 3. Cancel

```
// A cancelCtx can be canceled. When canceled, it also cancels any children
// that implement canceler.
type cancelCtx struct {
    Context

    mu       sync.Mutex            // protects following fields
    done     atomic.Value          // of chan struct{}, created lazily, closed by first cancel call
    children map[canceler]struct{} // set to nil by the first cancel call
    err      error                 // set to non-nil by the first cancel call
}
```

这里的组合非常。context。

然后 包含了一些标志位和 锁  done应该是标志位，children 应该是子的方法
}

```


func (c *cancelCtx) Value(key interface{}) interface{} {
	if key == &cancelCtxKey {
		return c
	}
	return c.Context.Value(key)
}

func (c *cancelCtx) Done() <-chan struct{} {
	d := c.done.Load()
	if d != nil {
		return d.(chan struct{})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d = c.done.Load()
	if d == nil {
		d = make(chan struct{})
		c.done.Store(d)
	}
	return d.(chan struct{})
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
	d, _ := c.done.Load().(chan struct{})
	if d == nil {
		c.done.Store(closedchan)
	} else {
		close(d)
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

比较有看点的就这两个函数，首先Done 

先看有没有就返回



然后 cancel 也挺有意思，要设置 err
