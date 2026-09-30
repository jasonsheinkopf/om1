# Module 6 — Practical Go for an FDE Interview

This module is not about pretending five hours makes you a senior Go engineer.

The target is more useful:

> Be able to read unfamiliar Go, explain what it is doing, make sensible small changes, and debug it systematically.

That is credible for an FDE role.

Your strongest language can still be Python.

---

# 1. What interviewers may actually care about

For an FDE-style role, they may care less about obscure Go trivia than whether you can:

- understand an unfamiliar codebase
- reason about interfaces
- trace data flow
- identify failure boundaries
- handle concurrency carefully
- work in Linux
- debug integrations
- communicate with customers
- make safe changes
- learn fast

So if you see unfamiliar syntax, do not panic.

Use a process.

---

# 2. The six-pass code-reading method

When handed a Go function, do this.

## Pass 1 — identify the signature

Ask:
- function or method?
- receiver?
- arguments?
- returns?

## Pass 2 — identify local state

Look for:
- variables
- maps
- slices
- structs
- pointers

## Pass 3 — identify failure paths

Search visually for:
- err
- nil
- ok
- return
- default cases

## Pass 4 — identify concurrency

Look for:
- go
- chan
- select
- context
- mutex
- WaitGroup
- atomic

## Pass 5 — identify external effects

Look for:
- network calls
- publisher Put
- HTTP
- files
- model calls
- logs
- hardware providers

## Pass 6 — summarize purpose before details

Say:

> "This method receives X, validates Y, updates Z, then calls Q."

Do not start by narrating punctuation.

---

# 3. Example: read Connect

If shown a connector's Connect method, a good first statement is:

> "This is a pointer-receiver method on the connector. It accepts a context and generic action input and returns generic output plus an error. The first thing it does is validate the dynamic input shape, then it checks runtime/safety state and dispatches based on the requested action."

That is already strong.

Then go deeper if asked.

---

# 4. What not to do

Avoid:

> "Uh, there's a star here, then parentheses, then map brackets..."

That communicates syntax uncertainty rather than reasoning.

Even if one symbol is unfamiliar, first extract the function's role.

---

# 5. How to handle an unknown Go API

Suppose you see:

~~~go
thing.DoSomething(x)
~~~

and you do not know the library.

Say:

> "I don't know this API from memory, but based on the type and call site, it appears to return X / mutate Y. I'd jump to the interface or definition to confirm the contract before changing it."

That is professional.

Do not bluff.

---

# 6. Small change exercise: add validation

Suppose action input currently does:

~~~go
action, _ := args["action"].(string)
~~~

A safer version might conceptually validate:

~~~go
action, ok := args["action"].(string)
if !ok || action == "" {
    return nil, fmt.Errorf("missing or invalid action")
}
~~~

What did you add?

- type check
- empty-value validation
- explicit error

That is a realistic small FDE patch.

---

# 7. Small change exercise: add observability

Suppose a connector silently decides not to move.

Bad operational behavior:

~~~go
if blocked {
    return nil, nil
}
~~~

Better:

~~~go
if blocked {
    log.Warn("movement blocked", zap.String("reason", reason))
    return nil, nil
}
~~~

Why?

Because field debugging needs evidence.

The customer sees:
> robot did nothing

You need to know:
> was it intentionally blocked, or did execution fail?

Observability turns behavior into diagnosable behavior.

---

# 8. But do not log everything blindly

More logs are not always better.

Bad:
- huge raw camera data
- secrets
- customer-sensitive payloads
- every 50 ms movement tick at high verbosity in production
- full prompts containing private data without policy review

Good logs answer:
- what subsystem?
- what operation?
- what identifier?
- what state transition?
- what error?
- what measured value?
- what decision?

FDE quality means useful evidence, not noise.

---

# 9. Small change exercise: timeout

Imagine a remote call can hang.

A robust pattern is:

~~~go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()

result, err := remoteCall(ctx)
~~~

Interview explanation:

> "I would bound external operations with a context timeout so one unhealthy dependency cannot stall the whole runtime indefinitely."

That is a strong field-systems answer.

---

# 10. Small change exercise: shared state

Suppose two goroutines access:

~~~go
pending *Command
~~~

Ask:

1. Can both access it concurrently?
2. Can one write?
3. Is there a mutex or ownership rule?
4. Is the check/update atomic as a logical operation?

Do not merely add locks everywhere.

Understand the invariant.

---

# 11. Check-then-act race

This pattern can be unsafe if not protected:

~~~go
if pending == nil {
    pending = newCommand
}
~~~

Two goroutines could both observe nil.

Correctness may require:

~~~go
mu.Lock()
defer mu.Unlock()

if pending == nil {
    pending = newCommand
}
~~~

The lock protects the entire check-and-update sequence.

That is the key concept.

---

# 12. Race detector

A practical command to remember:

~~~text
go test -race ./...
~~~

The race detector instruments code to find many concurrent memory races during tests.

It is not a formal proof.

But it is an excellent debugging tool.

If someone asks how you would investigate a suspected Go data race, mention it.

---

# 13. Tests: the basic shape

Go test files end in:

~~~text
_test.go
~~~

Test functions look like:

~~~go
func TestSomething(t *testing.T) {
    ...
}
~~~

Run:

~~~text
go test ./...
~~~

This runs tests recursively across packages.

---

# 14. OM1 action orchestrator tests

Open:

> internal/actions/orchestrator_test.go

This is a useful test-learning file.

It tests things such as:
- default execution mode
- concurrent results
- sequential ordering
- dependency ordering
- call parsing
- unknown action errors
- lifecycle cancellation

Reading tests is often easier than reading implementation.

Why?

Because tests expose:
- expected inputs
- expected outputs
- invariants

When learning a new codebase:

> implementation tells you how; tests tell you what must remain true.

---

# 15. Table-driven tests

A common Go style is:

~~~go
tests := []struct {
    name string
    input int
    want int
}{
    {"zero", 0, 0},
    {"one", 1, 2},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := f(tt.input)
        if got != tt.want {
            t.Fatalf(...)
        }
    })
}
~~~

Python analogy:
- pytest parameterization

This pattern is idiomatic because structs and slices make test cases concise.

---

# 16. gofmt

Go has a standard formatter.

Command:

~~~text
gofmt -w path/to/file.go
~~~

or many editors format automatically.

This dramatically reduces style debates.

If writing code in an interview, idiomatic formatting helps readability, but do not waste mental energy aligning everything manually.

---

# 17. go vet

Another useful command:

~~~text
go vet ./...
~~~

It performs static checks for suspicious constructs.

It is not a replacement for tests.

Think:
- compiler catches type/syntax errors
- vet catches certain suspicious patterns
- tests check behavior
- race detector checks many concurrency races

---

# 18. go test targeted package

You do not always need the whole repository.

Examples:

~~~text
go test ./internal/actions
go test ./internal/inputs
~~~

During an FDE incident, targeted tests can shorten iteration.

Then run broader validation before merging/deploying.

---

# 19. Go modules

At repo root:

> go.mod

This declares the module and dependencies.

Python analogy:
- pyproject.toml / requirements-style dependency metadata

Commands to recognize:

~~~text
go mod download
go mod tidy
~~~

Do not run tidy casually on a customer branch if you do not intend dependency changes.

It can modify go.mod/go.sum.

---

# 20. go.sum

go.sum stores cryptographic checksums for dependency module versions.

It helps ensure dependency content integrity.

It is not simply a list of imports.

Normally it is committed.

---

# 21. Build command

Common:

~~~text
go build ./...
~~~

or build a command/package target.

For a field deployment:
- compile failure
- test failure
- runtime failure

are distinct stages.

Always know which stage is failing.

---

# 22. Read compiler errors carefully

Go compiler errors are often direct.

Examples:
- unused import
- undefined symbol
- type mismatch
- wrong number of return values
- interface not satisfied

