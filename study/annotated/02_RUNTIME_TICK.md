# Annotated Source: `Runtime.tick` — The Cortex Cycle

Source:

> `internal/runtime/runtime.go`

This is probably the **single most important function** to understand in OM1.

---

# Function signature

```go
func (rt *Runtime) tick(
    ctx context.Context,
    current *modeState,
    tickStart time.Time,
)
```

Python-ish:

```python
def tick(self, ctx, current: ModeState, tick_start: datetime):
    ...
```

Arguments:

- `ctx` — cancellation/shutdown signal
- `current` — current mode's running components
- `tickStart` — used to measure cycle timing

---

# 1. Exit if canceled

```go
if ctx.Err() != nil {
    return
}
```

Before doing expensive work, make sure shutdown has not been requested.

---

# 2. Check whether mode reload/transition is active

```go
rt.mu.Lock()
reloading := rt.isReloading
rt.mu.Unlock()

if reloading {
    rt.log.Debug("skipping tick during mode transition")
    return
}
```

Why Mutex?

The transition state may be touched by another goroutine.

So the mutex protects concurrent access.

FDE intuition:

> Don't let the old mode perform new work while the runtime is changing modes.

---

# 3. Advance tick counter

```go
rt.ioProvider.IncrementTick()
```

The IO provider tracks which data belongs to which cortex cycle.

This is useful for deciding whether an input such as voice belongs to the current tick.

---

# 4. Take a snapshot of sensor state

```go
sensorBuffers := current.inputOrchestrator.Buffers()
```

This is a major boundary.

The sensors have been running concurrently.

The Cortex does **not** directly stop and query every physical device here.

Instead it asks:

> "What is the latest formatted state from each sensor?"

That creates a snapshot suitable for prompt fusion.

Mental model:

```text
microphone worker ─┐
camera worker ─────┼─ latest buffers ─> cortex tick
localization ──────┘
```

---

# 5. Check for a mode transition

```go
if rt.scheduleTransition(
    rt.manager.CheckTransitions(ctx, sensorBuffers),
) {
    return
}
```

Input state can cause the system to move into another configured mode.

If a transition is scheduled, this tick stops rather than doing work for a mode that is leaving.

---

# 6. Fuse the prompt

```go
prompt, err := current.promptFuser.Fuse(ctx, sensorBuffers)
if err != nil {
    rt.log.Warn("fuse failed", zap.Error(err))
    return
}
```

This converts system state into model context.

The Fuser adds:
- persona
- governance
- current time
- observations
- KB
- memory
- available actions
- MCP tools
- examples

This is where **robot state becomes language-model-readable context**.

Important debugging question:

> If the LLM made a bad decision, was its prompt actually correct?

---

# 7. Check cancellation again

```go
if ctx.Err() != nil {
    return
}
```

Why repeat this?

Fusion may take time or involve external operations such as a knowledge-base query.

A shutdown could have happened while it was working.

---

# 8. Log the Cortex tick

```go
rt.log.Info(
    "cortex tick",
    zap.String("mode", rt.manager.CurrentMode()),
    zap.String("prompt", prompt),
)
```

This can be extremely useful operationally.

It lets a debugger ask:

> "What prompt did the model actually see?"

That is much stronger evidence than guessing from expected config.

---

# 9. Call the LLM

```go
response, err := current.cortexLLM.Call(ctx, prompt, nil)
if err != nil {
    rt.log.Warn("llm call failed", zap.Error(err))
    return
}
```

The Cortex Orchestrator wraps the concrete model and handles history/schema behavior.

The runtime does not need:

```text
if Gemini...
else if Ollama...
else if Qwen...
```

That provider-specific variation is abstracted away.

---

# 10. Trace response quality/output

```go
rt.tracer.Gauge(prompt, traceOutput(response))
```

Operational systems need observability around model behavior, not only CPU/process health.

This layer helps evaluate prompt/response behavior.

---

# 11. Extract tool calls

```go
toolCalls := response.ToolCalls
```

This is the moment where reasoning begins to become action.

Example conceptual response:

```json
{
  "tool_calls": [
    {
      "name": "move",
      "arguments": {
        "action": "move forwards"
      }
    }
  ]
}
```

---

# 12. Optional MCP resolution

If MCP is configured:

```go
toolCalls = current.mcpOrchestrator.Resolve(...)
```

The resolver may:
- inspect/resolve MCP-related calls
- call the Cortex again with additional information
- execute action calls during that process

For your first OM1 understanding, treat this as an optional extension around the basic loop.

Core mental path remains:

```text
prompt → LLM → tool calls → actions
```

---

# 13. Execute actions

```go
rt.executeActions(ctx, current, toolCalls)
```

Inside `executeActions`:

```text
raw LLM ToolCalls
      ↓
toolCallsToMaps
      ↓
ActionOrchestrator.ParseCalls
      ↓
typed Call objects
      ↓
ActionOrchestrator.Submit
      ↓
Connector.Connect
```

The Action Orchestrator selects the configured execution strategy.

---

# 14. Record memory

If memory is enabled, it checks whether fresh voice input exists for this tick.

Then it records:
- user input
- spoken model response
- user identity context

and triggers summarization.

Important architecture point:

> Memory is not simply "LLM history."

OM1 has an explicit memory manager in addition to short conversation history in the LLM orchestrator.

---

# 15. Record tick timing

```go
rt.ioProvider.RecordTick(tickStart)
```

This closes the cycle with timing/telemetry.

Latency matters in robotics.

An agent that is semantically correct but responds five seconds too late may be operationally useless.

---

# The whole tick as pseudo-Python

This is not actual OM1 Python. It is a translation for understanding:

```python
def tick(self, current):
    if shutting_down():
        return

    if self.is_reloading:
        return

    self.io.increment_tick()

    sensor_state = current.inputs.latest_buffers()

    next_mode = self.manager.check_transitions(sensor_state)
    if next_mode:
        self.schedule_transition(next_mode)
        return

    prompt = current.fuser.fuse(sensor_state)

    response = current.llm.call(prompt)

    self.tracer.record(prompt, response)

    tool_calls = response.tool_calls

    if current.mcp:
        tool_calls = current.mcp.resolve(tool_calls)

    current.actions.execute(tool_calls)

    if current.memory:
        current.memory.record_current_interaction()

    self.io.record_tick_latency()
```

If that pseudo-code makes sense, the Go source is already becoming readable.

---

# FDE diagnostic map for one tick

```text
1. Sensors produced state?
2. Buffers contain correct state?
3. Fuser prompt contains it?
4. LLM call succeeded?
5. Model emitted correct tool call?
6. Tool call parsed?
7. Connector accepted command?
8. Middleware transmitted it?
9. Lower robot stack acted?
10. Feedback confirms effect?
```

This is the operational value of understanding the code.

---

# Your oral check

Try answering without rereading:

> "What are the five most important steps in an OM1 cortex tick?"

A strong simplified answer:

> capture current sensor state → fuse context → call LLM → dispatch structured actions → record state/telemetry.
