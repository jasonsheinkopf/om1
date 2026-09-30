# Go Starter Pack for a Python Developer Reading OM1

> **Looking for the full course?** This file is now the compact reference. For the detailed offline course with OM1 examples, concurrency, exercises, end-to-end runtime tracing, and interview practice, start at [study/go-course/README.md](go-course/README.md).

---

This is **not** a complete Go course.

It contains the Go concepts you need so OM1 source code stops looking unfamiliar.

---

# 1. Packages are roughly Python modules/packages

Go:

```go
package runtime
```

Python analogy:

```python
# runtime.py
```

Files in the same Go package share package-level names.

Imports:

```go
import (
    "context"
    "time"

    "github.com/openmind/om1/internal/actions"
)
```

Python analogy:

```python
import context_like_thing
import time
from om1.internal import actions
```

Go is stricter: unused imports are compile errors.

---

# 2. `func` = function

Go:

```go
func main() {
    // ...
}
```

Python:

```python
def main():
    ...
```

With parameters and return values:

```go
func Add(a int, b int) int {
    return a + b
}
```

Python:

```python
def add(a: int, b: int) -> int:
    return a + b
```

---

# 3. `:=` means "declare and assign"

Very common in OM1:

```go
cfg, err := config.Load(*configName)
```

Rough Python analogy:

```python
cfg, err = config.load(config_name)
```

The Go difference is that `:=` creates local variables and lets the compiler infer their types.

Normal assignment to an existing variable uses `=`.

---

# 4. Go returns errors as values

You will see this constantly:

```go
cfg, err := config.Load(*configName)
if err != nil {
    log.Fatal("failed to load config", zap.Error(err))
}
```

Python style would more often be:

```python
try:
    cfg = config.load(config_name)
except Exception as err:
    ...
```

In Go:

- success often means `err == nil`
- failure often means `err != nil`

This is one of the most important patterns to recognize.

---

# 5. `struct` ≈ lightweight Python class holding fields

Go:

```go
type Options struct {
    HotReload     bool
    CheckInterval float64
}
```

Python equivalent:

```python
@dataclass
class Options:
    hot_reload: bool
    check_interval: float
```

Go does not use Python-style classes as its central abstraction.

Instead, you combine:
- structs for data/state
- methods on structs
- interfaces for behavior/contracts

---

# 6. Methods use a receiver

OM1:

```go
func (rt *Runtime) Run(ctx context.Context) error {
    // ...
}
```

Read:

> "Run is a method on Runtime."

Python:

```python
class Runtime:
    def run(self, ctx):
        ...
```

The receiver:

```go
(rt *Runtime)
```

is analogous to Python's `self`.

The `*` means the method receives a pointer to the Runtime, so it can work with/mutate the actual object rather than a copy.

---

# 7. Pointer basics: `*` and `&`

You do not need C-level pointer expertise to start.

Think:

```go
*Runtime
```

means:

> reference/pointer to a Runtime object.

And:

```go
&Runtime{...}
```

means:

> create a Runtime value and give me its address/reference.

Python objects are already reference-like, so Python programmers rarely write this explicitly.

Example:

```go
return &Runtime{
    systemConfig: systemConfig,
    log: log,
}
```

Python mental model:

```python
return Runtime(
    system_config=system_config,
    log=log,
)
```

Do not get stuck on pointer syntax during your first read.

---

# 8. `interface` ≈ protocol / abstract behavior contract

OM1 defines:

```go
type Sensor interface {
    Listen(ctx context.Context) (<-chan any, error)
    Poll(ctx context.Context) (any, error)
    RawToText(ctx context.Context, rawInput any) (*Message, error)
    FormattedLatestBuffer() string
    Stop()
}
```

Python analogy:

```python
class Sensor(Protocol):
    def listen(self, ctx): ...
    def poll(self, ctx): ...
    def raw_to_text(self, ctx, raw_input): ...
    def formatted_latest_buffer(self) -> str: ...
    def stop(self): ...
```

Important Go idea:

