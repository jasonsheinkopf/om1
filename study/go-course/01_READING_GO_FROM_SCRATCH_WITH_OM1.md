# Module 1 — Reading Go from Scratch with OM1

## What this lesson is for

You know Python. You know what programs do. But Go may still look like a dense wall of symbols:

```go
func Load(typeName string, cfg map[string]any) (Sensor, error) {
    f, ok := registry[typeName]
    if !ok {
        return nil, &UnknownPluginError{Kind: "input", Name: typeName}
    }
    return f(cfg)
}
```

By the end of this lesson, that should read almost like English.

You should be able to say:

> "Load takes a type name and a configuration dictionary. It looks up a factory in a registry. Go's map lookup gives both the value and a boolean saying whether the key existed. If it didn't, it returns nil plus an error. Otherwise it calls the factory and returns the Sensor plus any error."

That is the level of fluency we want.

---

# Part 1 — Before syntax: how a Go repository is organized

A Go source file normally ends in:

```text
.go
```

Examples in OM1:

```text
cmd/main.go
internal/fuser/fuser.go
internal/inputs/sensor.go
internal/actions/orchestrator.go
```

A directory normally corresponds to a **package**.

For example:

```text
internal/inputs/
    sensor.go
    orchestrator.go
    sensor_test.go
    orchestrator_test.go
```

Those files begin with:

```go
package inputs
```

The important idea:

> Several Go files in the same directory can collectively form one package.

This differs a little from the mental model many Python programmers use, where you often think in terms of a particular `.py` file as the module.

In Go, think first:

> "What package am I in?"

Then:

> "Which file happens to contain this particular definition?"

---

# Part 2 — `package main` and the program entry point

Open:

> `cmd/main.go`

At the top:

```go
package main
```

A Go executable uses a package named `main` and a function named:

```go
func main() {
    ...
}
```

The simplest possible Go program is:

```go
package main

import "fmt"

func main() {
    fmt.Println("hello")
}
```

Python mental model:

```python
print("hello")
```

or, in a more explicit script:

```python
def main():
    print("hello")

if __name__ == "__main__":
    main()
```

OM1's real process starts at:

> `cmd/main.go -> func main()`

You do not need to understand every line of `main` yet.

For now, recognize:

1. parse command-line arguments
2. create logger/metrics
3. load config
4. create cancellation context
5. construct runtime
6. call `rt.Run(ctx)`

That is the executable's startup sequence.

---

# Part 3 — imports

Go:

```go
import (
    "context"
    "time"

    "github.com/openmind/om1/internal/config"
)
```

Python analogy:

```python
import context_like_module
import time
from om1.internal import config
```

A few things are different.

## 3.1 Imports are strict

In normal Go code, an unused import is a compile error.

That sounds annoying at first, but it keeps files clean.

## 3.2 Package-qualified names

If you import:

```go
import "context"
```

you may see:

```go
context.Context
context.Background()
context.WithTimeout(...)
```

That is similar to:

```python
import context
context.Context
```

## 3.3 The strange underscore import

OM1's `cmd/main.go` includes imports like:

```go
_ "github.com/openmind/om1/plugins/actions"
```

The underscore says:

> "Load this package for its initialization side effects even though I will not refer to its package name directly."

Why would OM1 want that?

Because plugin packages register themselves when loaded.

We will study that deeply later. For now:

> blank import = load it so its setup code runs.

---

# Part 4 — variables: `var`, `:=`, and `=`

This is one of the first things that makes Go look unfamiliar.

There are several ways to create variables.

## 4.1 Explicit declaration

```go
var count int
```

This says:

> Create an integer variable named `count`.

You did not give it a value explicitly, so it receives its **zero value**.

For `int`, zero value is:

```text
0
```

For `bool`:

```text
false
```

For many reference-like types:

```text
nil
```

Go relies heavily on useful zero values.

---

## 4.2 Declaration with value

```go
var count int = 5
```

Go can infer the type:

```go
var count = 5
```

---

## 4.3 Short declaration: `:=`

Extremely common:

```go
count := 5
```

Read:

> "Create a new local variable called count and infer its type from the value."

Python:

```python
count = 5
```

The difference is that Python's syntax does both creation and reassignment.

Go distinguishes them.

---

## 4.4 Reassignment: `=`

Once the variable exists:

```go
count = 6
```

So remember:

```text
:=   declare + assign
=    assign
```

