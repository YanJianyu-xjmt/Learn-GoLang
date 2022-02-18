# log 笔记

## - 介绍

- ```
  type Logger struct {
      mu     sync.Mutex // ensures atomic writes; protects the following fields
      prefix string     // prefix to write at beginning of each line
      flag   int        // properties
      out    io.Writer  // destination for output
      buf    []byte     // for accumulating text to write
  }
  ```
  
- 这里是结构体 锁 前缀 优先级 输入writer buf
  

实际上ioWriter 是一个interface

这里可以看出

```
func New(out io.Writer, prefix string, flag int) *Logger {
    return &Logger{out: out, prefix: prefix, flag: flag}
}
```

通过New 新的结构体

```
var std = New(os.Stderr, "", LstdFlags)
```

这里std是打印到屏幕上的结构体

然后这里打印的时候 会formatHeader 函数，打印新的前缀

然后output输出

```
func (l *Logger) Output(calldepth int, s string) error {
    now := time.Now() // get this early.
    var file string
    var line int
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.flag&(Lshortfile|Llongfile) != 0 {
        // release lock while getting caller info - it's expensive.
        l.mu.Unlock()
        var ok bool
        _, file, line, ok = runtime.Caller(calldepth)
        if !ok {
            file = "???"
            line = 0
        }
        l.mu.Lock()
    }
    l.buf = l.buf[:0]
    l.formatHeader(&l.buf, now, file, line)
    l.buf = append(l.buf, s...)
    if len(s) == 0 || s[len(s)-1] != '\n' {
        l.buf = append(l.buf, '\n')
    }
    _, err := l.out.Write(l.buf)
    return err
}
```

以后可以看下 这个 runtime.Caller(calldepth) 这里一般是2

对于 Print Println 直接就是打印给的

Fatal Fatalf 就是打印字符串之后 os。exit

panic panicf panicln 就是打印后panic

New prefix 可以设置为 Info error 这样

flag是优先级

然后runtimer.Calle好像是打印 堆栈的深度

比如 func a（）{

log。Ouput 

}

那么会打印 a这个函数的文件和文件名
