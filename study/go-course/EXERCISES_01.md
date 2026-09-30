# Module 1 Exercises — Go Reading with OM1

These are designed to work offline.

Do the questions first. The answer key is at the bottom.

---

# Exercise 1 — Translate the signature

Given:

```go
func NewMessage(text string) *Message
```

Answer in plain English:

1. Is this a function or a method?
2. What argument does it take?
3. What does it return?
4. What does the `*` mean here?
5. Why is the name capitalized?

---

# Exercise 2 — Declaration versus assignment

Explain the difference:

```go
x := 5
x = 6
```

What would this mean in Python?

---

# Exercise 3 — Multiple returns

Given:

```go
sensor, err := Load("camera", cfg)
```

Explain:

1. How many values does `Load` return?
2. Why does Go often return `err` explicitly?
3. What does success usually mean for `err`?

---

# Exercise 4 — Read the map type

Explain:

```go
map[string]Factory
```

Then explain:

```go
map[string]any
```

What are the closest Python typing equivalents?

---

# Exercise 5 — The comma-ok lookup

Given:

```go
f, ok := registry[typeName]
if !ok {
    return nil, err
}
```

What is `ok` telling you?

What does `!ok` mean?

---

# Exercise 6 — Struct

Given:

```go
type Result struct {
    ActionName string
    Err        error
}
```

Explain this to a Python developer.

---

# Exercise 7 — Struct literal

Explain:

```go
Result{
    ActionName: "move",
    Err:        nil,
}
```

What Python construct is this most similar to?

---

# Exercise 8 — Method receiver

Given:

```go
func (o *Orchestrator) Buffers() []string
```

Explain every major piece:

- `o`
- `*Orchestrator`
- `Buffers`
- `()`
- `[]string`

---

# Exercise 9 — Interface

Given:

```go
type Speaker interface {
    Speak(text string) error
    Stop()
}
```

Does a concrete type need to declare `implements Speaker`?

What must it do to satisfy the interface?

---

# Exercise 10 — Optional interface capability

Imagine:

```go
type Sensor interface {
    Read() string
}

type Resettable interface {
    Reset()
}
```

Why might this be useful?

Can something be a Sensor but not Resettable?

---

# Exercise 11 — Type assertion

Explain:

```go
r, ok := sensor.(Resettable)
```

What does `ok` mean?

What does `r` become if the check succeeds?

---

# Exercise 12 — Factory

Explain:

```go
type Factory func(cfg map[string]any) (Sensor, error)
```

Is Factory:

- a struct?
- an interface?
- a function?
- a named function type?

Why is this useful in a plugin registry?

---

# Exercise 13 — Read real OM1 registry code

Without looking at the lesson, explain this:

```go
var registry = map[string]Factory{}

func Register(typeName string, f Factory) {
    registry[typeName] = f
}
```

Use the words:

- map
- plugin name
- factory
- constructor-like function

---

# Exercise 14 — Read real OM1 Load

Explain this line by line:

```go
func Load(typeName string, cfg map[string]any) (Sensor, error) {
    f, ok := registry[typeName]
    if !ok {
        return nil, &UnknownPluginError{Kind: "input", Name: typeName}
    }
    return f(cfg)
}
```

Do not merely paraphrase symbols. Explain the purpose.

---

# Exercise 15 — Why no `implements`?

OM1's `UnknownPluginError` has:

```go
func (e *UnknownPluginError) Error() string
```

Why can it be returned where Go expects an `error`?

---

# Exercise 16 — Slice reading

Explain:

```go
sensors []Sensor
```

Does this mean:

- one Sensor?
- fixed array?
- dynamic sequence/slice of values satisfying Sensor?

---

# Exercise 17 — `make`

Explain:

```go
snapshot := make([]string, len(o.sensors))
```

Why might OM1 create exactly that many elements?

---

# Exercise 18 — Loop

Translate:

```go
for i, sensor := range o.sensors {
    snapshot[i] = sensor.FormattedLatestBuffer()
}
```

into Python-like pseudocode.

---

# Exercise 19 — Architecture from syntax

This matters more than memorizing language terms.

Given:

```go
sensorBuffers := current.inputOrchestrator.Buffers()
prompt, err := current.promptFuser.Fuse(ctx, sensorBuffers)
```

Explain the architectural transition happening here.

What is being converted from what into what?

---

# Exercise 20 — Explain the plugin architecture aloud

Without code, explain:

1. what a Sensor interface is
2. what a Factory is
3. what the registry stores
4. what Register does
5. what Load does
6. why the central runtime benefits from this

If you can do this smoothly, you understand more than syntax. You understand design.

---

# Exercise 21 — Spot the category

Classify each identifier:

```text
Message
NewMessage
Sensor
sensor
Factory
registry
Register
Load
UnknownPluginError
Error
```

Possible categories:

- struct type
- interface type
- variable
- function
- method
- named function type
- map variable
- custom error struct

Some may reasonably receive more than one descriptive label.

---

# Exercise 22 — Python translation

Translate this conceptually into Python:

```go
type Camera struct {
    Name string
}

func (c *Camera) Read() string {
    return c.Name
}
```

Then answer:

Why is `Read` called a method instead of a plain function?

---

# Exercise 23 — Export rules

