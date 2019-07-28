# go-trace：通用 trace 信息存储 #

通过现代的方法在服务间传递 trace 信息。

## 接口文档 ##

详见 godoc 文档：[git.xiaojukeji.com/soda-framework/go-trace](https://git.xiaojukeji.com/soda-framework/soda-go-framework-explained/tree/master/projects/go-trace)

## 使用方法 ##

### 创建新 context ###

在实现任何 server 接口时需要先创建一个 context，如果接口传入了上游的 trace 信息，需要作为参数一起参与初始化。

```go
// traceInfo 是一个 map[string]string 类型，可以为 nil。
ctx := trace.NewContext(context.Background(), trace.Trace(traceInfo))
```

注意这里的父 context 是 `context.Background()`，可以根据实际情况继承已有的 ctx 或者直接指定一些额外参数。后续 server 所有的耗时操作之前需要检查 ctx 是否已经超时。

### 使用 trace 信息 ###

可以从 ctx 里面拿到 Trace 数据结构，详见 Trace 的方法。

```go
tr := trace.FromContext(ctx)

// 使用 tr 中的内容。
traceid := tr.Traceid()

// 可以直接输出关键信息，Trace 实现了 String() 方法。
log.Tracef("some msg.||%v||foo=%v", trace.FromContext(ctx), foo)
```

### 为一个 ctx 添加调用者信息 ###

当调用一个 rpc 服务的时候，我们需要指定各种调用者信息，可以使用 `WithInfo` 来设置这些信息。

```go
// 假设 client.Server1 创建一个 server1 的 thrift client。
server1 := client.Server1(trace.WithInfo(ctx, &Info{
    SrcMethod: "src/method/name",
    Caller: "caller.name",
    Callee: "callee.name",
}))

// 调用一个 rpc 方法。
server1.Foo()
```

需要注意，`Info` 一般都是跟单次 rpc 调用相关的信息，一般推荐在使用时调用 `WithInfo` 来附加这个信息到 `ctx` 里面。Go 的 `ctx` 是不可变的（immutable），里面的数据一旦设置进去就不会发生变化，`WithInfo` 也符合这个约定，调用这个函数并不会修改原 `ctx` 的任何数据。

### 为 ctx 指定超时时间 ###

可以使用标准库的 `context.WithTimeout` 或 `context.WithDeadline` 来设置超时，一旦设置，所有的 rpc 都会自动的通过 `Trace` 来传播这个超时时间，从而做到整个服务调用链都能感知到这个超时。

```go
// 为 server1 设置一个超时（100ms）。
// 如果 ctx 本身就有超时并且小于这个值，会取最小值。
server1Ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
defer cancel() // 一定记得要 cancel。

// 创建 server1 client，这里略过代码……
```

### 设置虚拟时间 ###

在一些仿真场景下，业务期望使用一个自定义的时间为`time.Now`， `trace` 的 `FakeNow` 可以满足这个需求，timeout 等特性依然有效，通过 `trace.Now` 函数可以获取设置的仿真时间

```go
// 假设 client.Server1 创建一个 server1 的 thrift client。
server1 := client.Server1(trace.WithInfo(ctx, &Info{
    SrcMethod: "src/method/name",
    Caller: "caller.name",
    Callee: "callee.name",
    FakeNow:now.Add(-time.Hour * 24),
}))
```

### Trace id 生成规则 ###

Trace id 根据[公司 traceid 生成规则](http://wiki.intra.xiaojukeji.com/pages/viewpage.action?pageId=91930217)来产生。为了方便从 traceid 中读取一些有用信息，使用者可以通过 `Traceid` 这个类型来解析其中的信息，详见 [Traceid](https://git.xiaojukeji.com/soda-framework/soda-go-framework-explained/tree/master/projects/go-trace/go-trace.md#Traceid) 接口文档。