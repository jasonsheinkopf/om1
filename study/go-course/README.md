# Go Through OM1 — Complete Deep Course for a Python Developer

This is the **full offline Go course** for this OM1 study fork.

It is designed for a Python-fluent engineer who wants to become comfortable enough with Go to:

- read real OM1 source
- explain unfamiliar Go code
- understand OM1 architecture
- reason about goroutines/channels/context/mutexes
- trace model ToolCalls into real robot actions
- make small targeted changes
- debug field integrations
- speak credibly about Go in an FDE interview

The goal is practical fluency, not language-lawyer mastery.

Everything needed for the course is in this repository after you pull it.

---

# Recommended order

## Module 1 — Go from scratch through real OM1 code
**60–90 min**

[01_READING_GO_FROM_SCRATCH_WITH_OM1.md](01_READING_GO_FROM_SCRATCH_WITH_OM1.md)

Covers:
- packages/imports/main
- variables and types
- functions and multiple returns
- explicit errors
- structs
- pointers
- methods/receivers
- slices and maps
- interfaces
- factories and registries
- type assertions
- nil
- defer
- exports
- reading internal/inputs/sensor.go

Then do:

[EXERCISES_01.md](EXERCISES_01.md)

This includes 25 exercises and a full answer key.

---

## Module 2 — Structs, interfaces, pointers, and composition
**45–60 min**

[02_STRUCTS_INTERFACES_AND_COMPOSITION.md](02_STRUCTS_INTERFACES_AND_COMPOSITION.md)

Covers:
- Go object model versus Python classes
- pointer versus value receivers
- implicit interface satisfaction
- interface values
- optional capability interfaces
- dependency injection
- composition
- typed nil
- action Connector architecture
- why OM1's plugin design fits Go

Primary OM1 anchors:
- internal/inputs/sensor.go
- internal/actions/action.go
- internal/fuser/fuser.go
- plugins/actions/unitree/go2/autonomy/move.go

---

## Module 3 — Data, errors, JSON-shaped values, and ToolCalls
**40–50 min**

[03_DATA_ERRORS_AND_TOOL_CALLS.md](03_DATA_ERRORS_AND_TOOL_CALLS.md)

Covers:
- slices and capacity
- maps
- map comma-ok
- any
- type assertions
- named scalar types
- struct tags
- ToolCall and Response
- schema-driven actions
- error values
- error wrapping
- nil map versus nil slice
- serialization boundaries
- defensive parsing
- model output → internal Call → Connector

Primary OM1 anchors:
- internal/llm/llm.go
- internal/actions/action.go
- internal/actions/orchestrator.go
- plugins/actions/unitree/go2/autonomy/move.go

---

## Module 4 — Concurrency without hand-waving
**75–90 min**

[04_CONCURRENCY_WITH_OM1.md](04_CONCURRENCY_WITH_OM1.md)

This is the most Go-specific part of the course.

Covers:
- concurrency versus parallelism
- goroutines
- anonymous goroutines
- channels
- signal channels
- channel direction
- buffered versus unbuffered channels
- non-blocking send
- select
- timers
- context cancellation
- context trees/timeouts
- WaitGroups
- mutexes
- atomics
- races
- deadlocks
- goroutine leaks
- shutdown
- Cortex event-loop reasoning

Primary OM1 anchors:
- internal/inputs/orchestrator.go
- internal/actions/orchestrator.go
- internal/runtime/runtime.go
- plugins/actions/unitree/go2/autonomy/move.go

---

## Module 5 — OM1 end to end in Go
**60 min**

[05_OM1_END_TO_END_IN_GO.md](05_OM1_END_TO_END_IN_GO.md)

Traces:

~~~text
config
→ main
→ Runtime
→ mode initialization
→ sensors
→ input orchestrator
→ buffers
→ Fuser
→ Cortex
→ ToolCalls
→ MCP resolution when configured
→ Action Orchestrator
→ Connector
→ Zenoh/service/robot
~~~

