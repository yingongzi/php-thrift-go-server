# go-log #

通用日志库封装，定义了标准的日志接口，并且允许使用者定制日志具体实现。

## 接口文档 ##

详见 godoc 文档：[git.xiaojukeji.com/soda-framework/go-log](https://git.xiaojukeji.com/soda-framework/soda-go-framework-explained/tree/master/projects/go-log)

## 基本用法 ##

基本使用流程：

* 调用 `Init` 方法初始化日志；
* 使用 `Infof` 等方法输出日志。

### 初始化日志 ###

读取配置文件信息，填充 `Config` 结构，然后调用 `Init` 来初始化日志库。

```go
import (
    "git.xiaojukeji.com/soda-framework/go-log"
)

type Config struct {
    Log log.Config `toml:"log"`
}

func main() {
    var config Config

    if _, err := toml.DecodeFile("your/config/file", &config); err != nil {
        // 初始化之前也可以使用 log，此时会直接输出到 stderr。
        log.Printf("fail to read config.||err=%v", err)
        os.Exit(1)
        return
    }

    log.Init(&config.Log)
    defer log.Close() // 记得 main 退出的时候要关闭日志，确保所有日志落盘。

    // 之后再调用 log.Infof 等方法就会按照 config 的配置来输出日志。
}
```

### 使用日志 ###

可用的日志函数包括：

* `log.Printf`/`log.Print`：直接输出日志，无论配置为什么 level 都会输出，且不会追加任何的额外信息，就像 `fmt.Printf`/`fmt.Print` 一样仅输出指定的内容。
* `log.Debugf`/`log.Debug`：输出 DEBUG 级别日志。
* `log.Infof`/`log.Info`：输出 INFO 级别日志。
* `log.Warnf`/`log.Warn`：输出 WARN 级别日志。
* `log.Errorf`/`log.Error`：输出 ERROR 级别日志。
* `log.Fatalf`/`log.Fatal`：输出 FATAL 级别日志，并且用 `os.Exit(1)` 退出程序。
* `log.Panicf`/`log.Panic`：输出 PANIC 级别日志，并且用 `panic` 退出程序。

### 创建标准库的 `*Logger` ###

有些第三方库会接受标准库 `log` 中实现的 `*Logger`，使用 `NewStdLogger` 方法可以轻松的创建一个基于任意 `Logger` 的标准 `*Logger`。

需要注意，打印文件位置的功能可能会不太正常。由于标准库 `*Logger` 的 `Output` 方法并没有将 `calldepth` 传递到底层，这个参数并不能用来控制调用栈深度，必须使用 `NewStdLogger(logger Logger, level Level, skip int, flag int)` 中的 `skip` 参数来控制。默认情况下当 `skip` 等于 0 时候，调用栈显示的是 `*Logger` 的 `Output` 的调用者位置。如果希望使用标准库的 `*Logger` 其他函数，比如 `Print`，那么需要在创建时将 `skip` 设置为 1，不过这样就会让 `Output` 不能正常工作。因此，调用 `NewStdLogger` 必须想清楚这个标准 `*Logger` 是如何使用的，根据使用情况来决定 `skip` 如何设置。

```go
l := New("")
logger := NewStdLogger(l, INFO, 0, 0) // 以 INFO 级别来输入日志。
```

### 创建一个新的日志文件 ###

默认 `log.Infof` 等日志方法只会往一个文件里面写日志，比如当前我们的日志名默认为 `./log/all.log`。如果需要新建一个日志文件，可以使用 `log.New` 方法初始化一个新 `Logger` 并且放在某个全局可访问的地方，未来使用即可。实际上，`log.Init` 就是 `log.Use(log.New(config))` 的一个语法糖。

```go
logger := log.New("./log/my.log")

// 将 logger 存储在某个全局可访问的地方。

// 使用时调用 logger，而不是默认的 log 里面方法。
logger.Infof("foo")
```

### 将日志同时输出到文件和命令行 stderr ###

配置里面可以设置 `debug = true` 来打开这个功能，这样可以方便线下进行调试。

正如这个配置所暗示的，不要在生产环境中开启这个开关。

## 定制 `Logger` ##

我们可以用任何的日志封装来替换默认的 `Configer` 实现，从而使得这个日志库的各种函数可以被重新定义。

只需要实现 `Configer` 接口，并且用 `Use` 方法将实现了 `Logger` 接口的日志封装设置到日志库里面，所有的 `log.Infof` 等函数都会使用新设置的 `Logger`。
一旦重新设置了 `Configer`，所有通过 `New` 生成的 `Logger` 都会根据新配置重新初始化。