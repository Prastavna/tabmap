<template>
  <div class="tablet-mockup select-none">
    <svg viewBox="0 0 640 400" class="w-full h-auto" role="img" aria-label="Graphics tablet illustration">
      <defs>
        <linearGradient id="tablet-body" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#3f4652" />
          <stop offset="100%" stop-color="#2b303a" />
        </linearGradient>
        <linearGradient id="tablet-surface" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stop-color="#22262e" />
          <stop offset="100%" stop-color="#181b21" />
        </linearGradient>
        <linearGradient id="pen-body" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0%" stop-color="#4a5160" />
          <stop offset="100%" stop-color="#343a45" />
        </linearGradient>
      </defs>

      <!-- Tablet body -->
      <rect x="10" y="20" width="560" height="360" rx="24" fill="url(#tablet-body)" stroke="#1a1d23" stroke-width="2" />

      <!-- Express keys -->
      <g v-for="(slot, i) in buttonSlots" :key="'btn-' + i">
        <rect
          :x="34" :y="44 + i * 34" width="60" height="26" rx="6"
          class="cursor-pointer transition-all duration-150"
          :fill="slot && highlightedButton === slot.number ? '#3b82f6' : slot ? '#454c59' : '#333842'"
          :stroke="slot && highlightedButton === slot.number ? '#93c5fd' : '#23272f'"
          stroke-width="1.5"
          @click="slot && $emit('select-button', slot.number)"
        />
        <text
          :x="64" :y="61 + i * 34" text-anchor="middle"
          class="pointer-events-none"
          font-size="12" font-family="ui-sans-serif, system-ui, sans-serif"
          :fill="slot ? '#e5e7eb' : '#565d6b'"
        >{{ slot ? slot.number : '·' }}</text>
      </g>

      <!-- Touch ring -->
      <g
        :class="hasWheel ? 'cursor-pointer' : ''"
        @click="hasWheel && $emit('select-wheel')"
      >
        <circle cx="64" cy="330" r="34"
          :fill="wheelHighlighted ? '#1d4ed8' : '#3a414d'"
          :stroke="hasWheel ? (wheelHighlighted ? '#93c5fd' : '#5b6472') : '#2a2f38'"
          stroke-width="2"
        />
        <circle cx="64" cy="330" r="13" fill="#22262e" stroke="#1a1d23" stroke-width="1.5" />
        <!-- ring direction arrows -->
        <path d="M 64 302 l -5 7 h 10 z" :fill="hasWheel ? '#9ca3af' : '#4b5261'" />
        <path d="M 64 358 l -5 -7 h 10 z" :fill="hasWheel ? '#9ca3af' : '#4b5261'" />
        <text v-if="!hasWheel" x="64" y="334" text-anchor="middle" font-size="8" fill="#565d6b"
          font-family="ui-sans-serif, system-ui, sans-serif">n/a</text>
      </g>

      <!-- Active drawing area -->
      <rect x="120" y="44" width="424" height="312" rx="10" fill="url(#tablet-surface)" stroke="#101216" stroke-width="2" />
      <!-- corner marks of active area -->
      <g stroke="#3d444f" stroke-width="2" fill="none">
        <path d="M 140 66 h 14 M 140 66 v 14" />
        <path d="M 524 66 h -14 M 524 66 v 14" />
        <path d="M 140 334 h 14 M 140 334 v -14" />
        <path d="M 524 334 h -14 M 524 334 v -14" />
      </g>
      <!-- status LED -->
      <circle cx="532" cy="34" r="4" fill="#34d399" />

      <!-- sketched squiggle on the surface -->
      <path d="M 190 260 C 240 160, 300 300, 360 200 S 460 160, 480 230"
        fill="none" stroke="#4f5866" stroke-width="3" stroke-linecap="round" stroke-dasharray="1 0" opacity="0.7" />

      <!-- Pen resting to the right -->
      <g transform="translate(588 60) rotate(8)">
        <rect x="0" y="0" width="18" height="230" rx="9" fill="url(#pen-body)" stroke="#1a1d23" stroke-width="1.5" />
        <rect x="2.5" y="80" width="13" height="34" rx="5" fill="#23272f" />
        <path d="M 2 228 L 9 260 L 16 228 Z" fill="#3a414d" stroke="#1a1d23" stroke-width="1.5" />
        <circle cx="9" cy="256" r="2.5" fill="#111318" />
      </g>
    </svg>

    <p class="text-xs text-gray-500 text-center mt-2">
      Illustration only — layout may differ from your tablet.
      <span v-if="buttons.length">Click a key{{ hasWheel ? ' or the touch ring' : '' }} to jump to its settings.</span>
    </p>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { main } from '../../wailsjs/go/models'

const props = defineProps<{
  buttons: main.Button[]
  hasWheel: boolean
  highlightedButton: number | null
  wheelHighlighted: boolean
}>()

defineEmits<{
  (e: 'select-button', buttonNumber: number): void
  (e: 'select-wheel'): void
}>()

const SLOT_COUNT = 7

// Map real device buttons onto the mock's express-key slots
const buttonSlots = computed<(main.Button | null)[]>(() => {
  const slots: (main.Button | null)[] = Array(SLOT_COUNT).fill(null)
  props.buttons.slice(0, SLOT_COUNT).forEach((b, i) => { slots[i] = b })
  return slots
})
</script>