A useful reading trick:

When you see:

```go
cfg, err := config.Load(*configName)
```

say:

> "Create cfg and err from the two things returned by config.Load."

---

# Part 5 — basic types

You do not need to memorize every numeric type right away.

Common types you will see:

```go
string
bool
int
int64
float64
byte
```

Examples:

```go
name := "robot"
enabled := true
count := 4
rate := 10.0
```

Go is statically typed.

Once `count` is inferred as an integer, it is not suddenly a string later.

Python is dynamically typed, so this is a meaningful difference.

---

# Part 6 — functions

Python:

```python
def add(a: int, b: int) -> int:
    return a + b
```

Go:

```go
func add(a int, b int) int {
    return a + b
}
```

Read the Go signature left to right:

```text
func
add
(a int, b int)
int
```

means:

> Function named add, takes two integers, returns one integer.

Go often compresses parameters with the same type:

```go
func add(a, b int) int
```

Both mean the same thing.

---

# Part 7 — multiple return values

Go commonly returns more than one value.

Example:

```go
func lookup() (string, bool) {
    return "sensor", true
}
```

Caller:

```go
name, ok := lookup()
```

This pattern is everywhere.

Most importantly:

```go
result, err := doSomething()
```

Meaning:

> Give me the useful result and also an explicit error value.

This leads to one of Go's central patterns.

---

# Part 8 — errors are usually values

Python often says:

```python
try:
    cfg = load_config()
except Exception as err:
    handle(err)
```

Go commonly says:

```go
cfg, err := config.Load(name)
if err != nil {
    return err
}
```

Read it literally:

1. call function
2. receive result and error
3. if error is not nil, something failed
4. handle or return it

`nil` means roughly:

> no value / nothing / absent

It is not identical to Python `None`, but that is a useful first analogy.

The pattern:

```go
if err != nil {
    ...
}
```

is so common that your eyes should eventually recognize it instantly.

It means:

> failure path.

---

# Part 9 — the `if` statement

Go does not use parentheses around the condition:

```go
if count > 5 {
    fmt.Println("large")
}
```

Python:

```python
if count > 5:
    print("large")
```

Go uses braces instead of indentation to define the block.

---

# Part 10 — a special Go trick: initialization inside `if`

OM1's main function ends with:

```go
if err := rt.Run(ctx); err != nil && err != context.Canceled {
    log.Fatal(...)
}
```

This can look awful until you split it mentally.

It means approximately:

```go
err := rt.Run(ctx)
if err != nil && err != context.Canceled {
    ...
}
```

The variable declared before the semicolon is scoped to that `if`.

Python has no exact everyday equivalent.

Read it as:

> "Call Run. Call the returned error err. If that error is a real failure, handle it."

---

# Part 11 — loops: Go mainly uses `for`

Go does not need separate `while` syntax.

A classic loop:

```go
for i := 0; i < 10; i++ {
    ...
}
```

A while-style loop:

```go
for running {
    ...
}
```

An infinite loop:

```go
for {
    ...
}
```

Iterating over a collection:

```go
for i, sensor := range sensors {
    ...
}
```

Python:

```python
for i, sensor in enumerate(sensors):
    ...
```

If you do not care about one of the values:

```go
for _, sensor := range sensors {
    ...
}
```

The underscore means:

> intentionally ignore this value.

---

# Part 12 — structs: the closest common analogy is a data-oriented class

Open:

> `internal/inputs/sensor.go`

You find:

```go
type Message struct {
    Timestamp float64
    Message   string
}
```

Read:

> Define a new type called Message. A Message has a Timestamp and a Message string.

Python analogy:

```python
from dataclasses import dataclass

@dataclass
class Message:
    timestamp: float
    message: str
```

Important:

Go does **not** organize everything around classes.

A common Go pattern is:

- struct = state/data
- methods = behavior attached to a type
- interface = required behavior
- composition = combine things

That is why trying to translate every Go type into "class" can eventually become misleading.

But at the beginning:

> struct ≈ data-holding class

is useful.

---

# Part 13 — constructing a struct

Example:

```go
m := Message{
    Timestamp: 10.5,
    Message:   "hello",
}
```

Python:

```python
m = Message(
    timestamp=10.5,
    message="hello",
)
```

OM1 uses:

```go
return &Message{
    Timestamp: float64(time.Now().UnixNano()) / 1e9,
    Message:   text,
}
```

