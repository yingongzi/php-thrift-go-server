# go-timer：适合于高并发服务的 Timer 库 #

`Timer` 是 Go 里面非常常用的组件，标准库的 `Timer` 内部实现已经十分高效，不过在高并发的情况下会造成 Go runtime 时钟队列过长，造成性能下降，考虑到实际中我们对 `Timer` 精度要求不高（例如精确到 1ms），且大多超时都很短（例如 1s 以内），所以使用合适的缓存机制可以显著降低时钟队列长度、减轻 GC 压力、提升整体程序性能。

## 接口文档 ##

详见 godoc 文档：[git.xiaojukeji.com/soda-framework/go-timer](https://git.xiaojukeji.com/soda-framework/soda-go-framework-explained/tree/master/projects/go-timer)

## 使用方法 ##

创建一个 `Timer`。

```go
timer := timer.After(100 * time.Millisecond)
```

创建之后，推荐使用 `timer.Expired()` 来判断是否超时，这个函数仅仅是简单的比较一下记录的超时时间是否超过 `time.Now()`，非常轻量。

如果一定需要同步等待，可以通过 `time.Done()` 得到一个 `chan struct{}`，通过 `select` 来等待超时发生。

```go
select {
case <-timer.Done():
    // 超时了。
}
```

需要注意，这个 `chan` 不会返回任何信息，只会被关闭，因此只能通过 `select` 或 `for...range` 来检查是否关闭，而不能写 `data := <-timer.Done()`。

## `Timer` 池 ##

默认的 `Timer` 精度是 1ms，如果有特殊需求希望改变精度，可以自行创建一个 `Pool`。

```go
// Pool 可以用于创建很多 timer。
pool := timer.NewPool(10 * time.Millisecond)

// 使用 pool。
timer := pool.After(2 * time.Second)
```
