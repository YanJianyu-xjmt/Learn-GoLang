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
