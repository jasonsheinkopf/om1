# Module 5 — Read OM1 End to End as a Go Program

This is the integration module.

Until now, you learned pieces of Go.

Now the goal is:

> Follow one real OM1 execution path and understand what each Go construct is doing in the architecture.

Do not memorize every function.

Learn the shape of the system.

---

# 1. The entire runtime in one picture

Keep this mental model:

~~~text
configuration
     ↓
cmd/main.go
     ↓
runtime.New(...)
     ↓
Runtime.Run(...)
     ↓
initializeMode(...)
     ↓
construct sensors / LLM / actions / Fuser / MCP / backgrounds
     ↓
startOrchestrators(...)
     ↓
input goroutines + action loops + background loops + cortex loop
     ↓
sensor state
     ↓
Fuser.Fuse(...)
     ↓
prompt
     ↓
cortexLLM.Call(...)
     ↓
llm.Response
     ├─ TextContent
     └─ []ToolCall
            ↓
       MCP resolution if configured
            ↓
       executeActions(...)
            ↓
       ParseCalls(...)
            ↓
       actions.Orchestrator.Submit(...)
            ↓
       Connector.Connect(...)
            ↓
       concrete integration
       Zenoh / HTTP / TTS / service / robot
~~~

That is the backbone.

---

# 2. Important correction to an oversimplified mental model

It is tempting to say:

> "OM1 sits above ROS2 and Zenoh is always the bridge."

That is too narrow.

A better statement is:

> OM1 is an agent runtime that can integrate with robot and service infrastructure through pluggable connectors. In some robot deployments that path uses Zenoh and ROS2-facing topics, while other connectors may use HTTP, TTS services, SDKs, or other interfaces.

For the Unitree Go2 autonomy connector in this repository, the concrete movement path does publish velocity data through a Zenoh publisher to a cmd_vel key/topic.

But that is an implementation of one connector, not a universal law of OM1.

This distinction is important in an interview.

---

# 3. Process entry: cmd/main.go

Open:

> cmd/main.go

Do not read line by line at first.

Read it as stages.

## Stage A — command-line configuration

The program declares flags such as:
- config
- log level
- hot reload
- check interval

Then:

~~~go
flag.Parse()
~~~

This converts command-line text into typed Go values.

---

# 4. Pointer surprise from flag.String

The standard flag package returns pointers for flag variables.

So:

~~~go
configName := flag.String(...)
~~~

conceptually gives:

~~~text
*string
~~~

Then:

~~~go
*configName
~~~

means:

> get the actual configured string.

Remember:
- star in a type: pointer to T
- star in an expression: dereference pointer

---

# 5. Logger and cleanup

main constructs logging and uses defer for cleanup.

This is a good Go lifecycle pattern:

~~~text
create resource
defer cleanup
continue setup
~~~

The cleanup code stays near resource acquisition rather than at the bottom of the function.

---

# 6. Load configuration

Conceptually:

~~~go
cfg, err := config.Load(...)
if err != nil {
    ...
}
~~~

This is ordinary Go error handling.

Architecture:

> external configuration becomes typed runtime configuration.

If OM1 fails here, no robot reasoning loop has started yet.

That matters for debugging.

---

# 7. Root cancellation context

main creates a context connected to OS signals such as SIGINT and SIGTERM.

Conceptually:

~~~text
Ctrl+C / process termination
       ↓
root context canceled
       ↓
Runtime notices
       ↓
child contexts stop
       ↓
orchestrators clean up
~~~

This is the root of OM1's shutdown tree.

---

# 8. Construct Runtime

main calls:

~~~go
rt := runtime.New(cfg, log, runtime.Options{...})
~~~

This is constructor-style Go.

Runtime.New:
- stores configuration
- stores logger/options
- creates the mode manager
- obtains shared providers/tracer
- creates mode-transition channel
- configures default Zenoh options

Notice the architecture:

> New builds the runtime object; Run owns the lifecycle.