There is one extra symbol:

```text
&
```

We will address that next.

---

# Part 14 — pointers, without making them scary

Python programmers often find Go pointers intimidating because Python hides references.

Start with this mental model:

> A pointer lets code refer to the original value rather than an independent copy.

A type:

```go
Message
```

means:

> a Message value.

A type:

```go
*Message
```

means:

> a pointer/reference to a Message value.

This:

```go
&Message{...}
```

means:

> build a Message, then give me a pointer to it.

Python objects already feel reference-like, so this:

```go
return &Message{Message: text}
```

can initially be read approximately as:

```python
return Message(message=text)
```

Do not try to mentally simulate RAM addresses yet.

The useful question is:

> "Is this function passing around a value or a reference to shared/mutable state?"

---

# Part 15 — methods and receivers

Python:

```python
class Robot:
    def stop(self):
        ...
```

Go:

```go
type Robot struct {
    ...
}

func (r *Robot) Stop() {
    ...
}
```

The part:

```go
(r *Robot)
```

is called the **receiver**.

Read:

> "Stop is a method on pointer-to-Robot."

`r` serves a role similar to Python's `self`.

For OM1:

```go
func (o *Orchestrator) Buffers() []string {
    ...
}
```

Read:

> "Buffers is a method on Orchestrator. It returns a slice of strings."

The receiver variable is:

```text
o
```

The receiver type is:

```text
*Orchestrator
```

---

# Part 16 — exported versus unexported names

Go uses capitalization to control package visibility.

Capitalized:

```go
Message
NewMessage
Sensor
Load
```

These are **exported** from the package.

Lowercase:

```go
registry
triggersTick
```

These are package-private/unexported.

That means the first letter has semantic meaning.

Python analogy:

Python often uses underscore conventions:

```python
_internal_helper()
```

but Go enforces the exported/unexported distinction through capitalization.

---

# Part 17 — slices: the list-like structure you will usually care about

Python list:

```python
sensors = []
```

Go slice:

```go
var sensors []Sensor
```

A slice is a dynamic view over an underlying array.

For your first pass:

> slice ≈ Python list

is good enough.

Create one:

```go
names := []string{"lidar", "camera", "imu"}
```

Append:

```go
names = append(names, "microphone")
```

Get length:

```go
len(names)
```

Index:

```go
names[0]
```

OM1 example:

```go
func (o *Orchestrator) Buffers() []string
```

means:

> Buffers returns a slice/list of strings.

---

# Part 18 — arrays versus slices

Go arrays have fixed length in their type:

```go
[3]string
```

A slice:

```go
[]string
```

does not put a fixed length in the type.

In application code like OM1, you will usually be thinking about slices.

If you see:

```go
[]Thing
```

say:

> "a sequence/list of Thing."

---

# Part 19 — `make`

Go often uses:

```go
snapshot := make([]string, len(o.sensors))
```

Read:

> Create a string slice whose length equals the number of sensors.

Python approximation:

```python
snapshot = [""] * len(self.sensors)
```

`make` is used for certain built-in reference-like types such as:

- slices
- maps
- channels

Do not think of it as the universal constructor.

---

# Part 20 — maps: dictionaries

Python:

```python
registry = {}
```

Go:

```go
registry := map[string]Factory{}
```

Read the type:

```text
map[string]Factory
```

as:

> map from string keys to Factory values.

Python typing:

```python
dict[str, Factory]
```

OM1 actually declares:

```go
var registry = map[string]Factory{}
```

This is the plugin registry for input sensors.

---

# Part 21 — map lookup and the famous `value, ok` pattern

Go:

```go
f, ok := registry[typeName]
```

This returns:

1. `f`: the value
2. `ok`: whether the key existed

Then:

```go
if !ok {
    ...
}
```

`!` means NOT.

So:

```text
!ok
```

means:

> not okay / key not found.

Python:

```python
if type_name not in registry:
    ...
f = registry[type_name]
```

or:

```python
f = registry.get(type_name)
if f is None:
    ...
```

Go's two-result lookup is explicit and avoids ambiguity.

---

# Part 22 — `any`

OM1 frequently uses:

```go
map[string]any
```

`any` means:

> a value of any type.

Python typing:

```python
dict[str, Any]
```

This is particularly useful around:

- decoded JSON
- configuration
- LLM tool-call arguments
- generic sensor values

