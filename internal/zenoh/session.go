package zenoh

// ============================================================================
// STUDY GUIDE — ZENOH ABSTRACTION
//
// OM1 hides middleware details behind the Session, Publisher, and Subscriber
// interfaces. Higher-level code can publish/subscribe without depending directly
// on every concrete Zenoh implementation detail.
//
// IMPORTANT INTERVIEW NUANCE:
//   Do not say "Zenoh is always OM1's ROS bridge." OM1 uses pluggable connectors.
//   Some robot paths use Zenoh/ROS-facing topics; other connectors can use HTTP,
//   TTS services, SDKs, WebSockets, or other interfaces.
//
// GO2 EXAMPLE:
//   move connector -> Publisher.Put(serialized Twist bytes) -> cmd_vel
// ============================================================================

import (
	"sync"

	"github.com/openmind/om1/internal/cloudsession"
)

// Session is the middleware-facing contract for declaring pub/sub resources and one-shot puts.\ntype Session interface {
	// DeclarePublisher declares a publisher for the given key and returns it.
	DeclarePublisher(key string) (Publisher, error)

	// DeclareSubscriber declares a subscriber for the given key and handler function.
	DeclareSubscriber(key string, handler func([]byte)) (Subscriber, error)

	//	Put publishes data to the given key (one-shot).
	Put(key string, data []byte) error

	// Close properly closes the session and releases resources.
	Close()
}

// Publisher owns a declared key expression and accepts serialized []byte payloads.\ntype Publisher interface {
	// Put publishes data on the publisher's key expression.
	Put(data []byte) error

	// Drop undeclares the publisher and releases resources.
	Drop()
}

type Subscriber interface {
	// Drop undeclares the subscriber and releases resources.
	Drop()
}

// Options configures how a Session is opened.
type Options struct {
	// Endpoint is the local Zenoh router endpoint (e.g. tcp/127.0.0.1:7447).
	Endpoint string

	// LocalNetwork forces the session to connect to a local Zenoh router and not attempt discovery.
	LocalNetwork bool

	// UseSim enables the hybrid session that routes cloud topics to a cloud session and non-cloud topics to a local Zenoh session.
	UseSim bool

	// APIKey is the API key for authenticating with the cloud broker.
	APIKey string

	// CloudURL is the WebSocket URL of the cloud broker.
	CloudURL string
}

var (
	defaultOptsMu sync.RWMutex
	defaultOpts   Options
)

func SetDefaultOptions(opts Options) {
	if opts.CloudURL == "" {
		opts.CloudURL = cloudsession.DefaultURL
	}
	defaultOptsMu.Lock()
	defaultOpts = opts
	defaultOptsMu.Unlock()
}

func defaultOptions() Options {
	defaultOptsMu.RLock()
	defer defaultOptsMu.RUnlock()
	return defaultOpts
}

// Open creates a Session using process defaults. Callers depend on the Session interface,\n// which keeps concrete middleware details behind this boundary.
func Open(endpoint ...string) (Session, error) {
	opts := defaultOptions()
	opts.LocalNetwork = true
	if len(endpoint) > 0 {
		opts.Endpoint = endpoint[0]
	}

	return OpenWithOptions(opts)
}

// OpenWithOptions creates a Session with explicit options.
func OpenWithOptions(opts Options) (Session, error) {
	if opts.UseSim {
		return newHybridSession(opts), nil
	}

	return openZenoh(opts)
}