This New/Run split is common in Go services.

---

# 9. Runtime.Run is lifecycle ownership

Open:

> internal/runtime/runtime.go

Runtime.Run conceptually does:

1. optionally start config watching
2. load global backgrounds
3. initialize initial mode
4. run startup hooks
5. start orchestrators
6. start global backgrounds
7. start tracer
8. block until context cancellation
9. stop everything
10. return

That is not business logic.

It is lifecycle orchestration.

---

# 10. initializeMode is dependency assembly

This is one of the most important architectural functions.

initializeMode:
- finds the selected mode config
- loads components
- creates runtime config
- creates optional knowledge base
- creates optional memory manager
- creates optional MCP orchestrator
- creates Fuser
- creates LLM orchestrator
- creates action orchestrator
- stores sensors
- creates hooks/background state

This is dependency construction.

If you know dependency injection frameworks from other languages, this is similar in purpose but explicit.

---

# 11. modeState is a bundle of active components

The modeState struct contains the objects belonging to one active mode.

Examples:
- runtime config
- Fuser
- Cortex LLM orchestrator
- action orchestrator
- MCP orchestrator
- background orchestrator
- sensors
- input orchestrator
- hooks
- memory
- cancellation function
- completion channels

This tells you:

> modes are not merely string labels. A mode owns a set of runtime components and lifecycle state.

---

# 12. Why pointers dominate modeState

Fields such as:

~~~text
*Fuser
*llm.Orchestrator
*actions.Orchestrator
~~~

are pointers to long-lived component instances.

They have:
- mutable state
- histories
- locks
- external resources
- lifecycle

Passing copies would be inappropriate.

---

# 13. startOrchestrators

After mode construction, Runtime starts the long-lived loops.

Conceptually:

~~~text
Input Orchestrator.Start
Action Orchestrator.Start
Background Orchestrator.Start
go runCortexLoop
~~~

These run concurrently under the mode context.

This is where Module 4's goroutine/context knowledge becomes useful.

---

# 14. Input path: concrete sensor to buffer

Open:

> internal/inputs/orchestrator.go

Each Sensor runs concurrently.

A sensor:
1. listens for raw readings
2. converts raw input using RawToText
3. maintains or exposes its latest formatted buffer
4. optionally signals immediate cortex wake-up

The important subtle point:

> The input orchestrator does not itself perform every sensor's domain logic.

The concrete sensor plugin owns that conversion/behavior through the Sensor interface.

---

# 15. RawToText does not imply every raw sensor becomes naive prose

The name can tempt you to think:

> all raw bytes are simply converted into an English sentence.

Do not overgeneralize.

The abstraction says the sensor provides a processed Message and formatted latest buffer suitable for the context layer.

Concrete plugins can perform substantial processing before that point.

For example:
- speech can become transcription
- visual systems can become descriptions
- localization can become formatted state
- robot providers can expose summarized state

The Fuser consumes the prepared textual buffer, not raw LiDAR packets.

---

# 16. Cortex wake-up

The Cortex loop can run:
- periodically at configured Hertz
- immediately when an input triggers TickNow
- in response to mode-context updates

This hybrid design matters.

It is not simply:
> every sensor message causes one model call.

And it is not simply:
> poll the model at a fixed rate forever.

The event loop supports both periodic and event-triggered behavior.

---

# 17. runCortexLoop is an event loop

Read:

> internal/runtime/runtime.go

The structure is conceptually:

~~~text
loop:
    create timer

    wait for:
        cancellation
        timer
        input wake-up
        mode context update

    possibly transition mode

    otherwise call tick(...)
~~~

This is the heartbeat of the agent runtime.

---

# 18. The tick function is the core reasoning cycle

The comment in code describes a cortex cycle.

Conceptually:

1. check cancellation/reload state
2. increment tick
3. snapshot sensor buffers
4. check mode transition
5. fuse prompt
6. call Cortex LLM
7. record tracing
8. resolve MCP tool activity if configured
9. execute remaining action calls
10. record memory interaction
11. record tick telemetry

That is the main "think and act" loop.

---

# 19. Sensor buffers are strings at this boundary

This line conceptually produces:

~~~text
[]string
~~~

Each entry is a formatted current observation from a sensor.

Then:

~~~text
Fuser.Fuse(ctx, sensorBuffers)
~~~

receives those strings.

This confirms an important mental model:

> sensor-specific processing happens before the Fuser; the Fuser assembles the agent's LLM context.

---

# 20. Fuser is the context builder

Open:

> internal/fuser/fuser.go

The Fuser writes several categories into one prompt.

Current code includes:
- base system persona
- governance
- current time
- current observations
- knowledge-base context when available
- long-term memory context when available
- available actions
- MCP tool descriptions
- prompt examples
- closing question

The result is:

~~~text
string
~~~

That string is passed to Cortex.

So a concise description is:

> "The Fuser turns the current agent state and capabilities into the textual context that the Cortex LLM reasons over."

---

# 21. Why the Fuser stores interfaces

Fuser does not need concrete implementation details for every collaborator.

It has small contracts such as:
- knowledge query
- MCP tool description

This is Module 2 in action.

The Fuser depends on what collaborators can do, not their concrete class hierarchy.

---

# 22. Cortex LLM is an orchestrator around an LLM

Open:

> internal/llm/orchestrator.go

This object wraps an LLM implementation and optionally manages conversation history.

If history is disabled, it delegates directly.

If history is enabled, it:
1. locks
2. snapshots history
3. unlocks
4. calls the model
5. locks again
6. appends new messages/tool-call text
7. trims history
8. unlocks

This is an excellent real example of:
- interface delegation
- mutex-protected state
- copying a slice snapshot
- error handling

---

# 23. Why snapshot history outside the lock

The code copies history while holding the mutex, then unlocks before calling the LLM.

That is a strong concurrency pattern.

An LLM call may take a long time.

Holding the mutex across a network/model call would block other operations unnecessarily.

So:

~~~text
lock
copy shared state
unlock
slow external call
lock
update state
unlock
~~~

This is exactly the kind of design judgment interviewers care about.

---

# 24. LLM Response is not just text

Open:

> internal/llm/llm.go

Response contains:
- TextContent
- ToolCalls
- Usage

A ToolCall contains:
- Name
- Arguments

So the proper sentence is:

> "Cortex returns a typed response that can include natural-language text and structured tool/action calls."

Do not say:

> "The LLM just outputs text and OM1 parses commands out of the prose."

That would misrepresent this architecture.

---

# 25. Action schemas guide structured output

During mode initialization, OM1 collects action schemas and gives them to the LLM orchestrator.

That means the model knows:
- which actions are available
- what arguments they expect

Then the response can contain structured ToolCalls.

This is ordinary modern tool/function calling.

---

# 26. Where MCP fits

MCP is not "the thing that converts the model output into ROS."

In the cortex tick:

1. model produces ToolCalls
2. if an MCP orchestrator exists, it can resolve MCP-related calls and may call Cortex again with returned tool context
3. OM1 then executes agent action calls through the action orchestrator

MCP is a tool integration mechanism.

The robot action path is through registered actions/connectors.

Some deployments could expose robot-related tools through MCP, but that is not the same as saying MCP is universally the ROS translation layer.

---

# 27. executeActions

The runtime passes structured calls toward the action system.

Conceptually:

~~~text
[]llm.ToolCall
       ↓
toolCallsToMaps
       ↓
ParseCalls
       ↓
[]actions.Call
       ↓
Submit
~~~

ParseCalls resolves the action name to the actual registered AgentAction.

This is the symbolic-to-concrete transition.

---

# 28. AgentAction connects metadata to behavior

Open:

> internal/actions/action.go