Static typing has not disappeared. Instead, Go is saying:

> this particular value intentionally has a dynamic/unknown concrete type.

Later code may inspect or assert what concrete type it contains.

---

# Part 23 — interfaces: one of the most important Go concepts in OM1

Here is the real OM1 interface:

```go
type Sensor interface {
    Listen(ctx context.Context) (<-chan any, error)
    Poll(ctx context.Context) (any, error)
    RawToText(ctx context.Context, rawInput any) (*Message, error)
    FormattedLatestBuffer() string
    Stop()
}
```

Do not try to understand the channel syntax yet.

Focus on the shape.

It says:

> Anything that wants to behave as a Sensor must provide these methods.

Python analogy:

```python
class Sensor(Protocol):
    def listen(self, ctx): ...
    def poll(self, ctx): ...
    def raw_to_text(self, ctx, raw_input): ...
    def formatted_latest_buffer(self) -> str: ...
    def stop(self): ...
```

The very important Go difference:

> A type does not normally write "implements Sensor."

If it has the required methods, it satisfies the interface automatically.

This is called **implicit interface satisfaction**.

---

# Part 24 — why interfaces matter in OM1

OM1 needs different sensor implementations:

- speech input
- camera/VLM input
- localization input
- robot odometry
- presence detection
- and others

The runtime should not have to know every implementation.

Instead it knows:

> "Give me anything that satisfies Sensor."

Then runtime code can call:

```go
sensor.Listen(...)
sensor.FormattedLatestBuffer()
sensor.Stop()
```

without caring whether the concrete object talks to:

- a microphone
- a camera
- Zenoh
- a cloud API
- a robot

This is dependency inversion in a very practical form.

The runtime depends on an abstraction.

Plugins provide the concrete behavior.

---

# Part 25 — function types

OM1 declares:

```go
type Factory func(cfg map[string]any) (Sensor, error)
```

This is easy to miss.

It says:

> Factory is a type representing a function with this exact signature.

In Python typing:

```python
Factory = Callable[[dict[str, Any]], tuple[Sensor, Exception | None]]
```

So when OM1 declares:

```go
map[string]Factory
```

it means:

> a dictionary from a string plugin name to a constructor-like function.

This is the backbone of the registry.

---

# Part 26 — the plugin registry, step by step

The real code:

```go
var registry = map[string]Factory{}
```

Start empty.

A plugin registers itself:

```go
func Register(typeName string, f Factory) {
    registry[typeName] = f
}
```

Conceptually:

```python
registry[type_name] = factory_function
```

Later, config contains a plugin type name.

OM1 loads it:

```go
func Load(typeName string, cfg map[string]any) (Sensor, error) {
    f, ok := registry[typeName]
    if !ok {
        return nil, &UnknownPluginError{Kind: "input", Name: typeName}
    }
    return f(cfg)
}
```

This is worth translating line by line.

### Function signature

```go
func Load(typeName string, cfg map[string]any) (Sensor, error)
```

Meaning:

> Load takes a string and a config dictionary. It returns a Sensor and an error.

### Lookup

```go
f, ok := registry[typeName]
```

Meaning:

> Find the factory registered under this type name, and tell me whether it existed.

### Failure

```go
if !ok {
    return nil, &UnknownPluginError{Kind: "input", Name: typeName}
}
```

Meaning:

> If the plugin was not registered, return no Sensor and a descriptive error.

### Success

```go
return f(cfg)
```

Meaning:

> Call the factory using the configuration. Whatever Sensor/error pair the factory returns becomes Load's return value.

Python-like pseudocode:

```python
def load(type_name, cfg):
    factory = registry.get(type_name)
    if factory is None:
        return None, UnknownPluginError(kind="input", name=type_name)

    return factory(cfg)
```

If this makes sense, you already understand an important chunk of OM1's extensibility.

---

# Part 27 — custom error types

OM1 defines:

```go
type UnknownPluginError struct {
    Kind string
    Name string
}
```

Then:

```go
func (e *UnknownPluginError) Error() string {
    return e.Kind + " plugin not found: " + e.Name
}
```

This is a beautiful Go interface example.

The standard `error` interface is conceptually:

```go
type error interface {
    Error() string
}
```

`UnknownPluginError` has an `Error() string` method.

Therefore it satisfies `error`.

It never had to declare:

```text
implements error
```

Go just sees:

> method set matches interface.

This is structural typing.

