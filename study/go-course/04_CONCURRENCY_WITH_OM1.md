# Module 4 — Go Concurrency Through OM1

This is the most Go-specific module in the course.

If Python is your strongest language, syntax such as this may initially feel alien:

~~~go
go func() {
    defer close(done)
    rt.runCortexLoop(modeCtx)
}()
~~~

or:

~~~go
select {
case <-ctx.Done():
    return
case <-timer.C:
case <-current.inputOrchestrator.TickNow():
}
~~~

The goal of this module is not merely to define goroutines and channels.

The goal is to make you able to look at concurrent OM1 code and answer:

- what is running independently?
- what data or signal moves between tasks?
- what wakes a loop?
- what stops it?
- what state is shared?
- what protects that state?
- where could a race or deadlock happen?
- how does OM1 shut everything down?

---

# 1. Why OM1 naturally needs concurrency

A robot/agent runtime has many things happening at once:

- microphone input
- camera/VLM processing
- localization updates
- robot state
- timers
- LLM ticks
- action execution
- movement heartbeats
- mode changes
- background behavior
- logging/metrics
- middleware callbacks

A purely sequential architecture would be awkward:

~~~text
listen microphone
then wait
then poll camera
then wait
then update robot
then wait
then maybe call model
...
~~~

Instead, independent activities can run concurrently.

Go gives very lightweight language/runtime primitives for this.

---

# 2. Concurrency versus parallelism

These terms are related but not identical.

**Concurrency** means:

> multiple tasks can make progress independently and their lifetimes overlap.

**Parallelism** means:

> multiple tasks are literally executing at the same instant on different CPU cores.

Goroutines give you concurrency.

The Go runtime may schedule them in parallel when resources allow.

A useful interview sentence:

> "Goroutines are lightweight concurrent tasks scheduled by the Go runtime. Concurrency is the program structure; parallel execution may occur underneath."

---

# 3. The simplest goroutine

Normal call:

~~~go
doWork()
~~~

The current goroutine waits until doWork returns.

Concurrent call:

~~~go
go doWork()
~~~

The go keyword says:

> start this function as a new goroutine and continue here.

Python mental model is somewhere between:
- creating an asyncio task
- launching a very lightweight thread

It is not exactly either.

---

# 4. Anonymous goroutine

You will often see:

~~~go
go func() {
    doSomething()
}()
~~~

Break it apart:

~~~go
func() {
    doSomething()
}
~~~

is an anonymous function.

The final:

~~~text
()
~~~

calls it.

Adding go:

~~~go
go func() {
    doSomething()
}()
~~~

starts that function concurrently.

---

# 5. Real OM1 example: one goroutine per sensor

Open:

> internal/inputs/orchestrator.go

The Start method conceptually does:

~~~go
for i, sensor := range o.sensors {
    wg.Add(1)

    go func(sensorIndex int, sensor Sensor) {
        defer wg.Done()
        o.runSensor(ctx, sensorIndex, sensor)
    }(i, sensor)
}
~~~

Architecture:

~~~text
sensor 0 ── goroutine ── listens continuously
sensor 1 ── goroutine ── listens continuously
sensor 2 ── goroutine ── listens continuously
...
~~~

Why?

A slow microphone read should not prevent localization updates.

A camera should not block another sensor's event stream.

---

# 6. Why pass i and sensor into the anonymous function?

This code:

~~~go
go func(sensorIndex int, sensor Sensor) {
    ...
}(i, sensor)
~~~

passes current loop values as explicit arguments.

That makes the goroutine's inputs clear.

Historically, capturing loop variables incorrectly was a famous Go bug pattern. Modern Go versions improved loop-variable semantics, but explicit arguments remain easy to read and make intent obvious.

The important reading rule:

> The values after the anonymous function body are arguments to that function.

---

# 7. WaitGroup: "wait until all workers are finished"

A sync.WaitGroup is a counter used to coordinate completion.

Typical pattern:

~~~go
var wg sync.WaitGroup

wg.Add(1)

go func() {
    defer wg.Done()
    doWork()
}()

wg.Wait()
~~~

Mental model:

~~~text
Add(1)  -> one worker is outstanding
Done()  -> one worker finished
Wait()  -> block until count reaches zero
~~~

