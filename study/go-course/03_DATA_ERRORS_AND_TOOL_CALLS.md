# Module 3 — Data Plumbing, Dynamic Values, Errors, and Tool Calls

This module teaches the Go needed to answer:

> How does structured information move through OM1?

The central example is especially relevant to the product:

~~~text
LLM response
  ↓
ToolCall
  ↓
[]ToolCall
  ↓
action parser
  ↓
Call
  ↓
Connector.Connect
  ↓
robot/service-specific behavior
~~~

This module is where Python's flexible dictionaries meet Go's static type system.

---

# 1. Start with the real LLM data types

Open:

> internal/llm/llm.go

OM1 defines conceptual shapes like:

~~~go
type ToolCall struct {
    Name      string
    Arguments map[string]any
}

type Response struct {
    TextContent string
    ToolCalls   []ToolCall
    Usage       Usage
}
~~~

This is extremely important.

The model response is not just "text."

It can contain:
- textual response content
- zero or more structured tool calls
- usage information

That is the bridge from language-model reasoning to executable behavior.

---

# 2. Read ToolCall as Python

Go:

~~~go
type ToolCall struct {
    Name      string
    Arguments map[string]any
}
~~~

Python mental model:

~~~python
@dataclass
class ToolCall:
    name: str
    arguments: dict[str, Any]
~~~

Example conceptual value:

~~~text
Name = "unitree_go2_autonomy"
Arguments = {
    "action": "move forwards"
}
~~~

The tool name is typed as string.

The arguments are intentionally dynamic because different tools accept different schemas.

---

# 3. Why map[string]any appears so often

Suppose one action expects:

~~~json
{"text": "hello"}
~~~

another expects:

~~~json
{"action": "move forwards"}
~~~

another expects:

~~~json
{"x": 1.2, "y": 0.5, "frame": "map"}
~~~

A single routing layer cannot use one fixed struct for all of those shapes.

So it often carries:

~~~go
map[string]any
~~~

That means:

> string keys, arbitrary concrete values.

It is the Go equivalent of a JSON-like Python dictionary.

---

# 4. Static outside, dynamic at boundaries

A mature Go system often has a useful pattern:

~~~text
strongly typed core
      ↕
dynamic boundary
      ↕
JSON / config / tool args / external data
~~~

OM1 does exactly this.

Dynamic data appears at places such as:
- config maps
- LLM tool arguments
- generic input/output boundaries

Then concrete components validate and interpret it.

This is not "Go giving up on typing."

It is Go acknowledging that external data is genuinely dynamic.

---

# 5. Type assertion revisited

A Unitree connector receives generic action input.

It does something conceptually like:

~~~go
args, ok := input.(map[string]any)
if !ok {
    return nil, fmt.Errorf("unexpected input type")
}
~~~

Translation:

> The interface says this could be anything. For this connector, I require a dictionary-like map. Check that assumption before using it.

Then:

~~~go
action, _ := args["action"].(string)
~~~

Translation:

> Get the value under "action" and try to interpret it as a string.

The underscore intentionally ignores the boolean success result.

In production code, sometimes you may want stricter validation.

---

# 6. The comma-ok pattern has several forms

You have already seen map lookup:

~~~go
value, ok := myMap[key]
~~~

You also see type assertion:

~~~go
text, ok := value.(string)
~~~

Same broad pattern:

> return the thing plus whether the operation succeeded.

This is very Go.

It avoids:
- hidden exceptions
- magic sentinel values
- ambiguity

---

# 7. Slices carry multiple items

OM1's response contains:

~~~go
ToolCalls []ToolCall
~~~

Read:

> a slice of ToolCall values.

The model might return:
- zero actions
- one action
- multiple actions

The runtime can iterate through the slice.

Python:

~~~python
for call in response.tool_calls:
    ...
~~~

Go:

~~~go
for _, call := range response.ToolCalls {
    ...
}
~~~

---

# 8. make versus append

There are two common ways to build slices.

## Allocate known length

~~~go
results := make([]Result, len(calls))
~~~

This gives exactly len(calls) slots.

Useful when each result belongs at a known index.

Then:

~~~go
results[i] = result
~~~

## Start empty and append

~~~go
var results []Result
results = append(results, result)
~~~

or:

~~~go
results := make([]Result, 0, len(calls))
~~~

The second form starts with length zero but reserves capacity.

Python uses append far more uniformly.

Go gives you more explicit control.

---

# 9. Length versus capacity

A slice has:
- length: how many elements currently exist
- capacity: how much backing storage can fit before reallocation

Example:

~~~go
make([]Result, 0, 10)
~~~

means:
- length 0
- capacity 10

You can append up to that capacity before the underlying storage needs to grow.

You do not need to optimize this manually in ordinary interview code, but recognize what the third argument means.

---

# 10. Arrays versus slices again

