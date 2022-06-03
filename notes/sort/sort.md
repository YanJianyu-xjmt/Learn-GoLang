# Sort

## 1 要实现的接口

这是要实现的接口

```
type Interface infterface{
    Len()int 
    Less(i,j int) bool
    Swap(i,j int)
}
```

len 记得是要记录的数目 

Less mean i是否一定要排在j的前面



## 2 排序的API

insertionSort(data Interface，a,b, int)

插入排序 前闭后开 和 切片一样



siftDown 实现 堆优先级在 data[lo:hi ] ,

就是实现siftDown 实现 调整



heapSort 堆排序，默认是把最大的放在堆顶



## 3 对外可以使用的借口



### Sort 接口

就是全部进行 快速拍寻，然后快排有个最大递归深度



### IsSorted

是否排序





然后[]int 是实现的







## 默认的快速排序



有一个递归栈的排序

如果到了最大递归深度 会结束快速排序变成堆排





## Slice

Slice（X INTERFACE{}, less func(i，x)）

```
然后slice 
```

排序





SliceStable

稳定的排序



SliceIsSorted



### Search

Search 就是int 类型

func Search(n int, f func(int) bool) int

用法

```
a := []int{1,2,3,4,5}
    d := sort.Search(len(a), func(i int) bool { return a[i]>=3})
    fmt.Println(d)
```





SearchInts

SearchFloat64s

SearchStrings1


