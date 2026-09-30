# Module 2 — Structs, Methods, Pointers, Interfaces, and Composition

This module answers the question that often bothers Python developers most:

> If Go does not really use classes the way Python does, how is a real application like OM1 organized?

The short answer is:

> Go separates state, behavior, and contracts more explicitly.

In Python, one class can simultaneously mean:
- a bag of data
- a namespace
- a constructor
- an inheritance point
- an interface
- a behavior implementation

Go tends to split those jobs among:
- structs
- methods
- constructor-like functions
- interfaces
- composition

That separation is one reason OM1 can be modular without having a giant inheritance tree.

---

# 1. Start with the three nouns

Keep these three concepts separate.

## Struct

A struct stores fields.

Tiny example:

~~~go
type Robot struct {
    Name    string
    Battery float64
}
~~~

Python mental model:

~~~python
@dataclass
class Robot:
    name: str
    battery: float
~~~

Do not read struct as "full Python class."

Read it first as:

> a named data shape.

---

## Method

A method is a function with a receiver.

~~~go
func (r *Robot) Stop() {
    // ...
}
~~~

Read:

> Stop is behavior associated with Robot.

Python:

~~~python
class Robot:
    def stop(self):
        ...
~~~

The receiver variable r plays roughly the role of self.

---

## Interface

An interface describes behavior that a value must provide.

~~~go
type Speaker interface {
    Speak(text string) error
}
~~~

Read:

> Anything with a compatible Speak method can be treated as a Speaker.

The type does not normally declare "implements Speaker."

Go checks the method set.

That is one of the most important ideas in this repository.

---

# 2. Why this matters in OM1

Open:

> internal/actions/action.go

OM1 defines a Connector interface.

Conceptually, the contract is:

~~~go
type Connector interface {
    Connect(ctx context.Context, input Input) (Output, error)
    Tick(ctx context.Context)
    Stop()
}
~~~

This tells you something architectural immediately.

The action orchestrator does not need to know whether an action:
- publishes a Zenoh message
- calls HTTP
- speaks through TTS
- moves a Unitree robot
- invokes some other service

It only needs a value that can:
- execute one decision
- perform recurring tick work if needed
- stop cleanly

That is polymorphism without an inheritance hierarchy.

---

# 3. Python inheritance versus Go interfaces

A Python design might look like:

~~~python
class Connector(ABC):
    @abstractmethod
    def connect(self, ctx, input):
        ...

class MoveConnector(Connector):
    ...
~~~

A Go implementation does not need to inherit from Connector.

It can simply define:

~~~go
type moveConnector struct {
    // state
}

func (c *moveConnector) Connect(ctx context.Context, input actions.Input) (actions.Output, error) {
    // ...
}

func (c *moveConnector) Tick(ctx context.Context) {
    // ...
}

func (c *moveConnector) Stop() {
    // ...
}
~~~

If those methods match the Connector interface, the type satisfies it.

The relationship is implicit.

This has a practical benefit:

> The concrete type stays focused on what it does instead of declaring a family tree.

---

# 4. A useful sentence for interviews

If someone asks what you have learned about Go interfaces, a strong answer is:

> "Go interfaces are structural. A type satisfies an interface by having the required methods, rather than explicitly inheriting from or declaring that interface. OM1 uses that heavily for sensors, LLM providers, action connectors, and middleware abstractions."

That is accurate and directly tied to the code.

---

# 5. Pointer receiver versus value receiver

This is where Python programmers often get stuck.

Compare:

~~~go
func (r Robot) NameText() string
~~~

with:

~~~go
func (r *Robot) Stop()
~~~

The first receiver is a Robot value.

The second receiver is a pointer to Robot.

For practical reading, ask:

> Does this method need to modify the same object or work with shared state?

If yes, pointer receivers are very common.

A pointer receiver also avoids copying a potentially large struct.

---

# 6. Python hides more reference behavior

In Python:

~~~python
robot.stop()
~~~

you do not visibly write whether robot is passed by pointer.

Objects already behave reference-like.

Go makes the distinction more explicit.

That is why this:

~~~go
func (c *moveConnector) queue(...)
~~~

looks unusual at first.

A useful translation is:

> queue operates on the actual connector instance and its shared state.

---

# 7. Read a real OM1 struct

Open:

> plugins/actions/unitree/go2/autonomy/move.go

The move connector contains state such as:

- logger
- odometry provider
- path provider
- Zenoh publishers/subscribers
- control enabled flag
- mode
- random generator
- mutex
- pending movement command

This is not merely a record.

It is the state needed by a long-lived action component.

Its methods operate on that state.

So the Go pattern is:

~~~text
struct
  +
methods with pointer receiver
  =