Go array:

~~~go
[3]int
~~~

The size 3 is part of the type.

Go slice:

~~~go
[]int
~~~

Dynamic-length view over underlying storage.

OM1 mostly uses slices where variable collections are needed.

One real fixed array appears in the Go2 movement code for path angles.

That is appropriate because the number of entries is fixed.

---

# 11. Custom named scalar types

OM1 contains a type like:

~~~go
type MoveAction string
~~~

This creates a distinct named type whose underlying representation is string.

Why do this instead of using plain string everywhere?

Because it gives semantic meaning.

A MoveAction is not just any random string in the programmer's mental model.

It represents one of the movement commands.

Named types can also have methods.

---

# 12. Methods on non-struct types

This surprises Python developers.

A Go method can be defined on a named type such as:

~~~go
type MoveAction string
~~~

Then:

~~~go
func (MoveAction) EnumValues() []string {
    ...
}
~~~

So methods are not restricted to struct classes.

The receiver must be a defined type in the same package, but that type can have an underlying scalar representation.

This is another reason "Go method = Python class method" is only an approximation.

---

# 13. Struct tags

The move input type includes metadata similar to:

~~~go
type MoveInput struct {
    Action MoveAction `json:"action" description:"The movement to perform"`
}
~~~

The text after the field type is a struct tag.

Libraries can inspect tags using reflection.

Common uses:
- JSON field names
- validation
- database mapping
- schema generation

Python analogies include:
- dataclass field metadata
- Pydantic Field metadata
- serializer annotations

In OM1, tags help describe action schemas.

---

# 14. Schemas create a contract for the LLM

The runtime gives the model function/action schemas.

The model then returns a ToolCall with:
- a name
- arguments

This is important conceptually:

> The action call is not arbitrary prose that OM1 must guess how to interpret.

The model is operating against structured tool definitions.

That is why the runtime can route actions reliably.

The conceptual chain is:

~~~text
Go type / action registration
      ↓
schema
      ↓
LLM sees available action shape
      ↓
LLM emits structured call
      ↓
Go parser maps name to action
      ↓
connector gets arguments
~~~

---

# 15. The action parser

Open:

> internal/actions/orchestrator.go

The parser receives raw tool-call-shaped maps.

Conceptually it does:

~~~go
actionName := rawToolCall["name"]
arguments := rawToolCall["arguments"]
agentAction := actions[actionName]
calls = append(calls, Call{
    Action: agentAction,
    Input: arguments,
})
~~~

The job is:

> convert externally structured action data into OM1's internal Call type.

That is an important boundary.

---

# 16. Internal Call is more strongly connected to runtime objects

The raw call conceptually says:

~~~text
name = "move"
arguments = {...}
~~~

After parsing, OM1 has:

~~~go
type Call struct {
    Action *AgentAction
    Input  Input
}
~~~

Now Action is not merely a string.

It points to the actual registered AgentAction, which contains the Connector implementation.

So parsing performs a transition:

~~~text
symbolic action name
       ↓ lookup
actual runtime action object
~~~

That is what lets execution happen.

---

# 17. Map lookup is routing

Inside the orchestrator:

~~~go
agentAction, ok := o.actions[actionName]
~~~

This is not merely data access.

Architecturally, it is dispatch.

The map was created earlier from action labels to action objects.

So the tool-call name determines which connector will eventually execute.

That is a useful debugging boundary:

> Did the model emit the right action name, and does the orchestrator have a registered action under that name?

---

# 18. Unknown action is an explicit error

If the action name is missing from the map, OM1 returns an error.

Conceptually:

~~~go
return nil, fmt.Errorf("unknown action %q", actionName)
~~~

This is safer than silently doing nothing.

A production runtime should make failed assumptions observable.

---

# 19. Error values: the basic contract

The built-in error interface is tiny.

Conceptually:

~~~go
type error interface {
    Error() string
}
~~~

Any type with the right Error method can represent an error.

Common function signature:

~~~go
func DoThing(...) (Result, error)
~~~

Convention:
- success: useful result + nil error
- failure: maybe zero/nil result + non-nil error

---

# 20. Returning versus logging errors

A useful engineering distinction:

## Return the error when the caller should decide what to do

~~~go
value, err := load()
if err != nil {
    return nil, err
}
~~~

## Log and continue when the failure is non-fatal and local policy allows degraded operation

Some OM1 connector initialization paths do this.

For example, if a middleware publisher cannot be created, a connector may log that movement is disabled and continue existing.

That is a product decision, not merely language syntax.

Always ask:

> Is this failure fatal to the whole runtime, or only to one capability?

That is an FDE-relevant question.

---

# 21. Error wrapping

OM1 uses patterns like:

~~~go
return fmt.Errorf("initialize mode %q: %w", modeName, err)
~~~

