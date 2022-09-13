# tcmallo

## 0 introduction

    thread caching malloc 思想是内存多级管理，进而降低锁的粒度，把内存按需                         划成大小不一样的块，减少内存的碎片化。

    每个协程go 调度模型维护一个mcache 结构体的独立内存池，从而不用加锁，加快内存分配速度。只有当内存池不足时，才会向全局mcentra和mheap结构体申请内存。