stateful component
~~~

That is very close to what you would use a Python object for.

---

# 8. Constructor-like functions are ordinary functions

Go does not have a special constructor syntax.

OM1 uses names like:

- NewMoveConnector
- NewOrchestrator
- NewFuser
- NewMessage

Example concept:

~~~go
func NewThing(...) *Thing {
    return &Thing{...}
}
~~~

Read:

> ordinary function that creates and initializes a Thing.

Python:

~~~python
thing = Thing(...)
~~~

There is no hidden constructor magic in the name New.

It is a convention.

---

# 9. Why return an interface from a constructor?

The real move constructor returns an actions.Connector rather than a concrete moveConnector type.

Conceptually:

~~~go
func NewMoveConnector(cfg map[string]any) (actions.Connector, error)
~~~

That is important.

The caller does not need the private implementation type.

The caller only needs:

> something that satisfies Connector.

This hides implementation details and keeps the runtime decoupled.

Python analogy:

~~~python
def new_move_connector(cfg) -> Connector:
    return MoveConnector(...)
~~~

---

# 10. Private implementation types

Notice the type name:

~~~text
moveConnector
~~~

starts lowercase.

That makes it unexported outside its package.

The constructor is capitalized:

~~~text
NewMoveConnector
~~~

so other packages can call it.

This is a common Go pattern:

~~~text
public constructor
       ↓
private implementation struct
       ↓
returned through public interface
~~~

That is a clean API boundary.

---

# 11. Composition instead of inheritance

Suppose a component needs:
- a logger
- an odometry provider
- a paths provider
- a publisher

Go commonly stores those as fields:

~~~go
type moveConnector struct {
    log   *zap.Logger
    odom  *OdomProvider
    paths *PathsProvider
    pub   Publisher
}
~~~

This is composition.

The object gets behavior by containing collaborators, not by inheriting from a base class.

Python can do this too, but Go culture leans heavily toward it.

A useful mantra:

> has-a instead of is-a.

A move connector:
- has a publisher
- has an odometry provider
- has a path provider

It does not need to inherit from each of them.

---

# 12. Dependency injection in OM1

Dependency injection sounds abstract, but the idea is simple:

> Give a component what it needs instead of having it secretly create everything itself.

OM1 does this in several forms.

For example, the Fuser stores collaborators:

- runtime configuration
- agent actions
- knowledge base
- memory manager
- MCP describer
- logger

Then NewFuser receives them.

Conceptually:

~~~go
f := NewFuser(config, actions, kb, memory, mcp, log)
~~~

Python analogy:

~~~python
f = Fuser(
    config=config,
    actions=actions,
    knowledge_base=kb,
    memory=memory,
    mcp=mcp,
    log=log,
)
~~~

Why is this good?

Because Fuser does not need to know how to construct:
- a knowledge base
- a memory system
- an MCP client

It only knows how to use their interfaces.

That improves:
- modularity
- testability
- replacement
- separation of concerns

---

# 13. Interface as "minimum required behavior"

Open:

> internal/fuser/fuser.go

The Fuser declares small interfaces for what it needs from collaborators.

For example, KnowledgeBase conceptually promises:

~~~go
Query(ctx, question, topK) ([]string, error)
~~~

The Fuser does not need every method of the concrete knowledge-base implementation.

It needs only Query.

This is an important Go design principle:

> Prefer small interfaces close to the consumer.

You do not need a huge interface with twenty methods if this component only needs one.

---

# 14. Interface values: two layers to remember

An interface variable has two conceptual pieces:

~~~text
interface value
   ├─ concrete type
   └─ concrete value
~~~

For example:

~~~go
var c Connector
c = someMoveConnectorPointer
~~~

The variable c has interface type Connector.

Inside it is a concrete value whose type might be:

~~~text
*moveConnector
~~~

Then when code says:

~~~go
c.Connect(...)
~~~

Go dispatches to the concrete type's method.

This is dynamic dispatch, but based on interfaces rather than class inheritance.

---

# 15. Type assertions

When a value has interface type, sometimes code wants to ask:

> What additional capability does the concrete value have?

From Module 1, OM1 uses the pattern:

~~~go
t, ok := sensor.(TickTrigger)
~~~

Read:

> Does the concrete value inside sensor also satisfy TickTrigger?

If yes:
- ok is true
- t can be used as a TickTrigger

This is not a dictionary lookup.

It is a type/interface capability check.

---

# 16. The optional-capability pattern

This is elegant and important.

OM1 has a required Sensor contract.

Then it has an optional TickTrigger interface.

That means:

> Every input must be a Sensor, but only some inputs need to say whether new data should immediately wake the cortex loop.

This avoids bloating Sensor with a method that not every implementation conceptually needs.

