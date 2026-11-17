/** WebSocket JSON contract — keep in sync with pkg/wsjtx/wsjson.go and docs/protocol.md. */

export const ServerMsgType = {
  heartbeat: 'heartbeat',
  status: 'status',
  decode: 'decode',
  clear: 'clear',
  qsoLogged: 'qsoLogged',
  close: 'close',
  wsprDecode: 'wsprDecode',
  loggedADIF: 'loggedADIF',
  audioDevices: 'audioDevices',
} as const

export type ServerMsgTypeName = (typeof ServerMsgType)[keyof typeof ServerMsgType]

export const ClientMsgType = {
  reply: 'reply',
  halt: 'halt',
  getAudioDevices: 'getAudioDevices',
  setAudioDevice: 'setAudioDevice',
} as const

export type ClientMsgTypeName = (typeof ClientMsgType)[keyof typeof ClientMsgType]

export interface DecodeMessage {
  Id?: string
  New?: boolean
  Time: number
  SNR: number
  DeltaTime: number
  DeltaFrequency: number
  Mode: string
  Message: string
  LowConfidence?: boolean
  OffAir?: boolean
}

export interface StatusMessage {
  Id?: string
  DialFrequency: number
  Mode: string
  DXCall: string
  Report: string
  TxMode?: string
  TxEnabled: boolean
  Transmitting: boolean
  Decoding?: boolean
  RxDF?: number
  TxDF?: number
  DECall?: string
  DEGrid?: string
  DXGrid?: string
  ConfigName?: string
  TxMessage?: string
}

export interface AudioDevice {
  id: string
  name: string
}

/** Client → server: forward a decode as a reply (field names match DecodeMessage). */
export interface ReplyCommand {
  Time: number
  SNR: number
  DeltaTime: number
  DeltaFrequency: number
  Mode: string
  Message: string
  LowConfidence: boolean
  Modifiers: number
}

export interface HaltCommand {
  AutoTxOnly: boolean
}

export interface SetAudioDeviceCommand {
  id: string
}

export type ServerEnvelope =
  | { type: typeof ServerMsgType.status; data: StatusMessage }
  | { type: typeof ServerMsgType.decode; data: DecodeMessage }
  | { type: typeof ServerMsgType.audioDevices; data: AudioDevice[] }
  | { type: typeof ServerMsgType.heartbeat; data: unknown }
  | { type: typeof ServerMsgType.clear; data: unknown }
  | { type: typeof ServerMsgType.qsoLogged; data: unknown }
  | { type: typeof ServerMsgType.close; data: unknown }
  | { type: typeof ServerMsgType.wsprDecode; data: unknown }
  | { type: typeof ServerMsgType.loggedADIF; data: unknown }

export type ClientEnvelope =
  | { type: typeof ClientMsgType.reply; data: ReplyCommand }
  | { type: typeof ClientMsgType.halt; data: HaltCommand }
  | { type: typeof ClientMsgType.getAudioDevices; data: Record<string, never> }
  | { type: typeof ClientMsgType.setAudioDevice; data: SetAudioDeviceCommand }
