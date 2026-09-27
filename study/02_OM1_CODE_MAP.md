# OM1 Code Map — What Each Important Area Does

This file answers:

> "I opened the repo. Where do I look?"

---

# Level 0 — the executable

## `cmd/main.go`

This is the Go program entry point.

Responsibilities:
- parse command-line flags
- initialize logging
- start metrics
- load the JSON5 system config
- create a cancellable context for SIGINT/SIGTERM
- construct the OM1 Runtime
- call `Runtime.Run`

It does **not** contain the intelligence itself.

Think of it as the ignition key.

---

# Level 1 — runtime lifecycle

## `internal/runtime/runtime.go`

This is the best single file for understanding OM1.

It owns:
- active mode state
- input orchestrator
- fuser
- Cortex LLM orchestrator
- action orchestrator
- backgrounds
- memory
- MCP integration
- mode transitions
- the Cortex tick loop
- startup/shutdown

Critical functions:

```text
New()
Run()
initializeMode()
startOrchestrators()
runCortexLoop()
tick()
executeActions()
```

### Core path

```text
Run
  ↓
initializeMode
  ↓
startOrchestrators
  ↓
runCortexLoop
  ↓
tick
  ├─ get sensor buffers
  ├─ Fuser.Fuse
  ├─ CortexLLM.Call
  ├─ resolve MCP calls if configured
  └─ executeActions
```

---

# Level 2 — the four pieces of one agent cycle

## A. Inputs — `internal/inputs/`

### `sensor.go`

Defines the `Sensor` interface.

A Sensor knows how to:
- listen continuously
- poll once
- turn raw data into text
- expose latest formatted state
- stop

This abstraction lets:
- microphone ASR
- camera/VLM
- localization
- robot state

all participate in the runtime through one common contract.

### `orchestrator.go`

Runs each sensor concurrently.

It:
- starts one goroutine per sensor
- receives readings
- calls `RawToText`
- signals the Cortex loop when appropriate
- provides snapshots of each sensor's latest buffer

Mental model:

```text
Sensor A ─┐
Sensor B ─┼─> InputOrchestrator ─> latest text buffers
Sensor C ─┘
```

---

## B. Fuser — `internal/fuser/fuser.go`

The Fuser creates the actual text prompt passed to the LLM.

It combines:

```text
system persona
+ governance
+ current time
+ current sensor observations
+ knowledge-base context
+ memory
+ available actions
+ MCP tool descriptions
+ examples
+ "What will you do next?"
```

This is a critical conceptual point:

> OM1 does not send raw robot state blindly to the LLM. It assembles a structured textual context from multiple sources.

For interview discussion, this is analogous to **context engineering** around an agent.

---

## C. LLM — `internal/llm/`

### `llm.go`

Defines the model abstraction and response/tool-call data structures.

### `orchestrator.go`

Wraps an LLM with:
- history handling
- tool schemas
- synchronized state

It receives:

```text
prompt + conversation history
```

and returns a response that can contain:
- text
- structured tool calls

The concrete provider comes from `plugins/llm/`.

Examples:
- Gemini
- OpenAI-compatible
- DeepSeek
- Ollama
- Qwen
- xAI

---

## D. Actions — `internal/actions/`

### `action.go`

Defines the `Connector` interface:

```go
type Connector interface {
    Connect(ctx context.Context, input Input) (Output, error)
    Tick(ctx context.Context)
    Stop()
}
```

Translation:

A connector can:
- execute an action
- perform recurring background/heartbeat work
- shut down cleanly

### `orchestrator.go`

Receives LLM action calls and executes them.

Execution modes:
- concurrent
- sequential
- dependency-aware

The concrete implementation comes from `plugins/actions/`.

---

# Level 3 — plugin registries

There are two layers:

```text
internal abstraction
        +
plugin registry
        +
concrete plugin package
```

Example input:

```text
internal/inputs.Sensor
        ↑
inputs.Register("VLMGemini", NewVLMGemini)
        ↑
plugins/inputs/vlm/vlm_gemini.go
```

Example model:

```text
internal/llm.LLM
        ↑
llm.Register("GeminiLLM", NewGemini)
        ↑
plugins/llm/gemini.go
```

Example action:

```text
internal/actions.Connector
        ↑
actions.Register("unitree_go2_autonomy/move", NewMoveConnector)
        ↑
plugins/actions/unitree/go2/autonomy/move.go
```

---

# Level 4 — configuration

## `config/*.json5`

This is where OM1 becomes a particular agent.

A config can choose:
- persona/governance
- inputs
- LLM
- actions
- knowledge base
- tracing
- modes
- execution behavior

Example from `conversation.json5`:

```json5
agent_inputs: [
  { type: "GoogleASRInput", ... },
  { type: "VLMGemini" }
],

cortex_llm: {
  type: "GeminiLLM",
  ...
},

agent_actions: [
  {
    name: "speak",
    connector: "elevenlabs_tts"
  },
  {
    name: "emotion",
    connector: "zenoh"
  }
]
```

The important design idea:

> behavior is assembled through configuration rather than hard-coded into one monolithic agent class.

---

# Level 5 — middleware / robot bridge

## `internal/zenoh/`

Wraps Zenoh sessions, publishers, and subscribers.

This creates a cleaner OM1-side abstraction so action plugins do not have to repeat session setup details everywhere.

A connector can do something like:

```text
session.DeclarePublisher("cmd_vel")
publisher.Put(bytes)
```

without owning every lower-level Zenoh detail.

---

# Level 6 — a real robot action

## `plugins/actions/unitree/go2/autonomy/move.go`

This is a concrete connector.

It:
- registers a tool/action interface
- defines valid high-level movement commands
- creates Zenoh publishers/subscribers
- reads odometry
- consults safe path state
- rejects unsafe/conflicting commands
- converts high-level movement into velocity behavior
- publishes `cmd_vel`

This file is where "LLM agent" becomes "physical robot".

---

# Level 7 — full autonomy architecture

The public core OM1 repository is only part of the complete robot stack.

The docs describe additional services around full autonomy, including:
- OM1 ROS2 SDK
- sensor services
- SLAM/navigation orchestration
- watchdog
- Zenoh bridge
- video processing
- OTA containers

For study, read:

> `docs/full_autonomy_guidelines/architecture_overview.md`

This is important because an FDE may debug a deployment across **multiple processes/services**, not only this Go binary.

---

# Three architecture boundaries to memorize

## Boundary 1: environment → OM1

```text
microphone / camera / robot state
            ↓
        input plugins
            ↓
       text/state buffers
```

Debugging questions:
- Is the sensor alive?
- Is data fresh?
- Is parsing correct?
- Is the buffer sane?

## Boundary 2: OM1 cognition

```text
buffers
  ↓
Fuser
  ↓
prompt
  ↓
LLM
  ↓
tool calls
```

Debugging questions:
- Did the right state enter the prompt?
- Was the right action exposed?
- Did the model choose the right tool?
- Was the schema valid?

## Boundary 3: decision → physical action

```text
tool call
  ↓
action orchestrator
  ↓
connector
  ↓
Zenoh / ROS2 / HTTP / SDK
  ↓
robot HAL
  ↓
hardware
```

Debugging questions:
- Did the action call parse?
- Did the connector run?
- Did it publish/send the command?
- Did middleware deliver it?
- Did the HAL accept it?
- Did physical feedback confirm execution?

This boundary-based model is exactly how to reason as an FDE.

---

# What to read next

1. [Annotated main entry point](annotated/01_CMD_MAIN.md)
2. [Runtime walkthrough](03_RUNTIME_WALKTHROUGH.md)
3. [Annotated cortex tick](annotated/02_RUNTIME_TICK.md)
4. [Plugin and robot flow](04_PLUGIN_AND_ROBOT_FLOW.md)