Do not treat compiler messages as noise.

They often tell you exactly which contract you violated.

---

# 23. Interface-not-satisfied errors are informative

If a type no longer satisfies Connector because a method signature changed, the compiler can catch it.

That is a major advantage of static interfaces.

In a dynamic language, the mismatch might appear only at runtime.

This is one reason Go can be attractive for long-running infrastructure/robotics services.

---

# 24. A realistic interview reading prompt

Imagine they show:

~~~go
func (o *Orchestrator) Start(ctx context.Context) <-chan struct{} {
    done := make(chan struct{})

    for _, action := range o.actions {
        o.wg.Add(1)
        go func(c Connector) {
            defer o.wg.Done()
            defer c.Stop()

            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    c.Tick(ctx)
                }
            }
        }(action.Connector)
    }

    go func() {
        o.wg.Wait()
        close(done)
    }()

    return done
}
~~~

You should explain:

> "Start creates a completion channel, launches one goroutine per connector, and tracks them with a WaitGroup. Each worker repeatedly ticks its connector until the context is canceled. Stop and WaitGroup.Done are deferred for cleanup. Another goroutine waits for all workers and closes the done channel, so the caller can observe lifecycle completion."

That is excellent.

---

# 25. Then identify one engineering question

A good follow-up:

> "I'd also inspect whether each Connector.Tick blocks or throttles appropriately, because this loop has a default select case and could spin if Tick returns immediately."

That demonstrates deeper concurrency reasoning.

Do not present it as a bug without evidence.

Present it as something you would verify.

---

# 26. Whiteboard system debugging prompt

Interviewer:

> "The robot works in the lab but drops commands at the customer site. What do you do?"

Strong framework:

1. safety
2. reproduce and define expected versus actual
3. isolate the layer
4. collect evidence at boundaries
5. form a hypothesis
6. change one thing
7. verify
8. stress/robustness test
9. preserve observability/runbook

Then map to OM1 boundaries:
- input
- Fuser/context
- model/tool call
- action parser
- connector
- Zenoh/network
- ROS/HAL
- actuator/hardware

---

# 27. A Go-specific version of the same debugging prompt

Suppose commands stop after ten minutes.

Possible hypotheses:
- goroutine exited
- context accidentally canceled
- goroutine leak/resource exhaustion
- channel blocked
- middleware session dropped
- connector Tick stuck
- mutex deadlock
- CPU starvation
- unexpected action parsing error
- downstream ROS path unhealthy

Evidence:
- logs
- goroutine profile
- CPU/memory
- context/lifecycle events
- topic rates
- middleware errors
- action-call traces
- race detector in reproducible test environment

This is stronger than saying:
> "I'd restart it."

---

# 28. How your Bosch deployment story maps to FDE

Your deployment experience can be described using the same engineering pattern:

~~~text
model worked offline
       ↓
deployment behaved differently
       ↓
inspect runtime evidence
       ↓
discover input timing/sampling mismatch
       ↓
identify unnecessary CPU load
       ↓
remove/disable extra work
       ↓
restore expected sampling
       ↓
model performance recovered
       ↓
keep observability in place
~~~

The important thing is not that the code was Go.

The important thing is that you already use the FDE debugging method.

Go is the implementation language you are adding.

---

# 29. How to answer "How much Go do you know?"

Use accurate confidence.

A good version:

> "Python is still my strongest language. I've been ramping on Go directly through OM1 rather than just doing syntax exercises. I'm comfortable reading structs, receiver methods, interfaces, explicit error handling, maps and slices, and I've been tracing the runtime's goroutines, channels, contexts, mutexes, and action connectors. I wouldn't claim years of production Go yet, but I can read the codebase, reason about it, and make targeted changes, and I'm actively closing the implementation gap."

That is credible.

Do not undersell yourself with:

> "I don't really know Go."

And do not overclaim:

> "I'm an expert."

---

# 30. How to answer "Why Go?"

Do not say merely:
> faster than Python.

Better:

