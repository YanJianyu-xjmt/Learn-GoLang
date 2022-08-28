## Garbage Collection

## 0 golang 1.0

mark-sweep 

完全stw 事件过长

## 1 后来采用三色标记法

    尽量渐少stw的阶段

### 1.1 基础概念

    root 对象 

- 全局变量  编译器就存在整个生命周期的变量

- 执行栈 goroutine都有执行栈，执行栈包含栈上变量还有堆

### 1.2 需要解决的问题

如下

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/d0dd76a2485c4cb6b91de118d3d3467f~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/97a29c63f04b47a48d8fe36f100ec336~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/2988cf657d3d4110833a4a55ee4e5d0c~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/60e52da7d2344b64af1d3f72ddde063a~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

## 2 学术上说法

### 2.0 造成误回收的条件

- 在并发之中，给黑色对象插上白色对象

- 在并发之中，灰色对象到白色对象的路径被破坏（但是这个对象又被 黑色引用，是不能被回，如果没有被黑色引用，那删了无所谓；如果灰色对象的引用没有破坏那无所谓）

### 2.1 强弱三色不变式

- 强三色不变
  
  - 迪捷斯特方法 就是禁止黑色对象直接引用白色对象

- 弱三色不变
  
  - 黑色对象可以引用白色对象，但是白色对象上游一定要有灰色对象

### 2.2 插入写屏障

- 思路 
  
  | 满足强三色不变，白色一定是被灰色引用

    就是在插入白色对象的时候，会把引用的指针标记成灰色的，那么新插入的对象就被灰色对象引用了

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/b83ce89d105b486eaae20536846353d7~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/f73d3f2199ac41eaa4b4381f9956d920~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/a919c3e04fe24df0b8349c644a79cc55~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/dcdc331c01484fd794263e36ae26c781~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

这里可看出 有一个弊端， 就是一个对象引用被删除之后 



主要的缺点是： golang实现这种需要靠编译插入代码，  尤其是栈上操作如果采用这种，会带来比较大的性能负担。 

下面这个存疑，我还没想通: 就是在正常三色标记结束之后，需要对栈上重新rescan 且 stw，可能是回收指针这种灰色对象？



### 2.3 删除写屏障

核心规则：在删除引用时候，如果对象是灰色或者白色，直接把对象搞成灰色

| 满足弱三色不变性，灰色对象到白色对象的路径不会断

白色对象始终会被灰色保护

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/b83ce89d105b486eaae20536846353d7~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/f73d3f2199ac41eaa4b4381f9956d920~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/a919c3e04fe24df0b8349c644a79cc55~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/dcdc331c01484fd794263e36ae26c781~tplv-k3u1fbpfcp-zoom-in-crop-mark:3024:0:0:0.awebp)

但是问题就是可能会减少 回收精度比较低，很多该收的没有收



### 2.4 混合写屏障

结合两种的优势

结合删除写的思想

## 2.5 GOLANG 的具体实现

- 打开是辅助GC 这一步STW

- GC 刚开始的时候 所有栈都被标记为黑色 

- GC 任务栈上新创建的对象，均为黑色



- 堆上被删除的对象标记为灰色

- 堆上新添加的对象标记为灰色

## 3 辅助GC

为了防止 在GC并发期间一直插入，插入比回收多那就会结束不了；为了解决这个 问题，引入辅助GC，在一开始就标记回收任务，标记完就结束。