---

# Part 28 — struct literal plus pointer in an error

This line:

```go
&UnknownPluginError{Kind: "input", Name: typeName}
```

means:

1. construct an UnknownPluginError
2. set Kind
3. set Name
4. take its address
5. use that pointer as an error value

Python approximation:

```python
UnknownPluginError(kind="input", name=type_name)
```

Again: Python hides the explicit pointer part.

---

# Part 29 — type assertions

OM1 has:

```go
t, ok := sensor.(TickTrigger)
```

This looks strange.

`sensor` has interface type `Sensor`.

The code is asking:

> Does the concrete object inside this Sensor ALSO satisfy the TickTrigger interface?

If yes:

- `t` is that value viewed as a TickTrigger
- `ok` is true

If not:

- `ok` is false

Then:

```go
return ok && t.TriggersTick()
```

This is a powerful pattern:

> mandatory interface + optional capability interface

Every Sensor must satisfy `Sensor`.

Only some sensors need to satisfy:

```go
type TickTrigger interface {
    TriggersTick() bool
}
```

Python mental model:

```python
if isinstance(sensor, TickTrigger):
    return sensor.triggers_tick()
return False
```

or more duck-typed:

```python
if hasattr(sensor, "triggers_tick"):
    return sensor.triggers_tick()
return False
```

The Go version keeps it type-safe.

---

# Part 30 — `nil`

You will see:

```go
return nil, err
```

or:

```go
if message == nil {
    ...
}
```

For now, think:

> nil = no object/value here

But note:

Not every Go type can be nil.

For example, a plain `int` is not nil.

Pointers, interfaces, maps, slices, channels, and functions can be nil.

We will revisit tricky interface-nil cases later.

---

# Part 31 — `defer`

OM1's `main` contains patterns like:

```go
defer cancel()
```

`defer` means:

> Run this call when the current function exits.

Python analogy:

```python
try:
    ...
finally:
    cancel()
```

or a context manager.

Another pattern:

```go
defer wg.Done()
```

means:

> no matter how this goroutine exits, tell the WaitGroup that this worker is finished.

This becomes extremely useful in concurrent code.

---

# Part 32 — naming patterns that help you read Go

Go code uses conventions heavily.

## `NewX`

```go
NewMessage(...)
NewOrchestrator(...)
NewFuser(...)
```

Usually means:

> constructor-like function that creates X.

Go has no special constructor syntax.

`NewX` is just convention.

## `Load`

Usually:

> construct/read something from config/storage/registry.

## `Start`

Often:

> begin a long-running subsystem.

## `Stop`

Often:

> terminate/release resources.

## `Run`

Often:

> own a lifecycle/event loop until shutdown or failure.

These are conventions, not compiler rules.

---

# Part 33 — reading a complete OM1 file: `internal/inputs/sensor.go`

Now read it as architecture rather than syntax.

The file contains five conceptual pieces.

## Piece 1 — normalized processed message

```go
type Message struct {
    Timestamp float64
    Message   string
}
```

Meaning:

> once raw input has been processed, OM1 can represent it as timestamped text.

Then:

```go
func NewMessage(text string) *Message
```

is a constructor-like helper.

---

## Piece 2 — Sensor interface

```go
type Sensor interface {
    ...
}
```

Meaning:

> this is the contract that all input sensor plugins must satisfy.

The central runtime can work with many concrete inputs because it depends on this interface.

---

## Piece 3 — optional TickTrigger capability

```go
type TickTrigger interface {
    TriggersTick() bool
}
```

Meaning:

> some sensors have an extra capability: they can declare whether new data should wake the cortex loop immediately.

Not all Sensors must do this.

---

## Piece 4 — factory type and registry

```go
type Factory func(cfg map[string]any) (Sensor, error)
var registry = map[string]Factory{}
```

Meaning:

> plugin name → function capable of building that sensor.

---

## Piece 5 — Register and Load

```go
func Register(...)
func Load(...)
```

Meaning:

> plugins register constructor functions, and later the runtime asks for a sensor by type name.

That is the entire conceptual file.

It is not "lots of Go magic."

It is:

> message model + sensor contract + optional capability + constructor registry + loader.

---

# Part 34 — now connect `sensor.go` to `orchestrator.go`

Open:

> `internal/inputs/orchestrator.go`

The struct:

```go
type Orchestrator struct {
    sensors []Sensor
    tickNow chan struct{}
    log     *zap.Logger
}
```

