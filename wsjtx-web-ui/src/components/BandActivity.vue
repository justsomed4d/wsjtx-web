<script setup lang="ts">
import { useWsjtxStore } from '@/stores/wsjtx'
import { storeToRefs } from 'pinia'
import { ClientMsgType, type DecodeMessage } from '@/types/ws'

const store = useWsjtxStore()
const { messages } = storeToRefs(store)

function formatTime(ms: number) {
  // ms is milliseconds since midnight
  const seconds = Math.floor((ms / 1000) % 60)
  const minutes = Math.floor((ms / (1000 * 60)) % 60)
  const hours = Math.floor((ms / (1000 * 60 * 60)) % 24)

  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

function handleReply(msg: DecodeMessage) {
  store.sendMessage(ClientMsgType.reply, {
    Time: msg.Time,
    SNR: msg.SNR,
    DeltaTime: msg.DeltaTime,
    DeltaFrequency: msg.DeltaFrequency,
    Mode: msg.Mode,
    Message: msg.Message,
    LowConfidence: msg.LowConfidence ?? false,
    Modifiers: 0,
  })
}
</script>

<template>
  <div class="bg-gray-900 text-gray-300 p-4 rounded-lg overflow-y-auto max-h-[500px]">
    <h2 class="text-xl mb-4 font-light">Band Activity</h2>
    <table class="w-full border-collapse font-mono text-sm">
      <thead>
        <tr>
          <th class="p-3 text-left border-b border-gray-700">Time</th>
          <th class="p-3 text-left border-b border-gray-700">dB</th>
          <th class="p-3 text-left border-b border-gray-700">DT</th>
          <th class="p-3 text-left border-b border-gray-700">Freq</th>
          <th class="p-3 text-left border-b border-gray-700">Mode</th>
          <th class="p-3 text-left border-b border-gray-700">Message</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(msg, index) in messages" :key="index" @dblclick="handleReply(msg)" class="cursor-pointer hover:bg-gray-700 even:bg-gray-800 transition-colors">
          <td class="p-3">{{ formatTime(msg.Time) }}</td>
          <td class="p-3">{{ msg.SNR }}</td>
          <td class="p-3">{{ msg.DeltaTime.toFixed(1) }}</td>
          <td class="p-3">{{ msg.DeltaFrequency }}</td>
          <td class="p-3">{{ msg.Mode }}</td>
          <td class="p-3 text-sky-300">{{ msg.Message }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