Python analogy:
- gathering tasks
- joining threads

---

# 8. Why defer wg.Done?

Instead of:

~~~go
doWork()
wg.Done()
~~~

code often uses:

~~~go
defer wg.Done()
doWork()
~~~

Meaning:

> whenever this function exits, mark this worker finished.

That protects cleanup even if there are multiple return paths.

This is one of the best uses of defer.

---

# 9. A completion channel

After launching sensor goroutines, OM1 also starts something conceptually like:

~~~go
go func() {
    wg.Wait()
    close(done)
}()
~~~

Meaning:

1. another goroutine waits for all sensor workers
2. when they finish, close the done channel

The caller receives done and can wait for the whole orchestrator to shut down.

This is a common pattern:

~~~text
many workers
   ↓
WaitGroup
   ↓
close(done)
   ↓
outside lifecycle manager knows subsystem ended
~~~

---

# 10. Channels: typed communication pipes

A channel has an element type.

Example:

~~~go
chan string
~~~

means:

> channel carrying strings.

Create one:

~~~go
ch := make(chan string)
~~~

Send:

~~~go
ch <- "hello"
~~~

Receive:

~~~go
msg := <-ch
~~~

Read the arrow as flow direction.

---

# 11. Signal-only channels

Often you do not need to send a payload.

You only need:

> something happened.

Go commonly uses:

~~~go
chan struct{}
~~~

Why struct{}?

An empty struct carries essentially no data.

So:

~~~go
tickNow chan struct{}
~~~

means:

> a channel used only to signal that a tick should happen.

Send:

~~~go
tickNow <- struct{}{}
~~~

Receive:

~~~go
<-tickNow
~~~

The receiver does not care about a value.

It cares that the event occurred.

---

# 12. Directional channel types

You may see:

~~~go
<-chan struct{}
~~~

This means:

> receive-only channel of empty structs.

And:

~~~go
chan<- string
~~~

would mean:

> send-only string channel.

Why use directional types?

They document and enforce intended use.

If a method returns:

~~~go
func (o *Orchestrator) TickNow() <-chan struct{}
~~~

the caller may receive the signal but may not send arbitrary events into the orchestrator's private channel.

That is encapsulation at the type level.

---

# 13. Buffered versus unbuffered channels

Unbuffered:

~~~go
make(chan string)
~~~

A send generally waits until a receiver is ready.

Buffered:

~~~go
make(chan string, 1)
~~~

The channel can hold one unread item.

Mental model:

~~~text
unbuffered: handoff
buffered: tiny queue
~~~

This distinction matters in OM1.

---

# 14. Real OM1 example: tickNow buffer size 1

Input orchestrator creates:

~~~go
tickNow: make(chan struct{}, 1)
~~~

Why buffer one signal?

Suppose several sensor events arrive before the cortex loop handles the first wake-up.

OM1 may not need five separate queued wakeups.

It mainly needs:

> there is fresh data; please tick.

A one-slot channel naturally coalesces repeated wake-up intent.

---

# 15. Non-blocking send with select/default

OM1 uses a pattern conceptually like:

~~~go
select {
case o.tickNow <- struct{}{}:
default:
}
~~~

Meaning:

> Try to send a wake-up signal. If the channel cannot accept it immediately, do nothing instead of blocking.

Because the channel only has room for one pending signal, if one is already queued, another signal can be dropped safely.

This is deliberate lossy signaling.

The system is not dropping sensor state itself.

It is dropping redundant:

> "please wake up"

notifications.

That distinction matters.

---

# 16. select: wait for whichever event happens

A select statement operates on channel operations.

Example:

~~~go
select {
case msg := <-messages:
    handle(msg)

case <-ctx.Done():
    return
}
~~~

Read:

> wait until either a message arrives or cancellation happens.

It is analogous to a multi-way event wait.

Python asyncio mental model:

> await whichever event/task becomes ready first.

---

# 17. select with a timer in the Cortex loop

OM1's cortex loop conceptually waits on:

~~~text
1. shutdown
2. regular timer
3. immediate sensor wake-up
4. mode-context update
~~~

That is exactly the kind of logic select is good for.

Instead of constantly polling:

~~~text
did shutdown happen?
did timer happen?
did sensor happen?
did mode update happen?
~~~

