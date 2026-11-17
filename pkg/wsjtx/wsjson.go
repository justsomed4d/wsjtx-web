package wsjtx

import (
	"encoding/json"
	"fmt"
)

// WebSocket JSON type names (server → client and client → server).
// Keep in sync with wsjtx-web-ui/src/types/ws.ts and docs/protocol.md.
const (
	TypeHeartbeat    = "heartbeat"
	TypeStatus       = "status"
	TypeDecode       = "decode"
	TypeClear        = "clear"
	TypeQSOLogged    = "qsoLogged"
	TypeClose        = "close"
	TypeWSPRDecode   = "wsprDecode"
	TypeLoggedADIF   = "loggedADIF"
	TypeAudioDevices = "audioDevices"

	CmdReply           = "reply"
	CmdHalt            = "halt"
	CmdGetAudioDevices = "getAudioDevices"
	CmdSetAudioDevice  = "setAudioDevice"
)

// Envelope is the JSON wrapper for text WebSocket messages.
type Envelope struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// TypeName maps a WSJT-X UDP message type to the WebSocket JSON type string.
func TypeName(t MessageType) string {
	switch t {
	case MsgHeartbeat:
		return TypeHeartbeat
	case MsgStatus:
		return TypeStatus
	case MsgDecode:
		return TypeDecode
	case MsgClear:
		return TypeClear
	case MsgQSOLogged:
		return TypeQSOLogged
	case MsgClose:
		return TypeClose
	case MsgWSPRDecode:
		return TypeWSPRDecode
	case MsgLoggedADIF:
		return TypeLoggedADIF
	default:
		return fmt.Sprintf("msg_%d", t)
	}
}

// MarshalEnvelope wraps a payload as {type, data} JSON.
func MarshalEnvelope(typeName string, data any) ([]byte, error) {
	return json.Marshal(Envelope{Type: typeName, Data: data})
}

// ReplyCommand is the client → server JSON body for CmdReply.
// Field names match DecodeMessage JSON so the UI can forward a decode.
type ReplyCommand struct {
	Time            uint32  `json:"Time"`
	SNR             int32   `json:"SNR"`
	DeltaTime       float64 `json:"DeltaTime"`
	DeltaFrequency  uint32  `json:"DeltaFrequency"`
	Mode            string  `json:"Mode"`
	Message         string  `json:"Message"`
	LowConfidence   bool    `json:"LowConfidence"`
	Modifiers       uint8   `json:"Modifiers"`
}

// HaltCommand is the client → server JSON body for CmdHalt.
type HaltCommand struct {
	AutoTxOnly bool `json:"AutoTxOnly"`
}

// SetAudioDeviceCommand is the client → server JSON body for CmdSetAudioDevice.
type SetAudioDeviceCommand struct {
	ID string `json:"id"`
}

// AudioDevice is one entry in a TypeAudioDevices payload.
type AudioDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