A type does **not** need to explicitly say:

> "I implement Sensor."

If it has the required methods, it satisfies the interface automatically.

This is extremely important to the OM1 plugin design.

---

# 9. `any` is similar to Python's `Any`

```go
map[string]any
```

means roughly:

> dictionary/map whose keys are strings and whose values can be anything.

Python:

```python
dict[str, Any]
```

This pattern appears frequently when JSON/config/tool-call data is decoded.

---

# 10. Maps ≈ Python dictionaries

Go:

```go
registry := map[string]Factory{}
```

Python:

```python
registry: dict[str, Factory] = {}
```

Lookup:

```go
factory, ok := registry[typeName]
if !ok {
    // key did not exist
}
```

Python mental model:

```python
factory = registry.get(type_name)
if factory is None:
    ...
```

The second value, `ok`, is a common Go idiom.

---

# 11. Slices ≈ Python lists

```go
var calls []Call
```

Think:

```python
calls: list[Call] = []
```

Append:

```go
calls = append(calls, call)
```

Python:

```python
calls.append(call)
```

A Go array has fixed length; a slice is the dynamic list-like structure you will usually care about.

---

# 12. `range` ≈ Python `for ... in ...`

Go:

```go
for i, sensor := range o.sensors {
    // ...
}
```

Python:

```python
for i, sensor in enumerate(self.sensors):
    ...
```

Ignore an unused value with `_`:

```go
for _, action := range actions {
}
```

Python:

```python
for action in actions:
    ...
```

---

# 13. Goroutines = very lightweight concurrent functions

This syntax matters a lot in OM1:

```go
go rt.watchConfig(ctx)
```

Read:

> "Run watchConfig concurrently in a goroutine."

Another:

```go
go func() {
    rt.runCortexLoop(modeCtx)
}()
```

A goroutine is not exactly a Python thread, but for your first mental model:

> tiny cheap concurrent task managed by the Go runtime.

OM1 uses goroutines heavily because multiple things happen simultaneously:
- microphone listening
- cameras
- action connector heartbeats
- background jobs
- model loop
- mode transitions

---

# 14. Channels = typed communication pipes between goroutines

Example:

```go
tickNow chan struct{}
```

Think:

> a queue/signal channel through which goroutines communicate.

Send:

```go
o.tickNow <- struct{}{}
```

Receive:

```go
<-o.tickNow
```

Receive-only channel type:

```go
<-chan struct{}
```

OM1 uses channels for things such as:
- "new sensor input arrived"
- "this worker is done"
- "mode transition requested"

Python analogy is closer to `asyncio.Queue` / event signaling than a simple variable.

---

# 15. `select` waits on multiple concurrent events

Very important in OM1:

```go
select {
case <-ctx.Done():
    return
case <-timer.C:
case <-current.inputOrchestrator.TickNow():
}
```

Read:

> Wait until one of these things happens:
> - shutdown requested
> - timer fires
> - sensor asks us to run immediately

Python mental model:

> a concurrency-aware "wait for whichever event happens first."

---

# 16. `context.Context` = cancellation/deadline propagation

You will see `ctx context.Context` everywhere.

At a high level:

> Context carries shutdown/cancellation/deadline signals down a call tree.

Example:

```go
ctx, cancel := context.WithCancel(parent)
defer cancel()
```

When the robot process is shutting down, canceling a context lets many goroutines notice and exit cleanly.

This is one reason OM1 can coordinate many concurrent subsystems.

---

# 17. `defer` = run this when the current function exits

```go
defer cancel()
```

Python analogy:

```python
try:
    ...
finally:
    cancel()
```

Or context-manager cleanup.

Used for:
- closing resources
- unlocking
- canceling contexts
- flushing logs

---

# 18. `sync.Mutex` protects shared state

OM1:

```go
o.mu.Lock()
snapshot := ...
o.mu.Unlock()
```

Python mental model:

```python
with self.lock:
    snapshot = ...
```

This matters because multiple goroutines can touch the same object.

---