the goroutine sleeps until one case is ready.

---

# 18. Timers and ticker-like behavior

OM1 creates a timer for the next cortex cycle.

Conceptually:

~~~go
timer := time.NewTimer(tickInterval)
~~~

Then:

~~~go
case <-timer.C:
~~~

means:

> the timer fired.

timer.C is a channel.

This illustrates a broader Go design style:

> asynchronous events are frequently exposed as channels.

---

# 19. Why stop the timer?

If another event wins the select before the timer fires, OM1 stops the timer.

That prevents unnecessary timer activity/resources.

This is lifecycle hygiene.

In concurrent systems, cleanup matters because tiny leaks can accumulate in long-running robot processes.

---

# 20. context.Context: cooperative cancellation

This is probably the most important lifecycle concept after goroutines.

A context can carry:
- cancellation
- deadline
- timeout
- request-scoped values, though overuse is discouraged

For OM1, the key role is cancellation.

A function accepts:

~~~go
ctx context.Context
~~~

Then checks:

~~~go
ctx.Err()
~~~

or waits:

~~~go
<-ctx.Done()
~~~

Meaning:

> has the parent asked me to stop?

---

# 21. Cancellation is cooperative

Calling cancel does not forcibly kill a goroutine.

Instead:

~~~go
cancel()
~~~

closes/signals the context's Done channel.

Well-behaved code notices and returns.

So think:

~~~text
parent: "please stop"
child:  checks Done, cleans up, returns
~~~

This is safer than abruptly terminating threads.

---

# 22. Context trees

You can derive a child context:

~~~go
modeCtx, cancel := context.WithCancel(ctx)
~~~

Now modeCtx is connected to its parent.

If parent ctx is canceled:
- modeCtx is canceled too

If cancel() is called:
- modeCtx is canceled
- parent remains alive

This creates a lifecycle tree.

OM1 uses that for runtime modes.

A mode can stop all of its own orchestrators without shutting down the entire process.

---

# 23. Runtime mode lifecycle

Conceptually:

~~~text
process context
      ↓
mode context
   ├─ input orchestrator
   ├─ action orchestrator
   ├─ background orchestrator
   └─ cortex loop
~~~

On mode transition:

~~~text
cancel mode context
      ↓
children stop
      ↓
wait for done channels
      ↓
construct next mode
      ↓
start new mode context
~~~

This is much cleaner than global stop flags scattered everywhere.

---

# 24. Context timeout

You may see:

~~~go
ctx, cancel := context.WithTimeout(parent, 15*time.Second)
defer cancel()
~~~

Meaning:

> create a child context that automatically cancels after 15 seconds.

Useful for:
- network calls
- shutdown waits
- external services
- operations that must not hang forever

For an FDE, timeouts are extremely important.

A field system should degrade and report rather than wait forever on one broken dependency.

---

# 25. Mutex: protect shared mutable state

A mutex is a lock.

Pattern:

~~~go
mu.Lock()
sharedValue = x
mu.Unlock()
~~~

Only one goroutine can hold the mutex at a time.

Python analogy:

~~~python
with lock:
    shared_value = x
~~~

Why?

Because concurrent reads/writes can cause:
- data races
- inconsistent state
- corrupted invariants

---

# 26. Real OM1 example: Runtime current state

Runtime contains a mutex and shared fields.

Code conceptually does:

~~~go
rt.mu.Lock()
current := rt.current
rt.mu.Unlock()
~~~

This creates a snapshot of the pointer while protected by the lock.

The lock is held for a very short time.

That is generally good practice:

> protect the shared state, copy what you need, release the lock, then do slower work outside the critical section.

---

# 27. defer with mutex unlock

Another pattern:

~~~go
mu.Lock()
defer mu.Unlock()

// protected work
~~~

This guarantees unlock on all returns.

But holding a mutex across slow work can be dangerous.

Always ask:

> How long is the critical section?

Good concurrent design keeps lock scope deliberate.

---

# 28. Real OM1 example: pending robot movement

The move connector stores:

~~~text
mu
pending *moveCommand
~~~

Connect may check or queue a pending command.

Tick reads and updates it.

Those paths may execute from different goroutines.

The mutex protects the invariant:

> there is one coherent pending movement state.

