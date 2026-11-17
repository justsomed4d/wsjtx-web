import { ref } from 'vue'
import { defineStore } from 'pinia'
import {
  ClientMsgType,
  ServerMsgType,
  type AudioDevice,
  type ClientEnvelope,
  type ClientMsgTypeName,
  type DecodeMessage,
  type StatusMessage,
} from '@/types/ws'

export const useWsjtxStore = defineStore('wsjtx', () => {
  const isConnected = ref(false)
  const messages = ref<DecodeMessage[]>([])
  const status = ref<StatusMessage | null>(null)

  let ws: WebSocket | null = null

  const audioQueue = ref<Float32Array[]>([])
  const audioContext = new (window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext)()

  const audioDevices = ref<AudioDevice[]>([])

  function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws = new WebSocket(`${protocol}//${window.location.host}/ws`)
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      isConnected.value = true
      console.log('WebSocket connected')
      fetchAudioDevices()
    }

    ws.onmessage = (event) => {
      if (event.data instanceof ArrayBuffer) {
        const int16Data = new Int16Array(event.data)
        const float32Data = new Float32Array(int16Data.length)
        for (let i = 0; i < int16Data.length; i++) {
          const sample = int16Data[i]
          if (sample !== undefined) {
            float32Data[i] = sample / 32768.0
          }
        }
        audioQueue.value.push(float32Data)
        if (audioQueue.value.length > 50) {
          audioQueue.value.shift()
        }
      } else {
        try {
          const data = JSON.parse(event.data as string) as { type: string; data: unknown }
          handleMessage(data)
        } catch (e) {
          console.error('Error parsing message:', e)
        }
      }
    }

    ws.onclose = () => {
      isConnected.value = false
      console.log('WebSocket disconnected')
      setTimeout(connect, 1000)
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  function handleMessage(msg: { type: string; data: unknown }) {
    switch (msg.type) {
      case ServerMsgType.decode:
        messages.value.unshift(msg.data as DecodeMessage)
        if (messages.value.length > 100) {
          messages.value.pop()
        }
        break
      case ServerMsgType.status:
        status.value = msg.data as StatusMessage
        break
      case ServerMsgType.audioDevices:
        audioDevices.value = msg.data as AudioDevice[]
        break
      default:
        break
    }
  }

  function sendMessage(type: ClientMsgTypeName, data: ClientEnvelope['data']) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const envelope: ClientEnvelope = { type, data } as ClientEnvelope
      ws.send(JSON.stringify(envelope))
    } else {
      console.error('WebSocket not connected')
    }
  }

  function fetchAudioDevices() {
    sendMessage(ClientMsgType.getAudioDevices, {})
  }

  function setAudioDevice(id: string) {
    sendMessage(ClientMsgType.setAudioDevice, { id })
  }

  return {
    isConnected,
    messages,
    status,
    connect,
    sendMessage,
    audioQueue,
    audioContext,
    audioDevices,
    fetchAudioDevices,
    setAudioDevice,
  }
})