> "For this kind of runtime, Go gives a useful combination of compiled deployment, explicit typing, relatively low operational overhead, and very approachable concurrency primitives. OM1 has multiple long-lived inputs, action loops, context cancellation, and middleware connections, so goroutines, channels, and contexts fit the workload well."

You can mention:
- easy single-binary deployment
- concurrency model
- memory footprint
- performance

without claiming every Go program is automatically faster/better.

---

# 31. How to answer "What is an interface in Go?"

Strong:

> "It's a behavioral contract defined by method signatures. A type satisfies it implicitly by implementing those methods. OM1 uses that to keep the runtime independent from concrete sensors, LLMs, action connectors, and Zenoh implementations."

---

# 32. How to answer "struct versus class?"

Strong:

> "A struct primarily defines fields. You attach behavior with receiver methods, and interfaces define contracts independently. So a struct plus methods can serve many of the roles I'd use a Python class for, but Go favors composition and implicit interfaces rather than inheritance."

---

# 33. How to answer "what does this star mean?"

If in:

~~~go
*Runtime
~~~

say:

> pointer to Runtime.

If in:

~~~go
*configName
~~~

say:

> dereference the pointer to obtain the value.

Context matters.

---

# 34. How to answer "what is a goroutine?"

Strong:

> "A lightweight concurrent function managed by the Go runtime. You start one with the go keyword. OM1 uses them for independent sensor loops, connector loops, backgrounds, and the cortex lifecycle."

---

# 35. How to answer "what is a channel?"

Strong:

> "A typed communication/synchronization pipe between goroutines. OM1 uses channels for things like wake-up signals, completion, mode transitions, and event streams."

---

# 36. How to answer "what is context?"

Strong:

> "It propagates cancellation and deadlines through a call/lifecycle tree. In OM1, canceling the mode context lets the cortex, input, action, and background goroutines shut down cooperatively."

---

# 37. How to answer "why mutex?"

Strong:

> "To protect shared mutable state from concurrent access. The important thing is not just locking a variable, but protecting the invariant across related reads and writes."

---

# 38. Coding exercise strategy

If asked to code in Go:

1. restate input/output
2. write simplest correct solution
3. prefer clear standard-library constructs
4. handle errors explicitly
5. do not add concurrency unless problem requires it
6. test edge cases aloud
7. discuss complexity
8. then improve

Do not try to impress with goroutines unnecessarily.

Concurrency adds failure modes.

---

# 39. Example coding prompt: find first unhealthy sensor

Input:

~~~go
type SensorStatus struct {
    Name string
    Rate float64
    MinRate float64
}
~~~

Task:
> return the first sensor whose rate is below minimum.

Simple answer:

~~~go
func FirstUnhealthy(statuses []SensorStatus) (SensorStatus, bool) {
    for _, s := range statuses {
        if s.Rate < s.MinRate {
            return s, true
        }
    }
    return SensorStatus{}, false
}
~~~

What concepts?
- slice
- range
- struct
- multiple return
- zero value
- bool success flag

This is better interview code than overengineering.

---

# 40. Example coding prompt: lookup action

~~~go
func FindAction(actions map[string]*AgentAction, name string) (*AgentAction, error) {
    action, ok := actions[name]
    if !ok {
        return nil, fmt.Errorf("unknown action %q", name)
    }
    return action, nil
}
~~~

This mirrors real OM1 patterns.

---

# 41. Example coding prompt: cancellation-aware worker

~~~go
func RunWorker(ctx context.Context, ticks <-chan time.Time) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticks:
            doWork()
        }
    }
}
~~~

Explain:
- infinite event loop
- exits on cancellation
- otherwise handles ticks

This is a common Go service pattern.

---

# 42. What to memorize versus understand

Memorize lightly:
- func
- struct
- interface
- :=
- if err != nil
- []T
- map[K]V
- go
- chan
- select
- context
- defer
- mutex
- WaitGroup

