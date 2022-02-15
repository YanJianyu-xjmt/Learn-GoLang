# Interface 实现

## 记录

```
type iface struct {
    tab  *itab
    data unsafe.Pointer
}
// eface empty face 结构体
type eface struct {
    _type *_type
    data  unsafe.Pointer
}
```

eface 指的空interface 没有一个函数的

iface 指非空inferace

eface 包含两个 字段一个指向 _type 另外一个是data的指针 如果数据本身就是一个指针 那么就是本身

ifcae 指向itab itab 中包含一个 _TYPE 指针

```
type itab struct {
    inter  *interfacetype
    _type  *_type
    link   *itab
    bad    int32
    unused int32
    fun    [1]uintptr // variable sized
}
```

这里是新的itab
