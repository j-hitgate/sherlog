# SherLog

<img src="logo.svg" alt="logo" width="300"/>

### Description

SherLog is a lightweight **Golang** library that provides a convenient and visual way to trace changes and actions performed on entities within a system. With SherLog, you can easily log:
- What happened to an entity;
- When and where it happened (in which service, process, component, or logic branch);
- And in what context, by attaching additional fields and labels.

SherLog is designed to quickly and clearly capture the execution path and details of changes to entities. This is especially useful for debugging, auditing, or tracing complex workflows in microservice architectures.

Here’s an example of terminal log output:
```
╭─[18:33:42]─( user : 54 )─{ order_service_3 >> post_order_API > DB_repo > AddUser }
╰─[ INFO ]─> User added
  ╰─> name: Juster
  ╰─> age: 26
```
Short format:
```
[18:33:42]─[ INFO ]─( user : 54 )─{ AddUser }─> User added
```

### Initializing

Before using the library, you need to initialize it by calling `sherlog.Init()` with a configuration struct and an optional log filter (you can pass `nil` if no filter is needed):
```go
sherlog.Init(sherlog.Config{}, &sherlog.Filter{})
```
The library is initialized only on the first call — any subsequent calls **are ignored silently without errors or panics**.

Configuration fields:
```go
type Config struct {
    Level           byte   // log level (from 0 to 7)
    SyncPrint       bool   // synchronous log printing
    NotShowLogs     bool   // suppress log output

    NotShowDatetime bool   // hide date and time
    ShowDate        bool   // show date
    ShowTimeDelta   bool   // show time delta between logs
    NotShowLevel    bool   // hide log level
    NotShowTraces   bool   // hide trace chain
    NotShowModules  bool   // hide module chain
    NotShowEntity   bool   // hide entity
    NotShowEntityId bool   // hide entity ID
    NotShowLabels   bool   // hide labels
    NotShowFields   bool   // hide additional fields
    ShortMode       bool   // use short log output format

    LogsDir         string // path to the directory where logs should be stored; if empty, logs will not be saved to disk
    AutodumpAfter   int    // number of logs to buffer before auto-dumping to disk (set to 0 to disable auto-dumping and use only manual flushing)
}
```

Log filter lets you exclude certain logs from being created. You can filter logs by the following fields:
```go
type Filter struct {
    Traces   []string          // by trace names
    Modules  []string          // by module names
    Entities []string          // by entity classes
    Labels   []string          // by labels
    Fields   map[string]string // by key-value pairs in additional fields
    Invert   bool              // if "true", only logs matching the filter will be created (inverse filtering)
}
```

### Start of tracing

To trace the execution flow in SherLog, you use the `sherlog.Trace` structure. It allows you to build a chain of events that represent the execution path of a specific process/goroutine.

To begin tracing, create a `Trace` by specifying the name of the current process/goroutine:
```go
trace := sherlog.NewTrace("some_process")
```
A `Trace` represents a single execution flow (process/thread/goroutine) and **is not thread-safe** — it is designed to be used synchronously.

If your execution flow transitions to another process, you can use the `Fork()` method to create a copy of the current trace with a new process name appended to the chain. This forked trace can be used concurrently in another flow. For branching into multiple parallel traces, use `ForkOnMap(name ...string)` method, which returns a map of trace copies using the provided names as both keys and process names:
```go
main_trace := sherlog.NewTrace("main_trace")
main_trace.TraceChain()  // main_trace

child_trace := main_trace.Fork("child_trace")
child_trace.TraceChain() // main_trace > child_trace

traces := child.ForkOnMap("trace_1", "trace_2", "trace_3")
traces["trace_1"].TraceChain() // main_trace > child_trace > trace_1
traces["trace_2"].TraceChain() // main_trace > child_trace > trace_2
traces["trace_3"].TraceChain() // main_trace > child_trace > trace_3
```
After you're done using a trace (e.g., the goroutine has finished), you should close it by calling `Close()` to avoid resource leaks. Alternatively, you can use `sherlog.CloseTraces(map[string]*Trace)` to close multiple traces in map at once.

You can also associate an entity class and ID with the trace using the `SetEntity(name, id string)` method. For example, to track a user with ID 5 or a task with ID 35a4f6:
```go
trace1.SetEntity("user", "5")
trace2.SetEntity("task", "35a4f6")
```

### Modules

To specify where a log was sent from (which microservice, class, function, etc.), you can add modules to a trace by calling the `AddModule(group, module string)` method, thereby creating chains of modules. It works as follows:
```go
popModule1 := trace.AddModule("", "module_1")
trace.INFO(nil, "message") // log will have the chain: module_1

popModule2 := trace.AddModule("", "module_2")
trace.INFO(nil, "message") // log will have the chain: module_1 > module_2

popModule1()               // remove modules from the trace
popModule2()
trace.INFO(nil, "message") // log will have no chain
```