This is a classic reason for a lock.

---

# 29. Race condition

A data race occurs when concurrent goroutines access the same memory and at least one access is a write without appropriate synchronization.

Example:

~~~go
count := 0

go func() { count++ }()
go func() { count++ }()
~~~

This is not safe just because increment looks simple.

Read-modify-write is multiple operations.

Use:
- mutex
- atomic
- channel ownership
- other synchronization

depending on the problem.

---

# 30. Atomic values

OM1's movement connector uses an atomic boolean for AI control enablement.

Atomic operations are designed for small values that need thread-safe reads/writes without a full mutex.

Conceptually:

~~~go
flag.Store(true)
enabled := flag.Load()
~~~

Use atomics for simple independent state.

Use mutexes when several fields form one invariant or operations need to be grouped.

---

# 31. Mutex versus atomic

Rule of thumb:

Use atomic when:
- one simple scalar state
- independent Load/Store/CAS semantics are enough

Use mutex when:
- multiple fields must change together
- you need a compound check-and-update
- protected logic is more complex

OM1 uses both for different reasons.

---

# 32. WaitGroup versus channel

These solve different coordination problems.

WaitGroup:

> wait until N workers finish.

Channel:

> communicate values/events between goroutines.

A common pattern uses both:
- channel for events
- WaitGroup for lifecycle completion

Do not treat channels as a replacement for every synchronization mechanism.

---

# 33. Closing a channel

Closing says:

> no more values will ever be sent.

Receiver can detect closure.

For a completion channel, closing is perfect because every waiter can observe:

> subsystem is done.

Important rule:

> Normally the sender/owner closes the channel, not arbitrary receivers.

And:

> sending on a closed channel panics.

---

# 34. Receiving from a closed channel

Receive can use:

~~~go
value, ok := <-ch
~~~

If ok is false:
> channel is closed and drained.

OM1's sensor loop uses this sort of pattern for sensor readings.

That lets it stop when the producer ends.

---

# 35. default in select

A select with default becomes non-blocking.

Example:

~~~go
select {
case value := <-ch:
    use(value)
default:
    // nothing ready
}
~~~

Without default:
> wait until a case is ready.

With default:
> if no channel operation is immediately ready, execute default now.

This is powerful but potentially dangerous.

---

# 36. Busy loops

Consider:

~~~go
for {
    select {
    case <-ctx.Done():
        return
    default:
        doSomething()
    }
}
~~~

If doSomething returns immediately, this loop can consume CPU continuously.

Therefore, when reading code like this, ask:

> Does the work block, wait, sleep, or throttle itself?

In OM1's action connector loop, Connector.Tick implementations are responsible for appropriate behavior. The Go2 move Tick waits on a short timer before proceeding.

This is an important field-debugging clue:

> unexpected high CPU may come from a loop that no longer blocks as expected.

---

# 37. Concurrent action execution

The action orchestrator supports different execution modes.

In concurrent mode, conceptually:

~~~go
for each call:
    wg.Add(1)
    go execute(call)
wg.Wait()
~~~

This means multiple independent actions can run at the same time.

Example:
- speak
- set facial emotion

may not need to block each other.

But concurrency is a product semantics decision.

Some actions should be sequential.

---

# 38. Sequential action execution

Sequential mode simply processes:

~~~text
call 1
then call 2
then call 3
~~~

This is useful when order matters.

For example:
- initialize
- then move
- then confirm

The key point:

> concurrency is not automatically "better." Correct ordering matters.

---

# 39. Dependency-aware execution

OM1 also has a dependencies mode.

Conceptually:

~~~text
child action depends on parent action
        ↓
wait until parent completes
        ↓
run child
~~~

This combines concurrency with ordering constraints.

The implementation tracks completed calls and protects shared completion state with a mutex.

Even if you never write this from memory, learn to identify:
- dependency graph
- waiting condition
- protected shared state
- cancellation escape path

---

# 40. Deadlock: what it means

A deadlock occurs when goroutines are permanently waiting on each other or on events that can no longer happen.

Classic example:

~~~text
goroutine A waits for B
goroutine B waits for A
~~~

or:
- sending on unbuffered channel with no receiver
- waiting on WaitGroup count that will never reach zero
- locking the same non-reentrant mutex incorrectly
- holding lock while calling code that waits for the same lock