The percent-w form wraps the original error.

Conceptually:

> add context without losing the underlying cause.

Python analogy:

~~~python
raise RuntimeError(f"initialize mode {mode}") from err
~~~

This matters because good operational errors answer:
- what operation failed?
- for what object?
- what was the original cause?

---

# 22. Sentinel and typed errors

Two common Go patterns are:

## Sentinel

A package-level specific error value.

Conceptually:

~~~go
var ErrNotFound = errors.New("not found")
~~~

Then callers can use errors.Is.

## Typed error

A struct type with fields.

OM1's UnknownPluginError is typed.

That lets the error carry structured information such as:
- kind
- name

Callers can potentially use errors.As to inspect its type.

You do not need to master the errors package deeply yet, but recognize both patterns.

---

# 23. Zero values

Go types have defaults.

Examples:

~~~text
int      -> 0
float64  -> 0
bool     -> false
string   -> ""
pointer  -> nil
map      -> nil
slice    -> nil
interface-> nil
~~~

This matters because a struct can often be useful without initializing every field explicitly.

But be careful:

A nil map cannot accept writes until initialized.

A nil slice can be appended to.

Those two behave differently.

---

# 24. Nil slice versus empty slice

For most iteration and len operations:

~~~go
var xs []string
~~~

and:

~~~go
xs := []string{}
~~~

both have length zero.

But they are not identical in every serialization/reflection context.

For ordinary code reading, treat both as "no elements" unless the distinction matters.

---

# 25. Nil map versus empty map

This is more important.

~~~go
var m map[string]string
~~~

m is nil.

Reading is okay.

Writing:

~~~go
m["x"] = "y"
~~~

will panic.

You need:

~~~go
m = make(map[string]string)
~~~

or:

~~~go
m := map[string]string{}
~~~

Then writes are safe.

---

# 26. Switch is common for dispatch

The move connector uses a switch over action values.

Conceptually:

~~~go
switch action {
case "turn left":
    ...
case "turn right":
    ...
case "move forwards":
    ...
default:
    ...
}
~~~

Python 3.10+ has match, but classic Python often uses if/elif.

Go switch is concise and common.

No automatic fall-through occurs by default.

That is different from C-style switch behavior.

---

# 27. Short variable scope can help contain errors

You may see:

~~~go
if err := publisher.Put(payload); err != nil {
    log.Error(...)
}
~~~

This means:
1. call Put
2. name the returned error err
3. keep err scoped to the if statement
4. log if it is non-nil

This pattern reduces accidental reuse of an old err variable.

---

# 28. JSON encoding

OM1's LLM orchestrator formats tool calls for history using JSON marshaling.

Conceptually:

~~~go
bytes, err := json.Marshal(value)
~~~

This converts a Go value into JSON bytes.

Then:

~~~go
string(bytes)
~~~

converts those bytes to a Go string.

Python analogy:

~~~python
json.dumps(value)
~~~

Again, Go makes byte/string boundaries more explicit.

---

# 29. Bytes versus strings

In Go:

~~~go
[]byte
~~~

means a byte slice.

Strings are immutable text/byte sequences.

Middleware/network code often works with bytes.

For example, Zenoh publishers accept serialized payload bytes.

The connector may:
1. build typed data
2. serialize it
3. publish []byte

That is the boundary from application-level meaning to transport representation.

---

# 30. Structured action to bytes

For a robot movement, conceptually:

~~~text
ToolCall
  Name: movement action
  Arguments: {"action":"move forwards"}
        ↓
ParseCalls
        ↓
Connector.Connect
        ↓
choose movement plan
        ↓
connector Tick loop
        ↓
serialize Twist velocity command
        ↓
[]byte payload
        ↓
Zenoh publisher Put
        ↓
robot-side middleware
~~~

This is the data transformation path you were asking about earlier.

The model never needs to output raw ROS bytes.

It chooses a structured action.

The connector owns the concrete integration details.

---

# 31. Defensive parsing

Whenever code converts dynamic values, ask:

> What happens if the model/config/external system gives the wrong shape?

Example:

~~~go
args, ok := input.(map[string]any)
if !ok {
    return nil, fmt.Errorf(...)
}
~~~

That check is important.

A weaker version might assume the type and panic.

For FDE-quality code, prefer:
- validate inputs
- return useful errors
- log enough context
- avoid crashing the entire process when degradation is acceptable

---

# 32. Avoiding panic as routine control flow

Go has panic and recover, but routine application errors usually use error values.

Do not think:

> Go exceptions = panic.

For normal failures:
- network unavailable
- bad config
- unknown plugin
- malformed action
- file missing

use error values.

Panic is more like:
> something has violated a fundamental invariant and normal recovery is not intended here.

---

# 33. A miniature tool-call pipeline

Imagine:

~~~go
type ToolCall struct {
    Name      string
    Arguments map[string]any
}

