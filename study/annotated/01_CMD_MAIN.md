# Annotated Source: `cmd/main.go`

This is the **actual executable entry point** for the Go runtime.

Do not worry about memorizing syntax. Read for intent.

---

## Package declaration

```go
package main
```

In Go, an executable program uses `package main`.

A `main` package with a `func main()` is analogous to a Python script with:

```python
if __name__ == "__main__":
    main()
```

---

## Imports

The standard imports cover:

- `cmp` — helper used here for selecting environment/default value
- `context` — cancellation/shutdown propagation
- `flag` — command-line flags
- `fmt` — formatted output
- `os` — environment, stderr, exit
- `os/signal` — OS signal handling
- `syscall` — SIGINT/SIGTERM constants
- `time` — durations/timeouts

OM1 imports:
- config
- logger
- metrics
- runtime

Then these strange imports:

```go
_ "github.com/openmind/om1/plugins/actions"
_ "github.com/openmind/om1/plugins/backgrounds"
_ "github.com/openmind/om1/plugins/inputs"
_ "github.com/openmind/om1/plugins/llm"
```

### Why underscore imports?

They load plugin packages for **side effects**.

The side effect is primarily:

> package `init()` functions register plugin constructors into registries.

Without these imports, the config might request `GeminiLLM`, but no Gemini constructor would have registered itself.

This is a major architecture pattern.

---

# `func main()`

```go
func main() {
```

This is where the process begins.

---

## CLI flags

```go
configName = flag.String("config", "", "config name or path (required)")
logLevel   = flag.String("log-level", ...)
hotReload  = flag.Bool("hot-reload", false, ...)
checkSecs  = flag.Float64("check-interval", 1.0, ...)
```

These functions return pointers to flag values.

That is why later code uses:

```go
*configName
```

The `*` dereferences the pointer.

Python mental model:

```python
args.config
args.log_level
args.hot_reload
```

---

## Parse arguments

```go
flag.Parse()
```

Similar to:

```python
args = parser.parse_args()
```

---

## Require a config

```go
if *configName == "" {
    fmt.Fprintln(os.Stderr, "error: --config is required")
    flag.Usage()
    os.Exit(1)
}
```

Straightforward validation.

The whole OM1 runtime depends on configuration, so it refuses to start without one.

---

# Logging setup

```go
log := logger.BuildLogger(*logLevel)
logger.Set(log)
defer func() { _ = log.Sync() }()
```

### `:=`

Declares and assigns `log`.

### `defer`

Says:

> when `main` exits, call `log.Sync()`.

This is cleanup behavior.

Python analogy:

```python
try:
    ...
finally:
    log.sync()
```

---

# Metrics server

```go
stopMetrics := metrics.StartServer(log)
```

Metrics are started early so runtime performance/health can be observed.

Later the deferred cleanup creates a two-second shutdown context and stops metrics cleanly.

This is an example of production-system concerns appearing before the AI logic:
- observability
- shutdown
- timeouts

---

# Load configuration

```go
cfg, err := config.Load(*configName)
if err != nil {
    log.Fatal("failed to load config", zap.Error(err))
}
```

Classic Go:

> call function → receive value and error → explicitly test error.

The config is the blueprint for the agent.

At this point no central runtime loop has started yet.

---

# Create shutdown-aware context

```go
ctx, cancel := signal.NotifyContext(
    context.Background(),
    syscall.SIGINT,
    syscall.SIGTERM,
)
defer cancel()
```

This creates a Context that becomes canceled when the process receives common shutdown signals.

Why this matters:

OM1 has many goroutines/subsystems.

Rather than manually telling each one:

> "please stop"

the runtime passes a shared context down. When canceled, workers can exit.

Conceptual tree:

```text
process context
  ├─ input workers
  ├─ cortex loop
  ├─ action workers
  ├─ backgrounds
  └─ tracer
```

Cancellation propagates downward.

---

# Construct Runtime

```go
rt := runtime.New(cfg, log, runtime.Options{
    HotReload:     *hotReload,
    CheckInterval: *checkSecs,
})
```

This is where configuration becomes the central Runtime object.

Important distinction:

> `runtime.New` constructs it.

> `rt.Run` starts it.

This constructor/start separation is common systems design.

---

# Start Runtime

```go
if err := rt.Run(ctx); err != nil && err != context.Canceled {
    log.Fatal("runtime exited with error", zap.Error(err))
}
```

This compact syntax:

```go
if err := rt.Run(ctx); condition {
}
```

means:

1. call `Run`
2. store its result in local variable `err`
3. evaluate the condition

The process treats normal context cancellation as expected shutdown rather than a fatal failure.

---

# Whole file in one sentence

`cmd/main.go`:

> parses startup options, initializes operational infrastructure, loads an agent configuration, creates a cancellable OM1 Runtime, and runs it until shutdown.

---

# What this file deliberately does NOT do

It does not:
- read cameras
- transcribe audio
- build prompts
- call Gemini directly
- move a robot
- publish `cmd_vel`

Those responsibilities are delegated into packages/plugins.

This tells you something important about architecture:

> `main.go` is composition/lifecycle glue, not business logic.

---

# Follow the call now

The next file to read is:

> `internal/runtime/runtime.go`

Specifically, find:

```go
func (rt *Runtime) Run(ctx context.Context) error
```

Then continue to:

```go
func (rt *Runtime) tick(...)
```

Use:

> [Runtime Walkthrough](../03_RUNTIME_WALKTHROUGH.md)

and then:

> [Annotated Cortex Tick](02_RUNTIME_TICK.md)