Go may detect some global deadlock situations at runtime, but not every logical stall is automatically obvious.

---

# 41. Channel ownership is a powerful design

One alternative to shared memory is:

> one goroutine owns the state; other goroutines send it commands through channels.

Then fewer mutexes are needed.

This is related to the common Go saying:

> Do not communicate by sharing memory; share memory by communicating.

Real systems use both channel ownership and mutexes.

OM1 also uses both.

The slogan is a design heuristic, not a law.

---

# 42. A full input concurrency walkthrough

Follow one sensor.

## Step 1

Input orchestrator starts a goroutine for the sensor.

## Step 2

The goroutine calls Sensor.Listen.

## Step 3

Listen returns a channel of raw readings.

## Step 4

The goroutine waits in select for:
- a reading
- context cancellation

## Step 5

A reading arrives.

## Step 6

RawToText processes it.

## Step 7

If that sensor should wake Cortex, orchestrator attempts a non-blocking signal on tickNow.

## Step 8

Cortex loop's select receives TickNow.

## Step 9

Cortex immediately performs a tick instead of waiting for the regular timer.

That is a full event-driven chain.

---

# 43. A full shutdown walkthrough

Suppose SIGTERM reaches the process.

## Step 1

The root context is canceled.

## Step 2

Runtime Run wakes from:

~~~go
<-ctx.Done()
~~~

## Step 3

Runtime asks orchestrators to stop.

## Step 4

Mode cancellation propagates through modeCtx.

## Step 5

Sensor loops see:

~~~go
<-ctx.Done()
~~~

and return.

## Step 6

Action connector tick loops also see cancellation and return.

## Step 7

Deferred cleanup calls Stop and Done.

## Step 8

WaitGroups reach zero.

## Step 9

done channels close.

## Step 10

Runtime finishes shutdown.

That is structured concurrency/lifecycle management in practice.

---

# 44. Goroutine leak

A goroutine leak happens when a goroutine should have ended but remains blocked/running indefinitely.

Causes can include:
- forgotten cancellation
- blocked send with no receiver
- external call without timeout
- channel never closed
- infinite loop
- worker not connected to lifecycle context

In a long-lived robot service, leaks matter.

Symptoms:
- increasing goroutine count
- rising memory
- stale tasks
- duplicated behavior
- shutdown hangs

---

# 45. Context should travel with operations

Notice many interfaces accept:

~~~go
ctx context.Context
~~~

That lets cancellation propagate into:
- LLM calls
- sensor work
- connectors
- knowledge-base requests
- MCP work

A good rule:

> If work belongs to a request/lifecycle, pass the context through rather than replacing it with context.Background deep in the call chain.

There are exceptions, but this is a strong default.

---

# 46. The Cortex loop as a state machine

Think about:

> internal/runtime/runtime.go

not as mysterious concurrent syntax, but as an event loop.

Conceptually:

~~~text
while mode alive:
    wait for:
        timer
        sensor wake-up
        mode update
        cancellation

    if ordinary tick:
        collect sensors
        fuse prompt
        call model
        execute actions
~~~

Once you see it this way, the Go syntax becomes implementation detail around a familiar architecture.

---

# 47. Concurrency debugging: ask what should be blocked

If CPU is unexpectedly high:

Ask:
- which loops are spinning?
- which Tick method should have been blocking?
- is a default select causing continuous iteration?

If system is frozen:

Ask:
- which goroutine is waiting?
- on which channel?
- on which mutex?
- on which external operation?
- was its context canceled?
- is there a timeout?

If shutdown hangs:

Ask:
- which done channel never closed?
- which WaitGroup count never reached zero?
- which worker ignored ctx.Done?

This is directly useful for an FDE.

---

# 48. Concurrency debugging tools to know conceptually

You do not need mastery for the recruiter screen, but know these exist:

- Go race detector: go test -race ./...
- pprof goroutine profiles
- stack dumps
- structured logs with lifecycle IDs
- metrics for worker/goroutine counts and latency
- context deadlines
- tracing around external calls

The race detector is especially important.

It can detect many unsafe concurrent memory accesses during testing/runtime.

---

# 49. A tiny concurrency example

~~~go
package main

import (
    "context"
    "fmt"
    "time"
)