AgentAction stores:
- action name
- LLM label
- schema
- prompt visibility
- Connector

This is a powerful struct.

It links:

~~~text
what the LLM knows the action as
           ↓
schema / label
           ↓
what code actually executes
           ↓
Connector
~~~

---

# 29. Action orchestrator decides execution policy

Open:

> internal/actions/orchestrator.go

Submit chooses among:
- concurrent
- sequential
- dependency-aware

The orchestrator does not know how to move a robot.

Its responsibility is:

> schedule and execute action calls using their connectors.

That separation of concerns is important.

---

# 30. Connector is the integration boundary

Connector interface:

~~~text
Connect
Tick
Stop
~~~

A connector is where high-level agent action becomes a concrete effect.

Different connectors can:
- publish middleware messages
- call APIs
- speak
- update robot state
- invoke device functionality

This is where an FDE often spends time.

Because this is where:
- customer environment
- middleware
- network
- robot SDK
- permissions
- timing
- hardware expectations

meet the product.

---

# 31. Real Go2 movement connector

Open:

> plugins/actions/unitree/go2/autonomy/move.go

This file is worth studying repeatedly.

It contains:
- custom named action type
- input struct and tags
- init registration
- constructor
- interface return
- dynamic input assertion
- switch
- shared movement state
- mutex
- atomic boolean
- timing
- odometry
- path/safety checks
- Zenoh session
- publisher/subscriber
- byte serialization
- error handling
- cleanup

It is almost a miniature Go course by itself.

---

# 32. init registration

The file's init function registers:
- action interface/schema description
- concrete connector factory

The package is loaded because OM1's executable imports plugin packages for side effects.

This creates the plugin chain:

~~~text
blank import
   ↓
package initialization
   ↓
init()
   ↓
Register(...)
   ↓
registry now knows constructor
   ↓
config can request connector
~~~

This is one of the stranger patterns for Python developers.

But once understood, it explains how config-driven plugins appear "magically."

---

# 33. Construction of the Go2 connector

NewMoveConnector:
- reads config
- creates connector struct
- gets providers
- opens Zenoh
- creates publisher/subscriber handles
- configures optional guard mode
- returns actions.Connector

This is where external middleware resources are acquired.

Therefore debugging a movement connector often begins by asking:

> Did initialization successfully establish the required publishers/subscribers?

---

# 34. Connect does not directly drive motors continuously

A high-level action arrives at Connect.

It validates:
- input shape
- guard status
- AI-control enablement
- robot movement status
- existing pending command
- localization readiness

Then it queues a movement plan.

The recurring Tick method advances that plan.

This is an important architectural distinction:

~~~text
Connect
  = accept/interpret new decision

Tick
  = progress long-running connector behavior
~~~

Not every connector has to use Tick heavily, but this one does.

---

# 35. Safety checks are part of action execution

The movement connector checks things such as:
- safe paths
- robot state
- whether motion is already active
- whether odometry exists
- whether progress stalls
- command timeout

This illustrates an important AI robotics principle:

> The LLM chooses high-level intent, but deterministic code should enforce operational constraints and safety conditions.

Do not make the LLM the low-level servo controller.

---

# 36. From high-level action to cmd_vel

For a command such as:

~~~text
move forwards
~~~

the path is conceptually:

~~~text
LLM ToolCall
   ↓
Action Orchestrator
   ↓
moveConnector.Connect
   ↓
queue movement state
   ↓
moveConnector.Tick
   ↓
calculate velocity
   ↓
serialize Twist
   ↓
Zenoh Publisher.Put
   ↓
cmd_vel
   ↓
downstream robot/ROS-facing stack
~~~

This is the concrete version of the architecture you were verbally practicing.

---

# 37. Serialization boundary

The connector eventually creates a byte payload for a Twist-style command and calls the Zenoh publisher.

This is where:
- semantic intent
becomes:
- transport representation

The LLM never needs to know:
- CDR bytes
- exact middleware serialization
- packet format

