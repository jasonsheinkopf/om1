# Go Through OM1 — Deep Course for a Python Developer

This course is designed for a Python-fluent engineer who wants to become comfortable enough with Go to **read, explain, debug, and make small changes in OM1**.

It is deliberately not a generic "learn Go in 20 minutes" guide. The goal is to make the syntax and design patterns feel unsurprising by repeatedly connecting them to:

1. Python concepts you already know.
2. Concrete OM1 files.
3. The runtime behavior of a robot/agent system.
4. Interview-style explanations.

The course is intended to work **offline** after you pull the repository.

---

## What "done" should feel like

After roughly five focused hours, you should be able to open a typical OM1 Go file and answer questions like:

- What package am I in?
- What state does this struct hold?
- Which functions are plain functions and which are methods?
- Why is this receiver a pointer?
- What interface does this type satisfy?
- What is this map or slice storing?
- What does `:=` create here?
- Why are there two return values?
- What does `if err != nil` mean?
- What is a type assertion?
- Why is a goroutine being started?
- What is flowing through this channel?
- What can cancel this code?
- Why is there a `select`?
- What state needs a mutex?
- What does this JSON-like tool call become?
- How does an LLM action reach a robot connector?
- Where would I put a breakpoint/log statement if the system stopped behaving correctly?

You do **not** need to become a language-lawyer or expert Go developer. The target is practical fluency.

---

# The mental model: learn Go by following one OM1 loop

Keep this runtime story in your head during the entire course:

```text
config
  ↓
cmd/main.go
  ↓
runtime.New / Runtime.Run
  ↓
input plugins
  ↓
input orchestrator
  ↓
sensor buffers
  ↓
Fuser
  ↓
prompt
  ↓
Cortex LLM
  ↓
structured tool/action calls
  ↓
action orchestrator
  ↓
connector
  ↓
Zenoh / HTTP / TTS / robot interface
  ↓
physical or software action
```

Go concepts are not learned in isolation. They are learned where they appear in this path.

---

# Five-hour course plan

The times are approximate. Do not rush to match them.

## Module 1 — Reading Go from scratch using OM1
**60–90 minutes**

File:

> `study/go-course/01_READING_GO_FROM_SCRATCH_WITH_OM1.md`

Starts from almost zero Go knowledge.

Topics:

- source files and packages
- imports
- `main`
- variables
- `var`, `:=`, and `=`
- primitive types
- functions and return types
- multiple returns
- errors as values
- `if`
- `for`
- structs
- struct literals
- pointers
- methods and receivers
- slices
- maps
- `any`
- interfaces
- implicit interface satisfaction
- function types
- registries
- type assertions
- `nil`
- exported versus unexported names
- `defer`
- a first complete reading of `internal/inputs/sensor.go`

This module deliberately repeats important ideas several times.

Exercise companion:

> `study/go-course/EXERCISES_01.md`

---

## Module 2 — Go's object model: structs, methods, pointers, interfaces, composition
**45–60 minutes**

Planned OM1 anchors:

- `internal/inputs/sensor.go`
- `internal/actions/action.go`
- `internal/fuser/fuser.go`

Topics to build deeply:

- why Go does not have Python-style classes
- data versus behavior
- receiver methods
- pointer receiver versus value receiver
- interfaces as behavioral contracts
- structural/implicit interface satisfaction
- dependency injection
- composition instead of inheritance
- interface values
- typed nil versus nil
- constructors as convention
- why OM1's plugin architecture fits Go especially well

Goal:

> Look at an OM1 interface and immediately understand what a plugin has promised to implement.

---

## Module 3 — Data plumbing: slices, maps, JSON-shaped values, schemas, and errors
**40–50 minutes**

Planned OM1 anchors:

- `internal/actions/orchestrator.go`
- `internal/actions/schema.go`
- config loading code
- LLM tool-call structures

Topics:

- arrays versus slices
- `make`
- `append`
- map lookup with `value, ok`
- `map[string]any`
- nested dynamic data
- type assertions
- JSON decoding mental model
- error values
- wrapping errors
- sentinel versus typed errors
- zero values
- defensive parsing
- what a structured LLM tool call looks like in Go

Goal:

> Be able to follow structured data from an LLM response into an OM1 action.

---

## Module 4 — Concurrency without hand-waving
**75–90 minutes**

This is the most important "Go-specific" module for OM1.

Planned OM1 anchors:

- `internal/inputs/orchestrator.go`
- `internal/actions/orchestrator.go`
- `internal/runtime/runtime.go`

Topics:

- what a goroutine actually means
- concurrency versus parallelism
- channels
- send and receive syntax
- buffered versus unbuffered channels
- receive-only and send-only channels
- `select`
- `context.Context`
- cancellation
- timeouts
- `defer`
- `sync.WaitGroup`
- `sync.Mutex`
- race conditions
- snapshots of shared state
- clean shutdown
- why robotics/agent runtimes naturally benefit from concurrency

Goal:

> Read OM1's cortex/input/action loops and explain which pieces are concurrent and how they stop.

---

## Module 5 — Read OM1 as a Go program, end to end
**60 minutes**

Planned reading path:

1. `cmd/main.go`
2. `internal/runtime/runtime.go`
3. `internal/inputs/orchestrator.go`
4. `internal/fuser/fuser.go`
5. LLM response/tool-call structures
6. `internal/actions/orchestrator.go`
7. one real Unitree/Zenoh action connector

Topics:

- process startup
- config-driven construction
- registries and plugin initialization
- runtime modes
- event loop
- fusing context
- model call
- action parsing
- connector dispatch
- middleware boundary
- where to debug each failure mode

Goal:

> Explain OM1 from code, not just from architecture diagrams.

---

## Module 6 — Practical Go for an FDE interview
**30–45 minutes**

Topics:

- reading unfamiliar Go aloud
- explaining a function before understanding every line
- changing a struct safely
- adding a small function
- adding an error check
- logging useful evidence
- recognizing a concurrency bug
- making a tiny connector
- testing
- `go test`
- table-driven tests
- debugging strategy
- "I do not know this exact API, but here's how I would reason about it"

Goal:

> Be credible working in Go even if your strongest implementation language remains Python.

---

# Recommended study method

For each concept, use this four-step loop:

### 1. Plain English
Explain what the construct means without code.

### 2. Python translation
Map it to the closest thing you already know.

### 3. Tiny Go example
See the construct without OM1 complexity.

### 4. Real OM1 example
Open the actual file and identify it in context.

If a Go concept still feels mysterious after step 4, do **not** keep reading forward. Re-read the concept using a different example.

---

# Important rule: syntax versus architecture

There are two different reasons code may feel confusing:

**Syntax confusion**
> "What does `:=` mean?"

**Architecture confusion**
> "Why does this orchestrator exist?"

Do not mix these.

When reading OM1:

1. First identify the syntax.
2. Then ask what the line does.
3. Then ask why OM1 needs it.

This prevents a single unfamiliar line from turning into "I don't understand Go."

---

# Your Python advantage

You already know programming concepts.

You are not learning:
- what a function is
- what a condition is
- what a list is
- what a dictionary is
- what an object is
- what an exception-like failure is
- what concurrency is conceptually
- what an API call is

You are mostly learning:

> "How does Go represent concepts I already understand?"

That is a much smaller problem.

---

# Python → Go translation table

| Python idea | Go idea |
|---|---|
| module / package | package |
| `def` | `func` |
| class mainly holding data | `struct` |
| instance method / `self` | method + receiver |
| Protocol / ABC | `interface` |
| list | slice |
| dict | map |
| `Any` | `any` |
| exception | usually explicit `error` return |
| `None` | often `nil` |
| `with` / `finally` cleanup | often `defer` |
| thread/task | roughly: goroutine |
| queue/event | often: channel |
| asyncio wait-for-one | roughly: `select` |
| cancellation token | `context.Context` |
| decorator/registration dict | often registry + `init()` |
| constructor call | often `NewX(...)` function |

These are analogies, not exact equivalences.

---

# The files you will repeatedly revisit

Keep these open or remember the paths:

```text
cmd/main.go
internal/inputs/sensor.go
internal/inputs/orchestrator.go
internal/fuser/fuser.go
internal/actions/action.go
internal/actions/orchestrator.go
internal/runtime/runtime.go
plugins/actions/unitree/go2/autonomy/
internal/zenoh/
```

You should expect the same concepts to appear repeatedly.

That repetition is intentional.

---

# First assignment

Start here:

> **[Module 1 — Reading Go from scratch using OM1](01_READING_GO_FROM_SCRATCH_WITH_OM1.md)**

Do not begin by trying to read the entire OM1 runtime.

The first goal is simpler:

> Make `internal/inputs/sensor.go` feel readable rather than foreign.

After Module 1, complete:

> **[Module 1 Exercises](EXERCISES_01.md)**

Then try reading `internal/inputs/sensor.go` again without the lesson open.

If you can explain roughly 80% of it aloud, Module 1 worked.
