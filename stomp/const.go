package stomp

// STOMP protocol version.
var STOMP = []byte("1.2")

// STOMP protocol methods.
var (
	MethodStomp       = []byte("STOMP")
	MethodConnect     = []byte("CONNECT")
	MethodConnected   = []byte("CONNECTED")
	MethodSend        = []byte("SEND")
	MethodSubscribe   = []byte("SUBSCRIBE")
	MethodUnsubscribe = []byte("UNSUBSCRIBE")
	MethodAck         = []byte("ACK")
	MethodNack        = []byte("NACK")
	MethodDisconnect  = []byte("DISCONNECT")
	MethodMessage     = []byte("MESSAGE")
	MethodRecipet     = []byte("RECEIPT")
	MethodError       = []byte("ERROR")
)

// STOMP protocol headers.
var (
	HeaderAccept       = []byte("accept-version")
	HeaderAck          = []byte("ack")
	HeaderExpires      = []byte("expires")
	HeaderDest         = []byte("destination")
	HeaderHost         = []byte("host")
	HeaderLogin        = []byte("login")
	HeaderPass         = []byte("passcode")
	HeaderID           = []byte("id")
	HeaderMessageID    = []byte("message-id")
	HeaderPersist      = []byte("persist")
	HeaderPrefetch     = []byte("prefetch-count")
	HeaderReceipt      = []byte("receipt")
	HeaderReceiptID    = []byte("receipt-id")
	HeaderRetain       = []byte("retain")
	HeaderSelector     = []byte("selector")
	HeaderServer       = []byte("server")
	HeaderSession      = []byte("session")
	HeaderSubscription = []byte("subscription")
	HeaderVersion      = []byte("version")
)

// Common STOMP header values.
var (
	AckAuto      = []byte("auto")
	AckClient    = []byte("client")
	PersistTrue  = []byte("true")
	RetainTrue   = []byte("true")
	RetainLast   = []byte("last")
	RetainAll    = []byte("all")
	RetainRemove = []byte("remove")
)

var headerLookup = map[string]struct{}{
	"accept-version": struct{}{},
	"ack":            struct{}{},
	"expires":        struct{}{},
	"destination":    struct{}{},
	"host":           struct{}{},
	"login":          struct{}{},
	"passcode":       struct{}{},
	"id":             struct{}{},
	"message-id":     struct{}{},
	"persist":        struct{}{},
	"prefetch-count": struct{}{},
	"receipt":        struct{}{},
	"receipt-id":     struct{}{},
	"retain":         struct{}{},
	"selector":       struct{}{},
	"server":         struct{}{},
	"session":        struct{}{},
	"subscription":   struct{}{},
	"version":        struct{}{},
}