# 19. `sync.WaitGroup` waits for concurrent tasks

Pattern:

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    // work
}()

wg.Wait()
```

Meaning:

> launch work concurrently, but later block until all workers have finished.

---

# 20. Constructors are convention, not a language feature

Functions like:

```go
NewOrchestrator(...)
NewFuser(...)
NewMoveConnector(...)
```

are just ordinary functions following the Go convention:

> `NewX(...)` constructs an X.

Python analogy:

```python
x = X(...)
```

---

# 21. Registries: OM1's plugin trick

This is one of the most important patterns in the repo.

A package-level map:

```go
var registry = map[string]Factory{}
```

A plugin calls:

```go
inputs.Register("VLMGemini", NewVLMGemini)
```

Later, config says:

```json5
{
  type: "VLMGemini"
}
```

The runtime can then effectively do:

```go
factory := registry["VLMGemini"]
sensor := factory(config)
```

Python analogy:

```python
INPUT_REGISTRY["VLMGemini"] = NewVLMGemini
sensor = INPUT_REGISTRY[type_name](config)
```

This is how OM1 becomes configurable and modular.

---

# 22. `init()` runs automatically when a package is imported

Example:

```go
func init() {
    inputs.Register("VLMGemini", NewVLMGemini)
}
```

You do not manually call `init()`.

When the package loads, Go calls it automatically.

That leads directly to OM1's blank-import pattern.

---

# 23. Blank imports: strange-looking but crucial

In `cmd/main.go`:

```go
_ "github.com/openmind/om1/plugins/actions"
```

The underscore means:

> Import this package for its side effects, even though I never refer to its name.

Why?

Because loading that package triggers the nested plugin packages' `init()` functions, and those functions register their factories.

Mental model:

```text
blank import
   ↓
package loads
   ↓
init() runs
   ↓
Register("plugin-name", constructor)
   ↓
plugin registry is populated
```

This looks bizarre the first time you see it. In OM1 it is architectural, not accidental.

---

# 24. Type assertions

Example:

```go
args, ok := input.(map[string]any)
```

Read:

> "Try to treat input as a map[string]any."

If it works:
- `args` has that map
- `ok == true`

If not:
- `ok == false`

Python analogy:

```python
if isinstance(input, dict):
    args = input
```

---

# 25. Struct tags connect Go fields to JSON

```go
type MoveInput struct {
    Action MoveAction `json:"action" description:"The movement to perform"`
}
```

Those backtick strings are metadata.

They help serialization/schema generation know:
- JSON field name
- description

Python analogy: Pydantic/dataclass field metadata.

---

# 26. Exported vs private names

Go capitalization matters.

```go
type Runtime struct {}
func New(...) {}
```

Capitalized → exported from package.

```go
type modeState struct {}
func triggersTick(...) {}
```

lowercase → package-private.

This is Go's visibility system.

---

# 27. The five Go concepts to focus on first

For OM1, prioritize:

1. structs + methods
2. interfaces
3. maps/slices
4. goroutines + channels + select
5. error handling

Do **not** spend your first hour worrying about:
- advanced generics
- reflection internals
- unsafe pointers
- compiler implementation
- obscure syntax

---

# Quick translation exercise

Read this OM1 code:

```go
func init() {
    inputs.Register("VLMGemini", NewVLMGemini)
}

func NewVLMGemini(configMap map[string]any) (inputs.Sensor, error) {
    return NewCameraSensor("VLMGemini", geminiDefaults, configMap)
}
```

Translate it mentally into Python:

```python
def new_vlm_gemini(config: dict[str, Any]) -> Sensor:
    return NewCameraSensor("VLMGemini", gemini_defaults, config)

INPUT_REGISTRY["VLMGemini"] = new_vlm_gemini
```

That is not line-for-line language equivalence, but it is the right architectural intuition.

---

# Next file

Go to:

> [OM1 Code Map](02_OM1_CODE_MAP.md)

Then:

> [Annotated `cmd/main.go`](annotated/01_CMD_MAIN.md)