type Action struct {
    Handler func(map[string]any) error
}

var actions = map[string]Action{}
~~~

Then:

~~~go
func Execute(tc ToolCall) error {
    action, ok := actions[tc.Name]
    if !ok {
        return fmt.Errorf("unknown tool %q", tc.Name)
    }
    return action.Handler(tc.Arguments)
}
~~~

That is the core idea behind many tool systems.

OM1 adds:
- richer action objects
- connector interfaces
- schemas
- concurrency modes
- context cancellation
- logging

But the central dispatch idea is simple.

---

# 34. Read the LLM Response carefully

The response type has both:

~~~text
TextContent
ToolCalls
~~~

This matters because an LLM can:
- speak natural language
- choose actions
- sometimes do both depending on provider/model behavior

Do not collapse "LLM output" into one plain string in your mental model.

A better mental model:

> Cortex returns a typed response object containing textual output plus zero or more structured calls.

---

# 35. SpeakText is a good reading exercise

OM1 has a helper that checks tool calls for a speak action and otherwise falls back to text content.

This contains several concepts:
- pointer receiver
- range over slice
- string comparison
- map lookup
- type assertion
- early return

When you can read that function comfortably, Module 3 is landing.

---

# 36. Configuration has the same dynamic boundary

JSON5 config is decoded into Go values.

At some point, plugin-specific config is often represented as:

~~~go
map[string]any
~~~

Then constructors pull expected keys from it.

This is analogous to Python receiving a dict from parsed JSON.

Difference:

> Go makes you consciously convert/validate values before treating them as specific types.

That explicitness can improve reliability if used well.

---

# 37. FDE debugging questions at a dynamic boundary

If an action is not happening, inspect in this order:

1. Did the model return a ToolCall?
2. Is the tool name what you expected?
3. Are the arguments present?
4. Are the argument types correct?
5. Does the action map contain that name?
6. Does ParseCalls succeed?
7. Does Connector.Connect receive the input?
8. Does the connector reject it?
9. Does it create/publish the expected downstream command?

This is the "find the first boundary where reality diverges from expectation" pattern.

---

# 38. Exercise: parse this by eye

~~~go
calls, err := current.actionOrchestrator.ParseCalls(toolCallsToMaps(toolCalls))
if err != nil {
    log.Warn("parse action calls failed", zap.Error(err))
    return
}
~~~

Explain:

- toolCalls currently have the llm.ToolCall representation
- helper converts them to generic maps expected by ParseCalls
- ParseCalls resolves names to registered actions
- on failure, this cortex tick stops executing actions

That is architecture plus syntax.

---

# 39. Exercise: why is Result a struct?

OM1 defines a Result containing:
- action name
- output
- error

Why not just return output?

Because multiple calls may be submitted.

The orchestrator needs to associate each outcome with:
- what action ran
- what it returned
- whether it failed

Structs make multi-field results explicit.

---

# 40. Exercise: explain these return styles

A connector may return:

~~~go
return nil, nil
~~~

Meaning:

> no output, but no error.

This can be legitimate for a command whose effect is side-effectful.

It may return:

~~~go
return nil, err
~~~

Meaning:

> no useful output because execution failed.

It may return:

~~~go
return value, nil
~~~

Meaning:

> successful output.

These are not special syntax. They follow the Output,error signature.

---

# 41. Self-test

### Question 1
Why does OM1 use map[string]any for tool arguments?

### Question 2
What does a type assertion do?

### Question 3
What is the architectural purpose of ParseCalls?

### Question 4
Why is a ToolCall more useful than arbitrary output text for robot actions?

### Question 5
What is the practical difference between a nil map and a nil slice?

### Question 6
Why wrap an error?

### Question 7
Where does serialization belong in the model-to-robot path?

---

# 42. Answers

### 1
Different actions have different argument schemas. The routing layer needs a generic JSON-shaped representation.

### 2
It checks/extracts the concrete type held by an interface value, often returning a boolean saying whether the assertion succeeded.

### 3
It translates symbolic model output into actual runtime action objects by looking up action names and attaching their connector implementations.

### 4
A ToolCall has a defined name and structured arguments that can be validated and routed. Free prose would require brittle interpretation.

### 5
A nil slice can safely be appended to; a nil map cannot be written to until initialized.

### 6
To add operational context while preserving the underlying cause for inspection.

### 7
Inside or near the concrete connector/middleware boundary, after high-level intent has been translated into a specific robot command.

---

# 43. Finish line

You are ready for Module 4 when you can explain this chain without hesitation:

~~~text
llm.Response
  → []ToolCall
  → ParseCalls
  → []Call
  → AgentAction
  → Connector.Connect
  → concrete command / service call / middleware message
~~~

Next:

> 04_CONCURRENCY_WITH_OM1.md