All logs sent from this trace after adding a module will include it in their chain, indicating the origin of the log. The `popModule()` is a callback function that removes the added module from the trace. This allows adding and removing modules in one line using `defer` if you are inside a function:
```go
func fn() {
    defer trace.AddModule("", "some_module")()
    trace.INFO(nil, "message")
}
```
Another way to add a module is using the `WithModule()` method, which takes a function as a parameter; logs created within this function will have the specified module:
```go
trace.WithModule("", "some_module", func() {
    trace.INFO(nil, "message")
})
```
Sometimes modules can be nested or part of another module. For example, if a class method calls another method of the same class, and both the class and method are modules, to avoid duplicates in the chain (like `class > method_1 > class > method_2`), modules can be grouped:
```go
type Class struct{
    Trace *sherlog.Trace
}

func (c *Class) Method_1() {
    defer c.Trace.AddModule("class", "method_1")()
    c.Trace.INFO(nil, "message") // chain: service > class > method_1
    c.method_2()
}

func (c *Class) method_2() {
    defer c.Trace.AddModule("class", "method_2")() // trace is already in "class" group, so it's ignored and not duplicated
    c.Trace.INFO(nil, "message") // chain: service > class > method_1 > method_2
}

func main() {
    sherlog.Init(sherlog.Config{}, nil)
    trace := sherlog.NewTrace("main_trace")
    defer trace.AddModule("", "service")()

    c := &Class{Trace: trace}
    c.Method_1()
}
```
The trace compares the given group with the current one; if they don't match, the group is added as a new module. Otherwise, it ignores it to prevent duplication.
Passing an empty string as the group means the module has no group and only the module itself is added.

### Logs

To create a log entry, you need to call one of the following logging methods on the trace. Each method corresponds to a certain importance level of the message:
- `FATAL()` (level 7, Fatal) – a critical error after which the program cannot continue (after calling, logs are flushed to disk if enabled, and the program exits with `os.Exit(1)`);
- `ERROR()` (level 6, Error) – a serious error, but the program can still continue running;
- `WARN()`(level 5, Warning) – a warning about abnormal or unexpected behavior;
- `INFO()`(level 4, Information) – reports about program actions during execution or operations on the tracked entity. These are not technical details but logs for regular users, analysts, DevOps, etc.;
- `NOTE()`(level 3, Note) – useful general information (mostly for analytics), e.g. user input errors, validation failures, some events, and more;
- `STAGE()` (level 2, Stage) – key stages of the program operation or the entity’s workflow;
- `DEBUG()` (level 1, Debug) – technical details for developers;
- `MICRO()` (level 0, Micro) – very fine-grained details of program execution; for example, values of variables on each loop iteration.

Levels can be roughly grouped as follows: 0–2 for developers, 3–4 for users, 5–7 is errors.

Each logging method accepts an attribute (or `nil` if none is needed) and a set of arguments to form the message:
```go
num := 5
msg := "do abcd"
trace.INFO(nil, "Task number ", num, " -> message: ", msg) // "Task number 5 -> message: do abcd"
```

There are also conditional methods with the suffixes `*_if_err` and `*_if_not_err` that create logs only if an error exists or does not exist (the error is passed as a parameter):
```go
err := errors.New("some error")
trace.DEBUG_if_not_err(err, nil, "message") // will not create a log because error exists
err = nil
trace.ERROR_if_err(err, nil, "message")     // will not create a log because error is nil
```

### Attributes

In log attributes, you can pass the following:
- Labels (`sherlog.Labels`) — a slice of strings used to group logs:
```go
trace.NOTE(sherlog.Labels{"auth", "user error"}, "Incorrect password")
```
- Additional fields (`sherlog.Fields`) — a map with extra information where both keys and values are strings:
```go
trace.INFO(sherlog.Fields{"name": "Juster", "age": "26"}, "User created")

// Or via a helper function that converts keys and values to strings using fmt.Sprint
trace.DEBUG(sherlog.WithFields("ID", 5, "is private": true), "Task added")
```
- A group of attributes (`sherlog.Attr`) — both labels and fields together:
```go
trace.INFO(&sherlog.Attr{
    Labels: sherlog.Labels{"auth", "customer"},
    Fields: sherlog.Fields{
        "name": "Maria",
        "age": "23",
    },
}, "User created")
```

### Additional

The log output format with all blocks in the terminal looks like this:
```
╭─[26.03.2025 18:33:42 +10ms]─( entity : id )─{ trace1 > trace2 >> module1 > module2 }─(label1, label2)
╰─[ INFO ]─> Some message
  ╰─> key1: val1
  ╰─> key2: val2
```
Or the short format:
```
[26.03.2025 18:33:42 +10ms]─[ INFO ]─( entity : id )─{ module2 }─> Some message
```

- At the end of the program’s execution, you should call `sherlog.Close()` to properly close traces and flush buffered logs to disk (if enabled);
- You can manually flush logs from the buffer by calling `sherlog.DumpLogs()`;
- Logs are stored as JSON objects in files named like "*26.03.2025.log*", i.e., logs are grouped by day, which simplifies aggregation and archiving.