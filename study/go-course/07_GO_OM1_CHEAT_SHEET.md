# Go + OM1 Cheat Sheet

Use this for a 10–15 minute refresh before an interview or while reading the repository offline.

---

# Core Go syntax

~~~text
func F(...) T
~~~
Function returning T.

~~~text
func (x *Thing) F(...)
~~~
Method named F on pointer-to-Thing. x is roughly like Python self.

~~~text
x := value
~~~
Declare and assign local variable.

~~~text
x = value
~~~
Assign existing variable.

~~~text
result, err := F()
if err != nil { ... }
~~~
Normal Go error pattern.

~~~text
[]T
~~~
Slice of T. Think dynamic list.

~~~text
map[string]T
~~~
Map/dictionary from string to T.

~~~text
map[string]any
~~~
JSON-like dynamic dictionary.

~~~text
value, ok := m[key]
~~~
Map lookup plus key-existed boolean.

~~~text
value, ok := x.(T)
~~~
Type assertion: does interface value x contain/satisfy T?

~~~text
&T{...}
~~~
Construct T and take a pointer/reference to it.

~~~text
*T
~~~
Pointer-to-T when used as a type.

~~~text
*x
~~~
Dereference pointer x when used as an expression.

~~~text
defer cleanup()
~~~
Run cleanup when current function exits.

---

# Go object model

**struct**
> stored fields/state

**method**
> function with receiver

**interface**
> behavioral contract defined by methods

**pointer receiver**
> method works with the shared/original object rather than a copied value

**implicit interface satisfaction**
> no implements keyword; matching methods are enough

**composition**
> a component contains collaborators instead of inheriting from them

---

# Concurrency

~~~text
go F()
~~~
Start F in a goroutine.

~~~text
chan T
~~~
Channel carrying T.

~~~text
<-ch
~~~
Receive from channel.

~~~text
ch <- value
~~~
Send value.

~~~text
<-chan T
~~~
Receive-only channel.

~~~text
chan<- T
~~~
Send-only channel.

~~~text
select { ... }
~~~
Wait on multiple channel operations.

~~~text
ctx.Done()
~~~
Cancellation signal.

~~~text
sync.WaitGroup
~~~
Wait until a set of workers finish.

~~~text
sync.Mutex
~~~
Protect shared mutable state.

~~~text
atomic.Bool
~~~
Thread-safe simple boolean state.

---

# OM1 architecture in one line

~~~text
Inputs → buffers → Fuser → prompt → Cortex → ToolCalls → Action Orchestrator → Connector → middleware/service/robot
~~~

---

# OM1 file map

~~~text
cmd/main.go
~~~
Process entry point.

~~~text
internal/runtime/runtime.go
~~~
Lifecycle, active mode, Cortex event loop, main tick.

~~~text
internal/inputs/sensor.go
~~~
Sensor interface, input factory registry.

~~~text
internal/inputs/orchestrator.go
~~~
Runs sensors concurrently and exposes current buffers.

~~~text
internal/fuser/fuser.go
~~~
Builds the text prompt/context for Cortex.

~~~text
internal/llm/llm.go
~~~
LLM interface, ToolCall, Response.

~~~text
internal/llm/orchestrator.go
~~~
Wraps LLM and optional conversation history.

~~~text
internal/actions/action.go
~~~
Connector interface, AgentAction metadata, action factory registry.

~~~text
internal/actions/orchestrator.go
~~~
Parses action calls and executes them concurrent/sequential/dependency-aware.

~~~text
plugins/actions/unitree/go2/autonomy/move.go
~~~
Concrete example: high-level movement action to Zenoh cmd_vel behavior.

~~~text
internal/zenoh/session.go
~~~
Small Session/Publisher/Subscriber middleware interfaces.

---

# Fuser

Best sentence:

> The Fuser assembles the current agent context — persona, governance, processed observations, memory, knowledge, available actions, MCP descriptions, and examples — into the prompt sent to Cortex.

Important:

> It does not perform robotics sensor fusion such as Kalman filtering of raw LiDAR and IMU data.

---

# Cortex

Best sentence:

> Cortex is the model reasoning stage. It receives the fused prompt and returns an llm.Response containing text and zero or more structured ToolCalls.

Do not say:

> It just returns prose that OM1 parses into commands.

---

# ToolCall

Conceptual shape:

~~~go
type ToolCall struct {
    Name      string
    Arguments map[string]any
}
~~~

Meaning:

> structured action/tool selection with JSON-like arguments.

---

# Action Orchestrator

Best sentence:

> The action orchestrator resolves structured tool-call names to registered AgentActions and executes them according to the configured execution policy.

It does not itself know how to drive every robot.

---

# Connector

Best sentence:

> A connector is the concrete integration implementation. It translates a high-level action into a real side effect such as middleware publication, TTS, HTTP, or robot behavior.