The pattern generalizes:

~~~text
base required interface
+
small optional capability interfaces
~~~

This is a very Go-like design.

---

# 17. What is any?

The keyword any is an alias for the empty interface.

For practical purposes:

~~~go
any
~~~

means:

> a value whose concrete type is not known statically at this point.

OM1 uses it around:
- config
- generic sensor data
- action input/output
- tool arguments

This gives flexibility, but you lose compile-time knowledge about the concrete type.

That is why type assertions appear later.

---

# 18. A real example: action input

OM1 defines:

~~~go
type Input any
~~~

That means an action's input can hold arbitrary structured data.

Then a connector may do:

~~~go
args, ok := input.(map[string]any)
~~~

Read:

> I expect the concrete value inside input to be a map from strings to arbitrary values. Check whether that is true.

If false, the connector returns an error.

This is where a dynamic JSON-like world meets static Go code.

---

# 19. Why not make every action a different static input type?

You could imagine a highly static design where every connector has a different method signature.

But the action orchestrator wants one uniform contract:

~~~text
Connector.Connect(context, generic input)
~~~

because it is routing many action types.

The dynamic boundary occurs at the orchestrator/connector edge.

Inside a specific connector, the code can validate and interpret that generic input.

This is a common architecture:

~~~text
generic routing layer
        ↓
type-specific implementation
~~~

---

# 20. Typed nil versus nil: the subtle version

This is one of the few Go concepts that can surprise experienced programmers.

An interface value can conceptually contain:

~~~text
concrete type: *Thing
concrete value: nil
~~~

In that case, the interface itself may not compare equal to nil because it still contains type information.

You do not need to master every edge case for the interview.

The practical lesson is:

> Be careful when storing nil pointers inside interfaces. Interface nilness is slightly more subtle than Python None.

If you see confusing nil behavior later, remember this section.

---

# 21. Method sets: enough to read OM1

Go has precise rules about which methods belong to T versus *T.

For this course, remember:

- a method declared on T belongs to the value type's method set
- a method declared on *T belongs to the pointer type's method set
- pointer receiver methods are common for stateful components

If an interface expects a method that is only implemented on *Thing, then you usually provide a pointer to Thing.

That is why constructors often return:

~~~go
&Thing{...}
~~~

rather than:

~~~go
Thing{...}
~~~

---

# 22. Why OM1 uses pointers for orchestrators

Look at types such as:
- Runtime
- Orchestrator
- Fuser
- moveConnector

These represent long-lived components with shared state.

Typical state includes:
- logs
- buffers
- current mode
- history
- synchronization primitives
- publishers/subscribers
- pending actions

Copying these objects around would be undesirable.

Pointer receivers communicate:

> this is one shared component instance.

That mental model is more useful than thinking about raw memory addresses.

---

# 23. Struct embedding versus named fields

Go also supports embedding, where one type is placed into another without a field name.

You may encounter this in other Go projects.

Example:

~~~go
type LoggedThing struct {
    BaseThing
}
~~~

OM1 frequently uses named fields instead.

Named fields are often easier to read because dependencies are explicit:

~~~go
type Fuser struct {
    memory MemoryManager
    log    *zap.Logger
}
~~~

For your first five hours, focus more on named composition than on inheritance-like interpretations of embedding.

---

# 24. Fuser as a case study

Open:

> internal/fuser/fuser.go

The Fuser struct is a good object-model example.

It contains:
- configuration
- actions
- knowledge-base interface
- memory interface
- MCP-describer interface
- logger

Its method:

~~~text
Fuse
~~~

uses that stored state plus current sensor buffers to build a prompt.

Python mental model:

~~~python
class Fuser:
    def __init__(self, runtime_config, actions, kb, memory, mcp, log):
        self.runtime_config = runtime_config
        ...

    def fuse(self, ctx, sensor_buffers):
        ...
~~~

So yes, a Go struct plus receiver methods can fill a role very similar to a Python class.

But Go's interface relationships are decoupled from that struct definition.

---

# 25. Action connector as a case study

Open:

> internal/actions/action.go

Then:

> plugins/actions/unitree/go2/autonomy/move.go

The architecture is:

~~~text
Connector interface
        ↑
moveConnector methods happen to satisfy it
        ↑
NewMoveConnector returns it as Connector
        ↑
registry stores constructor
        ↑
config selects it
        ↑
action orchestrator calls it
~~~

This is the entire plugin mechanism from an object-model perspective.

---

# 26. Why this design is useful for an FDE

Suppose a customer's robot does not use Unitree's movement path.

Maybe they need:
- a different ROS topic
- a vendor SDK
- HTTP
- serial
- proprietary middleware