Even before learning channels, you can read:

- `sensors []Sensor` = list of Sensor interface values
- `tickNow ...` = some signaling mechanism
- `log *zap.Logger` = pointer to logger

The constructor:

```go
func NewOrchestrator(sensors []Sensor, log *zap.Logger) *Orchestrator
```

means:

> Give me a list of sensors and a logger; I will return a pointer to a new Orchestrator.

Already readable.

---

# Part 35 — first glimpse of concurrency

Inside `Start`:

```go
for i, sensor := range o.sensors {
    wg.Add(1)
    go func(sensorIndex int, sensor Sensor) {
        defer wg.Done()
        o.runSensor(ctx, sensorIndex, sensor)
    }(i, sensor)
}
```

Do not worry about mastering this yet.

At a conceptual level:

> Start one concurrent worker per sensor.

That makes sense architecturally.

The microphone, camera, localization, and robot state should not have to wait for each other in one sequential loop.

The syntax will get a dedicated module later.

---

# Part 36 — a crucial distinction: value, type, function, method, interface

When Go feels confusing, classify the name.

Take:

```go
Sensor
```

Could be:
> type/interface

Take:

```go
sensor
```

Usually:
> variable/value

Take:

```go
Load
```

Here:
> function

Take:

```go
o.Buffers()
```

Here:
> method call on a value

Take:

```go
Factory
```

Here:
> named function type

A lot of Go comprehension becomes easier once you stop treating every capitalized identifier like a "class."

---

# Part 37 — the dot means "from this package/value"

Examples:

```go
config.Load(...)
```

means:

> Load from the config package.

```go
rt.Run(ctx)
```

means:

> call method Run on rt.

```go
time.Now()
```

means:

> call Now from the time package.

Same punctuation, different left-hand thing.

Ask:

> Is the left side a package name or a variable?

That tells you how to interpret it.

---

# Part 38 — dereferencing in `cmd/main.go`

OM1 uses the standard `flag` package:

```go
configName = flag.String(...)
```

`flag.String` returns a:

```go
*string
```

a pointer to string.

So later:

```go
*configName
```

means:

> get the actual string value pointed to by configName.

This is one place where explicit pointer syntax matters.

Do not generalize that every `*` means the same syntactic role.

Depending on context:

```go
*Runtime
```

in a type means:
> pointer to Runtime

while:

```go
*configName
```

in an expression means:
> dereference the pointer and get the value

Context tells you which meaning applies.

---

# Part 39 — a miniature OM1-like plugin system from scratch

Here is a tiny version.

```go
package main

import "fmt"

type Sensor interface {
    Read() string
}

type Camera struct {
    Name string
}

func (c *Camera) Read() string {
    return "image from " + c.Name
}

type Factory func() Sensor

var registry = map[string]Factory{}

func Register(name string, factory Factory) {
    registry[name] = factory
}

func Load(name string) (Sensor, error) {
    factory, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown sensor %q", name)
    }
    return factory(), nil
}

func main() {
    Register("camera", func() Sensor {
        return &Camera{Name: "front"}
    })

    sensor, err := Load("camera")
    if err != nil {
        panic(err)
    }

    fmt.Println(sensor.Read())
}
```

Translate this to Python mentally.

The architecture is basically:

```python
class Sensor(Protocol):
    def read(self) -> str: ...

class Camera:
    def __init__(self, name):
        self.name = name

    def read(self):
        return "image from " + self.name

registry = {}

def register(name, factory):
    registry[name] = factory

def load(name):
    if name not in registry:
        raise ValueError(...)
    return registry[name]()

register("camera", lambda: Camera("front"))
sensor = load("camera")
print(sensor.read())
```

The Go version is more explicit about types and failures.

But the core design is familiar.

---

# Part 40 — why OM1 feels "more complicated" than the tiny example

Because real OM1 adds:

- configuration
- contexts/cancellation
- concurrency
- logging
- multiple plugins
- tool schemas
- robot middleware
- LLM providers
- memory
- knowledge retrieval
- mode transitions
- lifecycle management

But the language building blocks remain the same.

You should repeatedly ask:

> "What simple pattern is hiding under the production detail?"

That is the correct way to read a mature repository.

---

# Part 41 — read this real function without panicking

From OM1:

