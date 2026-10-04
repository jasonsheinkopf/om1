package inputs

// ============================================================================
// STUDY GUIDE — INPUT ORCHESTRATOR
//
// PURPOSE:
//   Run configured Sensor implementations concurrently and expose their latest
//   formatted observations to the Runtime.
//
// MENTAL MODEL:
//   microphone / camera / localization / robot state / other input
//       -> concrete Sensor plugin
//       -> processing (RawToText / plugin-specific work)
//       -> latest formatted buffer
//       -> Orchestrator.Buffers()
//       -> Runtime.tick -> Fuser
//
// KEY CORRECTION:
//   This layer does NOT mean "all ROS2 data automatically enters OM1."
//   Only configured input plugins participate. Sensor-specific processing occurs
//   before the Fuser; the Fuser should not receive raw LiDAR packets.
//
// GO TO NOTICE:
//   goroutines = concurrent sensor listeners
//   channels = readings, completion, and optional TickNow wake-up
//   context = cancellation/lifecycle propagation
// ============================================================================

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Orchestrator manages multiple Sensors.
type Orchestrator struct {
	sensors []Sensor
	tickNow chan struct{}
	log     *zap.Logger
}

// NewOrchestrator creates a new Orchestrator with the given sensors and logger.
// NewOrchestrator constructs the input manager around already-created Sensor implementations.
func NewOrchestrator(sensors []Sensor, log *zap.Logger) *Orchestrator {
	return &Orchestrator{
		sensors: sensors,
		tickNow: make(chan struct{}, 1),
		log:     log,
	}
}

// TickNow returns a channel that is sent a signal whenever any sensor receives new input.
// TickNow exposes the event channel that can wake Cortex immediately instead of waiting for the timer.
func (o *Orchestrator) TickNow() <-chan struct{} {
	return o.tickNow
}

// Start launches one goroutine per sensor and returns a channel that is closed when all goroutines have finished.
// Start launches the long-lived input work. Construction and execution are separate phases.
func (o *Orchestrator) Start(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	var wg sync.WaitGroup

	for i, sensor := range o.sensors {
		wg.Add(1)
		go func(sensorIndex int, sensor Sensor) {
			defer wg.Done()
			o.runSensor(ctx, sensorIndex, sensor)
		}(i, sensor)
	}

	go func() {
		wg.Wait()
		close(done)
	}()
	return done
}

// runSensor listens for readings from a sensor, converts them to text, and signals the orchestrator to tick when new input is received.
func (o *Orchestrator) runSensor(ctx context.Context, sensorIndex int, sensor Sensor) {
	readings, err := sensor.Listen(ctx)
	if err != nil {
		o.log.Warn("sensor listen failed", zap.Int("sensorIndex", sensorIndex), zap.Error(err))
		return
	}
	for {
		select {
		case rawReading, ok := <-readings:
			if !ok {
				return
			}

			message, err := sensor.RawToText(ctx, rawReading)
			if err != nil || message == nil {
				continue
			}

			if !triggersTick(sensor) {
				continue
			}

			select {
			case o.tickNow <- struct{}{}:
			default:
			}
		case <-ctx.Done():
			sensor.Stop()
			return
		}
	}
}

// Buffers returns a snapshot of the latest buffer from each sensor, formatted as text.
// Buffers returns the latest formatted observation from each sensor for the next Cortex tick.
func (o *Orchestrator) Buffers() []string {
	snapshot := make([]string, len(o.sensors))

	for i, sensor := range o.sensors {
		snapshot[i] = sensor.FormattedLatestBuffer()
	}

	return snapshot
}