Understand deeply:
- data flow
- ownership
- lifecycle
- failure boundaries
- interfaces
- concurrency intent
- safe shutdown
- observability

The second group matters more.

---

# 43. Five interview traps to avoid

## Trap 1
Calling structs "classes" repeatedly.

Use Go terminology.

## Trap 2
Saying interfaces require explicit implementation declarations.

They usually do not.

## Trap 3
Saying goroutines are simply OS threads.

They are scheduled by the Go runtime and are much lighter-weight abstractions.

## Trap 4
Saying channels are just queues.

They are communication/synchronization primitives with blocking semantics; buffered channels can act queue-like.

## Trap 5
Saying every OM1 robot command "goes through MCP."

Action connectors are the normal action execution abstraction; MCP is a separate optional tool integration mechanism.

---

# 44. OM1-specific questions you should now answer

Try aloud.

### Architecture

What is the Fuser?

What is Cortex?

What is an AgentAction?

What is an action Connector?

What does the action Orchestrator do?

What is MCP doing?

Where does Zenoh appear?

What is a mode?

### Go

Why are Sensor and Connector interfaces?

Why are orchestrators pointers?

Why are there mutexes?

Why are there goroutines?

What is TickNow?

Why is tickNow buffered by one?

Why do methods receive context?

What does ParseCalls do?

---

# 45. Suggested 15-minute pre-interview refresh

Do not reread five hours of notes.

Read only:

1. study/go-course/07_GO_OM1_CHEAT_SHEET.md
2. internal/llm/llm.go
3. internal/actions/action.go
4. Runtime.tick in internal/runtime/runtime.go
5. first half of the Go2 move connector

Then say aloud:

> "input → Fuser → Cortex → ToolCall → Action Orchestrator → Connector → middleware."

Then say:

> "goroutine → channel → select → context → mutex."

That is enough to reactivate the map.

---

# 46. Final practical exercise

Pretend you are given this bug:

> "The model logs the correct move-forwards ToolCall, but the robot does not move."

Do not read the answer yet.

Write your own evidence plan.

---

# 47. One strong answer

1. Confirm ParseCalls maps the emitted tool name to the intended AgentAction.
2. Confirm Connector.Connect is entered with expected arguments.
3. Inspect connector rejection logs: control disabled, already moving, no localization, barrier, pending movement.
4. Confirm a moveCommand becomes pending.
5. Confirm Tick is still running and context is not canceled.
6. Confirm odometry/path state is sane.
7. Confirm velocity serialization produces expected values.
8. Confirm Zenoh publisher exists and Put succeeds.
9. Observe downstream cmd_vel at the ROS/middleware side.
10. If cmd_vel is correct there, continue downstream to controller/HAL/safety/hardware rather than changing Cortex.

That is an FDE answer.

---

# 48. What five hours should and should not accomplish

After this course, you should not expect to:

- know every standard-library package
- write complex concurrent systems from memory
- understand every subtlety of the Go memory model
- optimize the compiler/runtime
- pass a senior Go language-trivia interview

You should expect to:

- read normal Go syntax
- recognize important Go idioms
- understand OM1's plugin architecture
- trace its runtime
- understand its concurrency model at a useful level
- follow structured tool calls into connectors
- make small targeted changes
- ask intelligent questions when code is unfamiliar
- debug boundaries systematically

That is a realistic and valuable outcome.

---

# 49. Final verbal test

Without notes, say:

> "Go organizes state with structs, behavior with receiver methods, and contracts with interfaces. OM1 uses config-driven factories to build concrete sensor, model, and action implementations behind those interfaces. Inputs run concurrently, expose formatted sensor state, and can wake the Cortex loop. The Fuser builds the prompt; Cortex returns a typed response with structured ToolCalls; the action orchestrator resolves those calls to AgentActions and schedules them; and concrete connectors translate the decisions into robot or service behavior. Contexts manage cancellation, channels signal events, and mutexes or atomics protect shared state."

If you understand every clause, the course did its job.

Next:

> 07_GO_OM1_CHEAT_SHEET.md
