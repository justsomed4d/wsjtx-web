package wsjtx

import (
	"encoding/json"
	"testing"
)

func TestTypeName(t *testing.T) {
	cases := map[MessageType]string{
		MsgHeartbeat: TypeHeartbeat,
		MsgStatus:    TypeStatus,
		MsgDecode:    TypeDecode,
		MsgClear:     TypeClear,
	}
	for mt, want := range cases {
		if got := TypeName(mt); got != want {
			t.Fatalf("TypeName(%d)=%q want %q", mt, got, want)
		}
	}
}

func TestMarshalEnvelopeDecode(t *testing.T) {
	msg := &DecodeMessage{
		BaseMessage:    BaseMessage{Id: "WSJT-X"},
		Time:           12345,
		SNR:            -12,
		DeltaTime:      0.2,
		DeltaFrequency: 1500,
		Mode:           "~",
		Message:        "CQ K1ABC FN42",
	}
	raw, err := MarshalEnvelope(TypeName(msg.Type()), msg)
	if err != nil {
		t.Fatal(err)
	}
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.Type != TypeDecode {
		t.Fatalf("type=%q want %q", env.Type, TypeDecode)
	}
	data, _ := json.Marshal(env.Data)
	var decoded DecodeMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.DeltaFrequency != 1500 || decoded.Message != "CQ K1ABC FN42" {
		t.Fatalf("unexpected decode payload: %+v", decoded)
	}
}

func TestReplyCommandJSON(t *testing.T) {
	raw := []byte(`{"Time":1,"SNR":-1,"DeltaTime":0.1,"DeltaFrequency":700,"Mode":"~","Message":"A B","LowConfidence":false,"Modifiers":0}`)
	var cmd ReplyCommand
	if err := json.Unmarshal(raw, &cmd); err != nil {
		t.Fatal(err)
	}
	if cmd.DeltaFrequency != 700 {
		t.Fatalf("DeltaFrequency=%d", cmd.DeltaFrequency)
	}
}