```go
func (o *Orchestrator) Buffers() []string {
    snapshot := make([]string, len(o.sensors))

    for i, sensor := range o.sensors {
        snapshot[i] = sensor.FormattedLatestBuffer()
    }

    return snapshot
}
```

Translate it.

### Signature

```go
func (o *Orchestrator) Buffers() []string
```

> method named Buffers on pointer-to-Orchestrator; returns a list/slice of strings.

### Allocate

```go
snapshot := make([]string, len(o.sensors))
```

> make one output slot for every sensor.

### Loop

```go
for i, sensor := range o.sensors
```

> iterate over sensor list with index and value.

### Call interface method

```go
sensor.FormattedLatestBuffer()
```

> ask each concrete sensor for its standardized text representation.

### Store

```go
snapshot[i] = ...
```

> place it in the corresponding output slot.

### Return

```go
return snapshot
```

> return all current sensor buffer strings.

You do not need advanced Go knowledge to understand this function.

---

# Part 42 — now connect it to the Fuser

The runtime calls:

```go
sensorBuffers := current.inputOrchestrator.Buffers()
```

Then:

```go
prompt, err := current.promptFuser.Fuse(ctx, sensorBuffers)
```

So:

1. every sensor returns a formatted textual buffer
2. Buffers gathers them into `[]string`
3. the Fuser receives that slice
4. the Fuser writes non-empty observations into the prompt

This is why understanding:

```go
[]string
```

matters architecturally.

It is not syntax trivia.

That slice is carrying the robot/agent's current observations into the context-building stage.

---

# Part 43 — what you should be able to say out loud now

Try this without looking back:

> "In Go, a struct holds state, methods are functions with receivers, and interfaces define behavior. OM1 defines Sensor as an interface so different input plugins can implement the same contract. It keeps constructor functions in a registry map. The config provides a type name, Load finds the factory, builds a Sensor, and the input orchestrator works with the Sensor interface instead of a particular implementation."

If that sentence makes sense, this lesson is working.

---

# Part 44 — terminology checkpoint

Make sure these terms are no longer fuzzy.

**package**  
A namespace/unit of Go code, usually corresponding to a directory.

**function**  
A callable defined with `func`.

**method**  
A function with a receiver, attached to a type.

**receiver**  
The value before the method name in a method definition, roughly analogous to Python `self`.

**struct**  
A user-defined type containing fields.

**pointer**  
A value referring to another value's memory location; practically, a way to operate on shared/original state rather than just a copy.

**interface**  
A set of required methods: a behavioral contract.

**slice**  
A dynamic sequence; closest first analogy is Python list.

**map**  
Key/value mapping; closest analogy is Python dict.

**factory**  
A function that creates/returns another object/value.

**registry**  
A mapping from names to implementations/factories.

**error**  
A normal Go value representing failure; often returned as the final return value.

**nil**  
Absence/no value for certain Go types.

**type assertion**  
Check/extract the concrete type or another interface capability from an interface value.

**zero value**  
Default value a variable receives when declared without explicit initialization.

**exported**  
Capitalized package-level name accessible from other packages.

---

# Part 45 — what we intentionally have NOT mastered yet

You have seen these but should not expect mastery yet:

- channels
- `<-chan`
- goroutines
- `select`
- context cancellation
- mutexes
- WaitGroups
- generic concurrency safety
- JSON/tool-call schemas
- Go testing
- modules and dependency management
- method sets and subtle pointer/interface rules

Those get their own modules.

Do not judge your understanding of Go based on those yet.

---

# Part 46 — one final reading exercise before the exercise file

Read this:

```go
type Factory func(cfg map[string]any) (Sensor, error)

var registry = map[string]Factory{}

func Register(typeName string, f Factory) {
    registry[typeName] = f
}
```

You should now be able to say:

> "Factory is not a class. It is a named function type. It describes constructor functions that take a configuration dictionary and return a Sensor plus an error. The registry is a dictionary from plugin name to one of those factory functions. Register inserts a factory into the registry."

If you can say that comfortably, continue.

---

# Module 1 finish line

Go back and read:

> `internal/inputs/sensor.go`

Do it slowly.

For every unfamiliar line, label the construct:

- type?
- variable?
- function?
- method?
- pointer?
- interface?
- map?
- function type?
- return values?
- error?

Then complete:

> **[EXERCISES_01.md](EXERCISES_01.md)**

Do not move to the concurrency module until `sensor.go` feels mostly readable.
