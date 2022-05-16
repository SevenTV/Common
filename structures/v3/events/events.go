package events

import "time"

type Message[D any] struct {
	Op        Opcode `json:"op"`
	Timestamp int64  `json:"t"`
	Data      D      `json:"d"`
	Sequence  uint64 `json:"s,omitempty"`
}

func NewMessage[D any](op Opcode, data D) (Message[D], error) {
	msg := Message[D]{
		Op:        op,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}

	return msg, nil
}

type Opcode uint8

const (
	// Default ops (0-32)
	OpcodeDispatch      Opcode = 0 // R - Server dispatches data to the client
	OpcodeHello         Opcode = 1 // R - Server greets the client
	OpcodeHeartbeat     Opcode = 2 // R - Keep the connection alive
	OpcodeReconnect     Opcode = 4 // R - Server demands that the client reconnects
	OpcodeInboundSignal Opcode = 5 // R - A spectator signal is received

	// Commands (33-64)
	OpcodeIdentify  Opcode = 33 // S - Authenticate the session
	OpcodeResume    Opcode = 34 // S - Resume the previous session and receive missed events
	OpcodeSubscribe Opcode = 35 // S - Subscribe to one or multiple topics
	OpcodeSignal    Opcode = 36 // S - Emit a spectator signal
)

type CloseCode uint16

const (
	CloseCodeServerError       CloseCode = 4000 // an error occured on the server's end
	CloseCodeUnknownOperation  CloseCode = 4001 // the client sent an unexpected opcode
	CloseCodeInvalidPayload    CloseCode = 4002 // the client sent a payload that couldn't be decoded
	CloseCodeAuthFailure       CloseCode = 4003 // the client unsucessfully tried to identify
	CloseCodeAlreadyIdentified CloseCode = 4004 // the client wanted to identify again
	CloseCodeRateLimit         CloseCode = 4005 // the client is being rate-limited
	CloseCodeRestart           CloseCode = 4006 // the server is restarting and the client should reconnect
	CloseCodeMaintenance       CloseCode = 4007 // the server is in maintenance mode and not accepting connections
	CloseCodeTimeout           CloseCode = 4008 // the client was idle for too long
)

func (c CloseCode) String() string {
	switch c {
	case CloseCodeServerError:
		return "Internal Server Error"
	case CloseCodeUnknownOperation:
		return "Unknown Operation"
	case CloseCodeInvalidPayload:
		return "Invalid Payload"
	case CloseCodeAuthFailure:
		return "Authentication Failed"
	case CloseCodeAlreadyIdentified:
		return "Already identified"
	case CloseCodeRateLimit:
		return "Rate limit reached"
	case CloseCodeRestart:
		return "Server is restarting"
	case CloseCodeMaintenance:
		return "Maintenance Mode"
	case CloseCodeTimeout:
		return "Timeout"
	default:
		return "Undocumented Closure"
	}
}
