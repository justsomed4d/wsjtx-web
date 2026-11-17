import { describe, it, expect, beforeAll } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import BandActivity from '../BandActivity.vue'

beforeAll(() => {
  // wsjtx store constructs AudioContext at setup; jsdom has no real implementation
  class MockAudioContext {
    createGain() {
      return { connect() {}, gain: { value: 1 } }
    }
    createBufferSource() {
      return { connect() {}, start() {}, buffer: null }
    }
    decodeAudioData() {
      return Promise.resolve({})
    }
    get destination() {
      return {}
    }
  }
  ;(window as any).AudioContext = MockAudioContext
  ;(window as any).webkitAudioContext = MockAudioContext
})

describe('BandActivity', () => {
  it('renders the band activity heading', () => {
    const wrapper = mount(BandActivity, {
      global: {
        plugins: [createPinia()],
      },
    })
    expect(wrapper.text()).toContain('Band Activity')
  })
})