The connector owns that complexity.

That is a clean abstraction boundary.

---

# 38. What "Zenoh" means in this code

Open:

> internal/zenoh/session.go

OM1 defines interfaces:
- Session
- Publisher
- Subscriber

Again, the runtime depends on abstract behavior.

A Session can:
- declare publishers
- declare subscribers
- put data
- close

A Publisher can:
- put bytes
- drop resource

This wraps middleware specifics behind small interfaces.

Module 2 again.

---

# 39. Middleware abstraction helps testing and deployment

If business logic directly called deep Zenoh APIs everywhere, it would be harder to:
- test
- swap modes
- simulate
- centralize configuration
- support cloud/hybrid behavior

A wrapper lets the application reason at a cleaner level.

This is another example of FDE-relevant architecture:

> isolate environment-specific dependencies behind boundaries.

---

# 40. End-to-end debugging map

If the robot does not move after a voice command, do not randomly edit code.

Walk the boundaries.

## A — user/input

Did the voice input arrive?

Evidence:
- input logs
- ASR output
- sensor latest buffer

## B — Fuser

Does the prompt contain the intended user request and relevant state?

Evidence:
- cortex tick prompt/log
- Fuser sections

## C — model

Did the LLM return the correct ToolCall?

Evidence:
- response/tool-call trace

## D — action parser

Was the tool name registered and parsed?

Evidence:
- ParseCalls errors
- action map

## E — connector

Did Connect receive the action?

Evidence:
- connector log such as AI command

## F — connector guard/safety

Did it intentionally reject the action?

Evidence:
- "already moving"
- "waiting for location data"
- barrier/safety logs
- AI control disabled

## G — recurring execution

Was pending movement queued and Tick progressing?

Evidence:
- movement state/logs
- timing
- stall/timeout logs

## H — middleware

Was cmd_vel published successfully?

Evidence:
- publisher existence
- Put error
- downstream topic visibility

## I — robot stack/hardware

Did the receiving ROS/HAL/controller path consume the command?

Evidence:
- ROS topic inspection
- node graph
- downstream logs
- actuator state

This boundary-by-boundary method is far more valuable than memorizing source.

---

# 41. Where ROS2 tools fit

If the downstream deployment exposes ROS2 topics, you may use:
- ros2 topic list
- ros2 topic info
- ros2 topic hz
- ros2 topic echo
- rqt_graph
- RViz2
- rosbag2

But remember the names:

**rqt_graph**
> graph of nodes/topics

**RViz2**
> spatial/robot data visualization

Earlier, mixing those two is an easy mistake.

The tool you use depends on what evidence you need.

---

# 42. Model versus middleware debugging

Suppose the model returns:

~~~text
move forwards
~~~

correctly, but no cmd_vel is published.

That is probably not an LLM reasoning problem.

Focus downstream:
- parser
- connector
- safety gate
- pending command
- Tick
- Zenoh publisher

Suppose cmd_vel is visible and correct, but robot does not move.

Focus farther downstream:
- bridge
- ROS subscriber
- controller
- safety interlock
- firmware/hardware

This is the FDE skill:

> localize first, then investigate.

---

# 43. Read Runtime.tick aloud

A strong explanation:

> "Each cortex tick snapshots the current sensor buffers, checks whether the current observations imply a mode transition, fuses the current context into a prompt, calls the LLM, resolves MCP calls if configured, then routes structured action calls through the action orchestrator. Finally it records memory and telemetry."

If you can say that, you understand the core loop.

---

# 44. Read action execution aloud

A strong explanation:

> "The model returns structured ToolCalls. The runtime resolves their names against registered AgentActions. The action orchestrator applies the configured execution policy, and each AgentAction delegates to a concrete Connector. The connector is where high-level intent becomes a middleware message, API call, TTS operation, or robot behavior."

That is a very strong OM1 explanation.

---

# 45. Read the Go2 path aloud