---

# MCP

Best sentence:

> MCP is an optional tool integration mechanism in the Cortex cycle. It is not inherently the ROS translation layer.

---

# Zenoh

Best sentence:

> Zenoh is one middleware path OM1 can use. The Go2 movement connector uses a Zenoh publisher to send serialized velocity commands on cmd_vel, but not every OM1 connector necessarily uses Zenoh or ROS.

---

# Go2 movement path

~~~text
LLM ToolCall
  ↓
ParseCalls
  ↓
AgentAction
  ↓
moveConnector.Connect
  ↓
queue pending moveCommand
  ↓
moveConnector.Tick
  ↓
odometry + path/safety checks
  ↓
serialize Twist
  ↓
Zenoh Publisher.Put
  ↓
cmd_vel
  ↓
downstream robot stack
~~~

---

# Context cancellation

Best sentence:

> Context provides cooperative cancellation/deadline propagation. Canceling the OM1 mode context tells its sensor, action, background, and Cortex goroutines to stop cleanly.

---

# One-slot wake-up channel

OM1 input orchestrator uses a buffered signal channel.

Mental model:

> "There is fresh data; wake Cortex."

If one wake-up is already pending, another can be dropped because it is redundant.

---

# Mutex mental model

Do not say:

> Mutex protects a variable.

Better:

> Mutex protects an invariant across shared state while concurrent goroutines access or modify it.

---

# FDE debugging stack

~~~text
1. safety
2. reproduce expected vs actual
3. isolate layer
4. inspect evidence at boundaries
5. form hypothesis
6. change one thing
7. verify
8. robustness test
9. preserve observability/runbook
~~~

OM1 boundaries:

~~~text
input
→ formatted buffer
→ Fuser prompt
→ LLM ToolCall
→ action parse
→ connector
→ middleware
→ ROS/HAL/controller
→ hardware
~~~

Find the first boundary where actual behavior diverges from expected.

---

# ROS2 tools quick recall

~~~text
ros2 topic list
~~~
What topics exist?

~~~text
ros2 topic info /topic
~~~
Who publishes/subscribes and what type?

~~~text
ros2 topic hz /topic
~~~
What message rate?

~~~text
ros2 topic echo /topic
~~~
What values are flowing?

~~~text
rqt_graph
~~~
Node/topic connectivity graph.

~~~text
RViz2
~~~
Spatial robot data: map, TF, scans, pose, paths, costmaps.

~~~text
rosbag2
~~~
Record/replay topic data for offline debugging.

---

# Commands to recognize

~~~text
go test ./...
~~~
Run tests.

~~~text
go test -race ./...
~~~
Run tests with race detector.

~~~text
go test ./internal/actions
~~~
Run one package's tests.

~~~text
go build ./...
~~~
Compile packages.

~~~text
go vet ./...
~~~
Static suspicious-code checks.

~~~text
gofmt -w file.go
~~~
Format Go source.

---

# Interview answers

## "How much Go do you know?"

> Python is still my strongest language. I've been ramping on Go directly through OM1, so I'm comfortable reading structs, receiver methods, interfaces, explicit errors, maps and slices, and I'm tracing its goroutines, channels, contexts, mutexes, and connector architecture. I wouldn't claim years of production Go yet, but I can read the codebase, reason about it, and make targeted changes.

## "What is a Go interface?"

> A behavioral contract defined by method signatures. Types satisfy it implicitly by implementing those methods. OM1 uses interfaces to decouple the runtime from concrete sensors, LLMs, action connectors, and middleware implementations.

## "Why Go for OM1?"

> It gives a useful combination of compiled deployment, explicit typing, relatively low runtime overhead, and first-class concurrency primitives. OM1 has many long-lived inputs, action loops, cancellation paths, and middleware connections, so goroutines, channels, contexts, and synchronization fit that workload well.

## "What is a goroutine?"

> A lightweight concurrent function managed by the Go runtime.

## "What is a channel?"

> A typed communication and synchronization mechanism between goroutines.

## "What is context?"

> A standard way to propagate cancellation and deadlines through a call/lifecycle tree.

---

# Final 60-second verbal summary

> OM1 is a configurable agent runtime. Go structs hold state, receiver methods implement component behavior, and interfaces let the runtime depend on contracts rather than concrete plugins. Input plugins run concurrently and expose formatted sensor state. The Fuser assembles that state plus persona, memory, knowledge, and available tools/actions into the Cortex prompt. Cortex returns a typed response with structured ToolCalls. The action orchestrator resolves those calls to registered AgentActions, and concrete connectors turn the high-level decisions into actual middleware, service, or robot behavior. Go's goroutines, channels, contexts, mutexes, and explicit errors support the runtime's concurrency and lifecycle.

If every clause in that paragraph makes sense, you are ready to read the code.