The central runtime should not be rewritten.

Instead, you can create a new connector satisfying the same Connector interface.

That is the engineering value of the abstraction.

The FDE mental model is:

> preserve the stable interface, swap the integration implementation.

---

# 27. "Class" vocabulary: what to say and what not to say

When learning privately, it is okay to think:

> struct + methods ≈ class

But in an interview, use Go terminology.

Prefer:

> "This struct holds the connector state, and these pointer-receiver methods implement the Connector interface."

rather than:

> "This class inherits from Connector."

The second sentence would be technically wrong in Go.

---

# 28. A complete miniature example

Imagine a robotics runtime.

~~~go
type Actuator interface {
    Execute(command string) error
}

type Motor struct {
    Port string
}

func (m *Motor) Execute(command string) error {
    // send command to hardware
    return nil
}

func NewMotor(port string) Actuator {
    return &Motor{Port: port}
}
~~~

Translate:

~~~python
class Actuator(Protocol):
    def execute(self, command: str) -> None: ...

class Motor:
    def __init__(self, port):
        self.port = port

    def execute(self, command):
        ...

def new_motor(port) -> Actuator:
    return Motor(port)
~~~

The important point:

Motor never says:

~~~text
implements Actuator
~~~

It simply has the required method.

---

# 29. Design exercise: add another implementation

Suppose you define:

~~~go
type HTTPActuator struct {
    URL string
}

func (h *HTTPActuator) Execute(command string) error {
    // POST to service
    return nil
}
~~~

It also satisfies Actuator.

Code that depends only on Actuator does not care which implementation it receives.

That is the essence of substitutability in this style.

---

# 30. Checkpoint: explain these lines

From internal/actions/action.go:

~~~go
type Factory func(cfg map[string]any) (Connector, error)

var connectorRegistry = map[string]Factory{}
~~~

Strong explanation:

> Factory is a named function type for connector constructors. Each constructor accepts a decoded configuration map and returns a Connector plus an error. The registry maps configuration keys to those constructors, so OM1 can choose implementations dynamically while the rest of the runtime depends only on the Connector interface.

If you can say that, you understand the architecture, not just the syntax.

---

# 31. Checkpoint: explain NewMoveConnector

You should be able to say:

> NewMoveConnector is a constructor-like function. It reads dynamic configuration values, creates a moveConnector, opens its Zenoh session and publishers/subscribers, and returns the concrete connector through the actions.Connector interface.

You do not need to memorize every field.

---

# 32. Checkpoint: why the mutex field belongs in the struct

The move connector contains shared state for a pending movement command.

More than one goroutine/path can interact with connector state.

Therefore the connector stores a mutex alongside the protected state.

Read:

~~~text
mu + pending
~~~

as:

> synchronization mechanism plus shared mutable state.

Module 4 will make this precise.

---

# 33. Mental model summary

When reading a Go codebase, ask:

1. What structs represent state?
2. What methods operate on that state?
3. Which receivers are pointers?
4. What interfaces define replaceable behavior?
5. Which concrete types satisfy those interfaces?
6. Where are those concrete values constructed?
7. Are they returned as concrete types or interfaces?
8. Which collaborators are composed into each struct?
9. Where does dynamic data cross into typed code?

Answer those questions and much of the architecture becomes visible.

---

# 34. Self-test

Try these without looking back.

### Question 1
Why does Go not need an implements keyword for normal interface satisfaction?

### Question 2
What is the practical difference between a struct and an interface?

### Question 3
Why are pointer receivers common on OM1 runtime components?

### Question 4
Why can NewMoveConnector return actions.Connector even though it constructs a moveConnector?

### Question 5
What does this mean?

~~~go
args, ok := input.(map[string]any)
~~~

### Question 6
Why is composition useful for field deployment integrations?

---

# 35. Answers

### 1
Because Go uses structural interface satisfaction. If the concrete type has the interface's required methods, it satisfies the interface.

### 2
A struct describes stored fields/state. An interface describes required behavior/methods.

### 3
Those components are long-lived shared objects with mutable state, resources, synchronization, and collaborators. Pointer receivers operate on the same underlying instance and avoid copying it.

### 4
Because the concrete pointer satisfies the Connector interface. The caller only needs the contract.

### 5
It asks whether the concrete value stored inside the generic input has type map[string]any. If yes, args contains the map and ok is true.

### 6
A stable interface lets you replace the robot/customer-specific implementation without rewriting the core runtime.

---

# 36. Finish line

You are ready for Module 3 when you can look at:

> internal/actions/action.go

and explain it as:

> interface + action metadata + constructor function type + registry + loader + custom error.

Next:

> 03_DATA_ERRORS_AND_TOOL_CALLS.md