Includes a detailed Go2 movement walkthrough and an FDE debugging map.

This module also corrects several easy oversimplifications:
- OM1 is not merely "ROS with an LLM on top"
- not every connector goes through ROS/Zenoh
- MCP is not automatically the ROS translator
- Cortex returns structured ToolCalls, not merely prose

---

## Module 6 — Practical Go and FDE interview preparation
**30–45 min**

[06_PRACTICAL_GO_AND_FDE_INTERVIEW.md](06_PRACTICAL_GO_AND_FDE_INTERVIEW.md)

Covers:
- six-pass unfamiliar-code reading
- safe small changes
- validation
- observability
- timeouts
- synchronization reasoning
- race detector
- tests
- table-driven tests
- gofmt / go vet / go test / go build
- debugging scenarios
- coding strategy
- honest answers about current Go proficiency
- common interview traps
- OM1-specific verbal questions

---

## Final refresh — cheat sheet
**10–15 min**

[07_GO_OM1_CHEAT_SHEET.md](07_GO_OM1_CHEAT_SHEET.md)

Use this:
- before the recruiter/technical interview
- after a break
- on the plane when you need a fast reactivation
- before opening a complicated OM1 file

---

# Five-hour path

If you have one long study block:

~~~text
Hour 0:00–1:15  Module 1
Hour 1:15–1:35  Module 1 exercises
Hour 1:35–2:20  Module 2
Hour 2:20–3:00  Module 3
Hour 3:00–4:10  Module 4
Hour 4:10–4:50  Module 5
Hour 4:50–5:15  Module 6 + cheat sheet
~~~

If you are tired, do not force the timing.

Understanding beats completion speed.

---

# The one OM1 mental model to retain

~~~text
Inputs
  ↓
Sensor buffers
  ↓
Fuser
  ↓
Prompt
  ↓
Cortex LLM
  ↓
structured ToolCalls
  ↓
Action Orchestrator
  ↓
Connector
  ↓
Zenoh / HTTP / TTS / robot/service interface
~~~

Configuration chooses concrete implementations.

Go interfaces keep the central runtime decoupled from those implementations.

Concurrency lets the runtime manage many long-lived activities simultaneously.

---

# The learning rule

For every new Go idea, use four passes:

1. **plain English**
2. **Python analogy**
3. **tiny Go example**
4. **real OM1 example**

And keep two kinds of confusion separate:

**Syntax**
> What does this symbol mean?

**Architecture**
> Why does this component exist?

Solve syntax first, architecture second.

---

# What "done" means

After the course, you should be able to open a normal OM1 Go file and answer:

- What package is this?
- What are the important structs/interfaces?
- Is this a function or method?
- Why is the receiver a pointer?
- What data is in this slice/map?
- What is dynamic versus statically typed?
- What can fail?
- Which error path handles it?
- Is a goroutine being launched?
- What is this channel signaling?
- What can cancel this operation?
- What state is protected by this mutex?
- What ToolCall/action is being routed?
- What external side effect happens next?
- Where would I collect evidence if it broke?

You do not need to understand every library call immediately.

The target is:

> unfamiliar Go code is readable, navigable, and debuggable.

---

# Files worth keeping open while studying

~~~text
cmd/main.go
internal/inputs/sensor.go
internal/inputs/orchestrator.go
internal/fuser/fuser.go
internal/llm/llm.go
internal/llm/orchestrator.go
internal/actions/action.go
internal/actions/orchestrator.go
internal/runtime/runtime.go
plugins/actions/unitree/go2/autonomy/move.go
internal/zenoh/session.go
~~~

---

# Start

Go to:

[Module 1 — Reading Go from Scratch with OM1](01_READING_GO_FROM_SCRATCH_WITH_OM1.md)

If you already completed Module 1, continue directly to:

[Module 2 — Structs, Interfaces, and Composition](02_STRUCTS_INTERFACES_AND_COMPOSITION.md)
