# WebSocket protocol

JSON text frames wrap WSJT-X events and UI commands. Binary frames are raw Int16 PCM (when `-audio` is enabled).

Envelope shape (both directions):

```json
{ "type": "<name>", "data": { ... } }
```

Go constants live in `pkg/wsjtx/wsjson.go`. TypeScript mirrors them in `wsjtx-web-ui/src/types/ws.ts`.

## Server → client (`type`)

| `type` | Source | Notes |
|--------|--------|--------|
| `heartbeat` | WSJT-X UDP 0 | |
| `status` | WSJT-X UDP 1 | Dial/mode/TX flags |
| `decode` | WSJT-X UDP 2 | Band-activity rows |
| `clear` | WSJT-X UDP 3 | |
| `qsoLogged` | WSJT-X UDP 5 | |
| `close` | WSJT-X UDP 6 | |
| `wsprDecode` | WSJT-X UDP 10 | |
| `loggedADIF` | WSJT-X UDP 12 | |
| `audioDevices` | local | `{ id, name }[]` — empty when audio disabled |

Payload field names use the Go struct JSON tags (PascalCase for WSJT-X fields, e.g. `DeltaFrequency`, `DXCall`, `DialFrequency`).

## Client → server (`type`)

| `type` | `data` | Effect |
|--------|--------|--------|
| `reply` | see below | Encode UDP Reply to last WSJT-X peer |
| `halt` | `{ "AutoTxOnly": bool }` | Encode UDP HaltTx |
| `getAudioDevices` | `{}` | Respond with `audioDevices` |
| `setAudioDevice` | `{ "id": string }` | Start capture on device (no-op if `-audio` off) |

### `reply` fields

Must match a decode so the UI can forward a row:

| Field | Type |
|-------|------|
| `Time` | number (ms since midnight) |
| `SNR` | number |
| `DeltaTime` | number |
| `DeltaFrequency` | number |
| `Mode` | string |
| `Message` | string |
| `LowConfidence` | bool |
| `Modifiers` | number (uint8) |

Note: Go’s UDP `ReplyMessage` uses an internal field `DeltaFreq`; the **JSON** name is always `DeltaFrequency`.

## Binary frames

Server → client only: little-endian Int16 mono PCM @ 12 kHz when audio capture is running.