Which are exported from their package?

```text
Message
message
Load
load
registry
Register
triggersTick
TickTrigger
```

What is the rule?

---

# Exercise 24 — Pointer reading

Explain the two different uses of `*`:

```go
func (rt *Runtime) Run(...)
```

and:

```go
name := *configName
```

---

# Exercise 25 — Interview explanation

Pretend the interviewer asks:

> "I see you come from Python. What have you learned about Go from reading OM1?"

Give a 30-second answer that mentions:

- structs + methods instead of Python-style classes
- interfaces
- explicit errors
- slices/maps
- concurrency as an important next area

Do not claim mastery you do not have.

---

# Answer key

## 1

`NewMessage` is a plain function because there is no receiver before its name. It takes one string and returns a pointer to Message. The pointer means the caller receives a reference to the created Message rather than a copied Message value. The capitalization makes it exported from the package.

## 2

`:=` declares and assigns a new local variable. `=` changes an existing variable. Python uses `=` for both common cases.

## 3

Two values: a Sensor and an error. Go usually represents failures explicitly as returned values. Success usually means `err == nil`.

## 4

`map[string]Factory` is a map/dictionary from strings to Factory values. Rough Python: `dict[str, Factory]`.

`map[string]any` is a string-keyed map whose values may hold arbitrary concrete types. Rough Python: `dict[str, Any]`.

## 5

`ok` tells you whether the requested key existed in the map. `!ok` means it did not.

## 6

A Result is a struct type with two fields: ActionName of type string and Err of type error. A Python dataclass is a useful analogy.

## 7

It constructs a Result value and names the field values. Roughly like a dataclass constructor with keyword arguments.

## 8

`o` is the receiver variable, similar in role to `self`. `*Orchestrator` is a pointer to an Orchestrator. `Buffers` is the method name. Empty parentheses mean no explicit arguments. `[]string` means it returns a slice of strings.

## 9

No explicit `implements` declaration is required. The concrete type must provide `Speak(string) error` and `Stop()` with compatible method signatures.

## 10

It allows Reset to be optional. Every value can satisfy Sensor without also supporting Resettable.

## 11

The code asks whether the concrete value inside `sensor` also satisfies Resettable. If yes, `ok` is true and `r` can be used as a Resettable.

## 12

Factory is a named function type. It defines the signature every sensor-construction function must have, so those functions can be stored uniformly in a map.

## 13

The registry maps plugin names to factory functions. Register inserts a constructor-like factory under the name used later by configuration/loading.

## 14

Load receives a plugin name plus config. It looks up the factory. If none exists, it returns no Sensor and a custom error. If the factory exists, Load calls it with the configuration and returns its Sensor/error result.

## 15

The built-in error contract requires an `Error() string` method. UnknownPluginError has that method, so the pointer satisfies the error interface implicitly.

## 16

A slice/dynamic sequence of values that satisfy Sensor.

## 17

It creates one output string slot for every configured sensor so the code can fill `snapshot[i]` as it loops.

## 18

Roughly:

```python
for i, sensor in enumerate(self.sensors):
    snapshot[i] = sensor.formatted_latest_buffer()
```

## 19

The input orchestrator snapshots each sensor's latest formatted state into a slice of strings. The Fuser then consumes those observation strings as one component of the LLM prompt/context.

## 20

A strong answer:

> Sensor defines the behavior every input plugin must provide. A Factory is a constructor-like function that builds one of those Sensors from configuration. OM1 keeps factories in a registry keyed by plugin type name. Plugins register their factories, then the runtime can Load whichever implementation the config names. That lets the core runtime depend on the Sensor abstraction instead of hard-coding every microphone, camera, localization source, or robot input.

## 21

- Message: struct type
- NewMessage: function
- Sensor: interface type
- sensor: variable
- Factory: named function type
- registry: map variable
- Register: function
- Load: function
- UnknownPluginError: struct/custom error type
- Error: method

## 22

Python-like:

```python
class Camera:
    def __init__(self, name: str):
        self.name = name

    def read(self) -> str:
        return self.name
```

Read is a method because its Go declaration has the receiver `(c *Camera)`.

## 23

Exported:

- Message
- Load
- Register
- TickTrigger

Rule: identifiers beginning with uppercase are exported from the package.

## 24

In `*Runtime`, the star is part of a type and means pointer to Runtime.

In `*configName`, the star is an expression and dereferences a pointer to retrieve the pointed-to value.

## 25

Example answer:

> "My strongest language is Python, so I've been mapping Go constructs to concepts I already use. In OM1 I'm getting comfortable with structs plus receiver methods instead of Python-style classes, interfaces as implicit behavioral contracts, explicit error returns, and slices and maps. I'm also starting to trace the plugin registries. The biggest Go-specific area I'm deliberately building next is concurrency—goroutines, channels, context cancellation, and synchronization—because OM1 uses those heavily."

---

# Score yourself

**22–25 comfortable:** move on.

**17–21 mostly there:** reread the parts tied to missed questions, then move on.

**12–16 partial:** reread `internal/inputs/sensor.go` with the lesson beside it.

**Below 12:** repeat Module 1 before adding concurrency. The point is not speed.

The goal is that the syntax becomes boring.

That is when you are ready to focus on architecture.