A strong explanation:

> "The Go2 move connector receives a structured movement action, validates runtime and safety state, queues a movement command, and its Tick loop progresses the command using odometry and safe-path information. It serializes velocity commands and publishes them through the Zenoh cmd_vel publisher."

That is much more precise than:

> "MCP sends JSON back to ROS."

---

# 46. What belongs to Go versus what belongs to OM1

Separate them.

## Go concepts

- package
- struct
- interface
- pointer receiver
- map/slice
- type assertion
- goroutine
- channel
- select
- context
- mutex
- WaitGroup
- errors
- defer

## OM1 architecture

- config-driven plugins
- sensors
- input orchestrator
- Fuser
- Cortex
- ToolCalls
- action orchestrator
- connectors
- MCP
- memory
- modes
- Zenoh abstraction
- robot-specific plugins

Do not blame the language for architecture complexity.

---

# 47. Whiteboard exercise

Without looking at this page, draw:

~~~text
Voice / sensors
      ↓
Input plugins
      ↓
Buffers
      ↓
Fuser
      ↓
Cortex
      ↓
ToolCalls
      ↓
Action Orchestrator
      ↓
Connector
      ↓
Zenoh / service / robot
~~~

Then add:
- where goroutines exist
- where interfaces exist
- where dynamic maps exist
- where errors can occur
- where a mutex may be necessary

If you can do that, you are integrating language and system design.

---

# 48. Code-reading exercise

Open these files in exactly this order:

1. cmd/main.go
2. internal/runtime/runtime.go
3. internal/inputs/orchestrator.go
4. internal/fuser/fuser.go
5. internal/llm/llm.go
6. internal/llm/orchestrator.go
7. internal/actions/action.go
8. internal/actions/orchestrator.go
9. plugins/actions/unitree/go2/autonomy/move.go
10. internal/zenoh/session.go

For each file, answer only:

1. What responsibility does this file have?
2. What important types are defined here?
3. What does this file receive?
4. What does it output/call next?
5. What Go construct is most important here?

Do not try to explain every line.

---

# 49. Self-test

### Question 1
Where does the executable start?

### Question 2
What does initializeMode do conceptually?

### Question 3
What wakes the Cortex loop?

### Question 4
What exactly does the Fuser output?

### Question 5
What does Cortex return?

### Question 6
What is the purpose of ParseCalls?

### Question 7
What is the difference between the action orchestrator and a connector?

### Question 8
Where does MCP fit?

### Question 9
In the Go2 connector, what is the difference between Connect and Tick?

### Question 10
At what point does a high-level action become serialized middleware bytes?

---

# 50. Answers

### 1
cmd/main.go, in package main and func main.

### 2
It builds the active mode's dependency graph: sensors, Fuser, LLM orchestrator, action orchestrator, optional MCP/memory/knowledge/backgrounds, and lifecycle state.

### 3
Periodic timing, input-triggered TickNow signals, mode-context updates, and cancellation controls the loop lifecycle.

### 4
A single prompt string containing the current agent context and capabilities.

### 5
An llm.Response containing text content, structured tool calls, and usage information.

### 6
Resolve symbolic tool-call names/arguments into internal action Calls pointing to registered AgentAction objects.

### 7
The orchestrator schedules/routes actions. A connector implements the concrete side effect/integration.

### 8
As an optional tool integration path in the cortex cycle; it is not inherently the ROS translation layer.

### 9
Connect accepts/interprets a new high-level action. Tick progresses ongoing movement behavior over time.

### 10
Inside the concrete connector near the middleware boundary; in the Go2 example, velocity data is serialized before Publisher.Put.

---

# 51. Finish line

You now have the architecture-level Go understanding this course was aiming for.

Next:

> 06_PRACTICAL_GO_AND_FDE_INTERVIEW.md

That module shifts from:

> "Can I read this?"

to:

> "Can I reason about it under interview pressure?"