func sensor(ctx context.Context, out chan<- string) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            out <- "new reading"
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    readings := make(chan string)

    go sensor(ctx, readings)

    fmt.Println(<-readings)
    cancel()
}
~~~

Read it:

1. make root context
2. make channel
3. start sensor goroutine
4. sensor periodically sends readings
5. main receives one
6. main cancels
7. sensor notices cancellation and exits

That tiny program contains much of the lifecycle model used in larger systems.

---

# 50. Python comparison

A Python asyncio version might involve:
- asyncio.create_task
- asyncio.Queue
- asyncio.Event
- cancellation
- gather
- locks

The exact APIs differ, but the conceptual problems are familiar.

You are not learning why concurrency exists.

You are learning Go's vocabulary for it.

---

# 51. Read this OM1 pattern aloud

~~~go
select {
case <-ctx.Done():
    return
case <-timer.C:
case <-current.inputOrchestrator.TickNow():
}
~~~

Strong verbal explanation:

> "This goroutine blocks until one of three events occurs: cancellation, the periodic cortex timer, or an input-triggered immediate tick. Cancellation exits. Either timer or input wake-up proceeds to run the next cortex cycle."

That is interview-quality explanation.

---

# 52. Read this mutex pattern aloud

~~~go
rt.mu.Lock()
current := rt.current
rt.mu.Unlock()
~~~

Strong explanation:

> "Runtime current mode state is shared, so the code protects the read with a mutex, copies the pointer into a local variable, and releases the lock before doing more work."

---

# 53. Read this WaitGroup pattern aloud

~~~go
wg.Add(1)
go func() {
    defer wg.Done()
    work()
}()
wg.Wait()
~~~

Strong explanation:

> "The WaitGroup tracks one outstanding goroutine. Done is deferred so completion is reported on every return path, and Wait blocks until the worker finishes."

---

# 54. Read this non-blocking signal aloud

~~~go
select {
case tickNow <- struct{}{}:
default:
}
~~~

Strong explanation:

> "This tries to queue a wake-up signal without blocking. If one is already pending and the one-slot buffer is full, it drops the redundant signal."

That last word — redundant — is the key architectural insight.

---

# 55. Self-test

### Question 1
What is the difference between a goroutine and a normal function call?

### Question 2
Why does OM1 use a one-slot buffered tick channel?

### Question 3
What does select do?

### Question 4
What does context cancellation actually do to a goroutine?

### Question 5
What problem does a WaitGroup solve?

### Question 6
What problem does a Mutex solve?

### Question 7
Why might an atomic boolean be preferable to a mutex?

### Question 8
What is a goroutine leak?

### Question 9
Why can select with default create high CPU?

### Question 10
Why are cancellation and timeouts especially important in robotics/FDE systems?

---

# 56. Answers

### 1
A normal call runs synchronously in the current goroutine. Prefixing the call with go schedules it as a separate lightweight concurrent goroutine.

### 2
It needs to remember that a wake-up is pending, but repeated sensor events do not necessarily need to queue many identical wake-up requests.

### 3
It waits on multiple channel operations and executes a ready case. With default it can become non-blocking.

### 4
Cancellation signals through ctx.Done. The goroutine must cooperate by checking/waiting on the context and returning.

### 5
It counts outstanding workers so code can wait until all have completed.

### 6
It protects shared mutable state so multiple goroutines do not race while reading/writing an invariant.

### 7
For a single independent scalar flag, atomic Load/Store can provide simple thread-safe operations with less coordination machinery.

### 8
A goroutine that remains alive after its useful lifecycle should have ended.

### 9
If no channel is ready, default runs immediately; inside an infinite loop this can spin continuously unless some work blocks/throttles.

### 10
Field systems interact with unreliable networks, sensors, services, and hardware. Work must stop cleanly and must not hang indefinitely when a dependency fails.

---

# 57. Finish line

You are ready for Module 5 if you can explain:

> "OM1 uses goroutines for independent long-running activities, channels for events and completion, select for waiting on multiple events, contexts for cancellation, WaitGroups for joining workers, and mutexes/atomics for shared state."

More importantly, you should be able to point to where each appears in OM1.

Next:

> 05_OM1_END_TO_END_IN_GO.md
