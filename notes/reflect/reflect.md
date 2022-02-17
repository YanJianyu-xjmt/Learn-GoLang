# 反射笔记

## 简介

每一个结构体的 _type 指向真实 在编译的时候生成

在reflect/type.go 文件下

主要是两个方法

```
ValueOF 获取输入参数接口中的数据的值  如果为空返回0
TypeOf 动态获取输入参数接口中值的类型 如果为空返回nil 返回比如 int/float/struct等
```

## TypeOf

首先是 TyepOf 这里的代码

```
// TypeOf returns the reflection Type that represents the dynamic type of i.
// If i is a nil interface value, TypeOf returns nil.
func TypeOf(i interface{}) Type {
    eface := *(*emptyInterface)(unsafe.Pointer(&i))
    return toType(eface.typ)
}
```

先把输入 转化风险指针 然后转化为 eface  这里

另外说一句  这里的返回值Type 是一个interface 所有的go里面都这样

```
type Type interface {
    // Underlying returns the underlying type of a type.
    Underlying() Type

    // String returns a string representation of a type.
    String() string
}
```

然后这里的具体的Type Struct为 在reflect/type.go中

```
type Type interface {
    Align() int
    FieldAlign() int
    Method(int) Method
    MethodByName(string) (Method, bool)
    NumMethod() int
    Name() string
    PkgPath() string
    Size() uintptr
    String() string
    Kind() Kind
    Implements(u Type) bool
    AssignableTo(u Type) bool
    ConvertibleTo(u Type) bool
    Comparable() bool
    Bits() int

    // ChanDir returns a channel type's direction.
    // It panics if the type's Kind is not Chan.
    ChanDir() ChanDir

    // IsVariadic reports whether a function type's final input parameter
    // is a "..." parameter.  If so, t.In(t.NumIn() - 1) returns the parameter's
    // implicit actual type []T.
    //
    // For concreteness, if t represents func(x int, y ... float64), then
    //
    //    t.NumIn() == 2
    //    t.In(0) is the reflect.Type for "int"
    //    t.In(1) is the reflect.Type for "[]float64"
    //    t.IsVariadic() == true
    //
    // IsVariadic panics if the type's Kind is not Func.
    IsVariadic() bool

    // Elem returns a type's element type.
    // It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice.
    Elem() Type

    // Field returns a struct type's i'th field.
    // It panics if the type's Kind is not Struct.
    // It panics if i is not in the range [0, NumField()).
    Field(i int) StructField

    // FieldByIndex returns the nested field corresponding
    // to the index sequence.  It is equivalent to calling Field
    // successively for each index i.
    // It panics if the type's Kind is not Struct.
    FieldByIndex(index []int) StructField

    // FieldByName returns the struct field with the given name
    // and a boolean indicating if the field was found.
    FieldByName(name string) (StructField, bool)

    // FieldByNameFunc returns the first struct field with a name
    // that satisfies the match function and a boolean indicating if
    // the field was found.
    FieldByNameFunc(match func(string) bool) (StructField, bool)

    // In returns the type of a function type's i'th input parameter.
    // It panics if the type's Kind is not Func.
    // It panics if i is not in the range [0, NumIn()).
    In(i int) Type

    // Key returns a map type's key type.
    // It panics if the type's Kind is not Map.
    Key() Type

    // Len returns an array type's length.
    // It panics if the type's Kind is not Array.
    Len() int

    // NumField returns a struct type's field count.
    // It panics if the type's Kind is not Struct.
    NumField() int

    // NumIn returns a function type's input parameter count.
    // It panics if the type's Kind is not Func.
    NumIn() int

    // NumOut returns a function type's output parameter count.
    // It panics if the type's Kind is not Func.
    NumOut() int

    // Out returns the type of a function type's i'th output parameter.
    // It panics if the type's Kind is not Func.
    // It panics if i is not in the range [0, NumOut()).
    Out(i int) Type

    common() *rtype
    uncommon() *uncommonType
}
```

这就返回的结构体

然后可以通过 获取字段

FieldByIndex 通过来看 在这里 fieldByName如果不是struct 那么会panic 后续因该就是根据名字找



## - ValueOf

```

```


