# OM1 Integration Debug Lab

Practice fault isolation in a Linux and robotics style environment.

## Rules

Treat each exercise like an unfamiliar integration failure. Observe the system before editing code. State a hypothesis, test one boundary, make the smallest fix, and retest end to end.

## Python ROS 2 exercise

A simulated wheel speed publisher and monitor should communicate over ROS 2. The current version is intentionally misconfigured.

Run the two Python files in separate terminals. Your target behavior is for the monitor to print changing wheel speed values.

Use ROS graph and topic inspection, process inspection, and logs to determine where data flow stops.

## Go integration exercise

A tiny UDP sensor and consumer should exchange telemetry twice per second. The current version is intentionally misconfigured.

Run it with:

```bash
cd training/debug-lab/go_integration
go run .
```

Use application logs, Linux socket/process inspection, and VS Code/Delve as needed.

## Debrief

For each fault, be able to explain:
1. Symptom
2. Evidence
3. Failure boundary
4. Root cause
5. Minimal fix
6. How you would detect or prevent it in production
