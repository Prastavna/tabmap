<template>
  <UApp class="bg-white">
    <div class="h-screen flex flex-col max-w-7xl mx-auto w-full p-5 gap-4 overflow-hidden">
      <!-- Header: title + device dropdown -->
      <div class="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 class="text-2xl font-bold text-gray-800">TabMap</h1>
          <p class="text-sm text-gray-500">Buttons, touch ring &amp; wheel — changes are saved and re-applied on startup</p>
        </div>
        <div class="flex items-center gap-2">
          <select
            v-model="selectedDeviceId"
            :disabled="loadingDevices || !devices.length"
            class="w-80 h-9 px-3 text-sm border border-gray-300 rounded-md bg-white text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
          >
            <option :value="undefined" disabled>Select a device…</option>
            <option v-for="device in devices" :key="device.id" :value="device.id">
              {{ device.name }} ({{ device.type }})
            </option>
          </select>
          <button
            @click="refreshDevices"
            :disabled="loadingDevices"
            class="h-9 px-3 text-sm font-medium rounded-md border border-gray-300 text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
          >
            {{ loadingDevices ? 'Loading…' : 'Refresh' }}
          </button>
        </div>
      </div>

      <!-- Error banner -->
      <div
        v-if="errorMessage"
        class="flex items-center justify-between gap-4 bg-red-50 border border-red-200 rounded-lg px-4 py-2"
      >
        <p class="text-red-700 text-sm truncate" :title="errorMessage">{{ errorMessage }}</p>
        <button class="text-sm text-red-600 hover:text-red-800 shrink-0" @click="errorMessage = ''">Dismiss</button>
      </div>

      <!-- Main two-column area -->
      <div class="grid lg:grid-cols-2 gap-4 flex-1 min-h-0">
        <!-- Left: tablet mockup -->
        <div class="border border-gray-200 rounded-lg shadow-sm p-5 flex flex-col justify-center min-h-0 overflow-hidden">
          <template v-if="selectedDevice">
            <TabletMockup
              :buttons="buttons"
              :has-wheel="wheelActions.length > 0"
              :highlighted-button="highlightedButton"
              :wheel-highlighted="wheelHighlighted"
              @select-button="focusButton"
              @select-wheel="focusWheel"
            />
          </template>
          <template v-else-if="!loadingDevices && !devices.length">
            <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-5 text-center">
              <h3 class="font-medium text-yellow-800 mb-1">No devices detected</h3>
              <p class="text-yellow-700 text-sm mb-2">Make sure your tablet is connected and xsetwacom is installed.</p>
              <code class="bg-yellow-100 text-yellow-700 text-xs px-2 py-1 rounded">
                sudo apt install xserver-xorg-input-wacom
              </code>
            </div>
          </template>
          <template v-else>
            <p class="text-gray-400 text-center">Select a device to start configuring.</p>
          </template>
        </div>

        <!-- Right: configuration panel -->
        <div class="border border-gray-200 rounded-lg shadow-sm flex flex-col min-h-0">
          <div class="flex items-center justify-between gap-2 px-4 py-3 border-b border-gray-100">
            <!-- Section toggle -->
            <div class="flex bg-gray-100 rounded-lg p-0.5">
              <button
                class="px-3 py-1 text-sm rounded-md transition-colors"
                :class="activeSection === 'buttons' ? 'bg-white shadow text-gray-800 font-medium' : 'text-gray-500 hover:text-gray-700'"
                @click="activeSection = 'buttons'"
              >
                Buttons <span v-if="buttons.length" class="text-xs text-gray-400">({{ buttons.length }})</span>
              </button>
              <button
                class="px-3 py-1 text-sm rounded-md transition-colors"
                :class="activeSection === 'wheel' ? 'bg-white shadow text-gray-800 font-medium' : 'text-gray-500 hover:text-gray-700'"
                @click="activeSection = 'wheel'"
              >
                Touch Ring &amp; Wheel <span v-if="wheelActions.length" class="text-xs text-gray-400">({{ wheelActions.length }})</span>
              </button>
            </div>
            <div class="flex items-center gap-2">
              <button
                v-if="selectedDevice"
                @click="applyAll"
                :disabled="!pendingCount || applyingAll || hasConflicts"
                :title="hasConflicts ? 'Resolve duplicate mappings first' : ''"
                class="h-8 px-3 text-sm font-medium rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed"
              >
                {{ applyingAll ? 'Applying…' : `Apply All${pendingCount ? ` (${pendingCount})` : ''}` }}
              </button>
              <button
                v-if="selectedDevice"
                @click="loadDeviceConfig()"
                class="h-8 px-2 text-sm text-gray-500 hover:text-gray-700 rounded-md hover:bg-gray-50"
              >
                Reload
              </button>
            </div>
          </div>

          <div ref="panelEl" class="flex-1 overflow-y-auto p-3 space-y-2">
            <!-- Empty / loading states -->
            <p v-if="!selectedDevice" class="text-gray-400 text-sm text-center py-8">No device selected.</p>
            <p v-else-if="loadingButtons" class="text-gray-400 text-sm text-center py-8">Loading…</p>

            <!-- Buttons section -->
            <template v-else-if="activeSection === 'buttons'">
              <p v-if="!buttons.length" class="text-orange-600 text-sm text-center py-8">
                No configurable buttons found — pad buttons are usually on the <strong>PAD</strong> device.
              </p>
              <div
                v-for="button in buttons"
                :key="button.number"
                :id="`button-row-${button.number}`"
                class="border rounded-lg px-3 py-2 transition-colors"
                :class="highlightedButton === button.number ? 'bg-blue-50 border-blue-300' : 'bg-gray-50 border-gray-100'"
              >
                <div class="flex items-center gap-3">
                  <span class="bg-blue-100 text-blue-800 w-8 h-8 shrink-0 rounded-full text-sm font-medium flex items-center justify-center">
                    {{ button.number }}
                  </span>
                  <input
                    v-model="buttonActions[button.number]"
                    placeholder="e.g. ctrl+z, space, alt+tab"
                    class="flex-1 min-w-0 h-8 px-2.5 text-sm rounded-md border focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors"
                    :class="inputStateClass(buttonActions[button.number], savedButtons[button.number])"
                    @keydown.enter="applyButtonAction(button.number)"
                  />
                  <button
                    @click="applyButtonAction(button.number)"
                    :disabled="!buttonActions[button.number] || applyingButton === button.number"
                    class="h-8 px-3 text-xs font-medium rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed shrink-0"
                  >
                    {{ applyingButton === button.number ? '…' : 'Apply' }}
                  </button>
                  <button
                    @click="resetButton(button.number)"
                    :disabled="resettingButton === button.number"
                    class="h-8 px-3 text-xs font-medium rounded-md border border-gray-300 text-gray-600 bg-white hover:bg-gray-50 disabled:opacity-40 shrink-0"
                  >
                    {{ resettingButton === button.number ? '…' : 'Reset' }}
                  </button>
                </div>
                <p class="text-xs text-gray-400 mt-1 ml-11 truncate">Current: {{ button.currentAction || 'Not set' }}</p>
              </div>
            </template>

            <!-- Touch ring / wheel section -->
            <template v-else>
              <p v-if="!wheelActions.length" class="text-orange-600 text-sm text-center py-8">
                No touch ring or wheel properties found on this device.
              </p>
              <div
                v-for="wheel in wheelActions"
                :key="wheel.property"
                class="border rounded-lg px-3 py-2 bg-gray-50 border-gray-100"
              >
                <div class="flex items-center gap-3">
                  <div class="w-40 shrink-0">
                    <p class="text-sm font-medium text-gray-700 leading-tight">{{ wheel.label }}</p>
                    <p class="text-[11px] text-gray-400">{{ wheel.property }}</p>
                  </div>
                  <input
                    v-model="wheelInputs[wheel.property]"
                    placeholder="e.g. ctrl+plus"
                    class="flex-1 min-w-0 h-8 px-2.5 text-sm rounded-md border focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors"
                    :class="inputStateClass(wheelInputs[wheel.property], savedWheels[wheel.property])"
                    @keydown.enter="applyWheelAction(wheel.property)"
                  />
                  <button
                    @click="applyWheelAction(wheel.property)"
                    :disabled="!wheelInputs[wheel.property] || applyingWheel === wheel.property"
                    class="h-8 px-3 text-xs font-medium rounded-md bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed shrink-0"
                  >
                    {{ applyingWheel === wheel.property ? '…' : 'Apply' }}
                  </button>
                  <button
                    @click="resetWheelAction(wheel.property)"
                    :disabled="resettingWheel === wheel.property"
                    class="h-8 px-3 text-xs font-medium rounded-md border border-gray-300 text-gray-600 bg-white hover:bg-gray-50 disabled:opacity-40 shrink-0"
                  >
                    {{ resettingWheel === wheel.property ? '…' : 'Reset' }}
                  </button>
                </div>
                <p class="text-xs text-gray-400 mt-1 ml-[10.75rem] truncate">Current: {{ wheel.currentAction || 'Not set' }}</p>
              </div>
            </template>
          </div>

          <div class="px-4 py-2 border-t border-gray-100 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500">
            <span class="flex items-center gap-1">
              <span class="w-3 h-3 rounded border border-yellow-300 bg-yellow-50 inline-block"></span> unapplied
            </span>
            <span class="flex items-center gap-1">
              <span class="w-3 h-3 rounded border border-red-300 bg-red-50 inline-block"></span> duplicate mapping
            </span>
            <span class="text-gray-300">|</span>
            <span>Examples:</span>
            <code class="bg-gray-100 px-1.5 py-0.5 rounded">ctrl+z</code>
            <code class="bg-gray-100 px-1.5 py-0.5 rounded">space</code>
            <code class="bg-gray-100 px-1.5 py-0.5 rounded">ctrl+plus</code>
            <code class="bg-gray-100 px-1.5 py-0.5 rounded">bracketleft</code>
            <code class="bg-gray-100 px-1.5 py-0.5 rounded">button 3</code>
          </div>
        </div>
      </div>
    </div>
  </UApp>
</template>

<script lang="ts" setup>
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue'
import { GetDevices, GetDeviceButtons, GetWheelActions, GetSavedMappings, SetButtonAction, SetWheelAction } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'
import TabletMockup from './components/TabletMockup.vue'

// Reactive state
const devices = ref<main.Device[]>([])
const selectedDeviceId = ref<string | undefined>(undefined)
const buttons = ref<main.Button[]>([])
const wheelActions = ref<main.WheelAction[]>([])
const buttonActions = reactive<Record<number, string>>({})
const wheelInputs = reactive<Record<string, string>>({})
const savedButtons = reactive<Record<number, string>>({})
const savedWheels = reactive<Record<string, string>>({})
const errorMessage = ref('')
const activeSection = ref<'buttons' | 'wheel'>('buttons')
const panelEl = ref<HTMLElement | null>(null)

// Loading states
const loadingDevices = ref(false)
const loadingButtons = ref(false)
const applyingButton = ref<number | null>(null)
const resettingButton = ref<number | null>(null)
const applyingWheel = ref<string | null>(null)
const resettingWheel = ref<string | null>(null)
const applyingAll = ref(false)

// Mockup interaction state
const highlightedButton = ref<number | null>(null)
const wheelHighlighted = ref(false)

const selectedDevice = computed(() =>
  devices.value.find(d => d.id === selectedDeviceId.value) ?? null
)

watch(selectedDeviceId, () => {
  highlightedButton.value = null
  wheelHighlighted.value = false
  activeSection.value = 'buttons'
  loadDeviceConfig()
})

onMounted(() => {
  refreshDevices()
})

// --- Input state: unsaved (yellow) / conflicting (red) ---

function normalize(action: string | undefined): string {
  return (action ?? '').trim().toLowerCase()
}

// How many controls on this device share each non-empty action
const actionCounts = computed(() => {
  const counts: Record<string, number> = {}
  buttons.value.forEach(b => {
    const v = normalize(buttonActions[b.number])
    if (v) counts[v] = (counts[v] ?? 0) + 1
  })
  wheelActions.value.forEach(w => {
    const v = normalize(wheelInputs[w.property])
    if (v) counts[v] = (counts[v] ?? 0) + 1
  })
  return counts
})

const hasConflicts = computed(() => Object.values(actionCounts.value).some(c => c > 1))

function isDirty(value: string | undefined, saved: string | undefined): boolean {
  return normalize(value) !== normalize(saved)
}

function inputStateClass(value: string | undefined, saved: string | undefined): string {
  const v = normalize(value)
  if (v && actionCounts.value[v] > 1) return 'bg-red-50 border-red-300'
  if (isDirty(value, saved)) return 'bg-yellow-50 border-yellow-300'
  return 'bg-white border-gray-200'
}

// Unapplied changes (dirty inputs), used by the Apply All button
const pendingCount = computed(() => {
  let n = 0
  buttons.value.forEach(b => { if (isDirty(buttonActions[b.number], savedButtons[b.number])) n++ })
  wheelActions.value.forEach(w => { if (isDirty(wheelInputs[w.property], savedWheels[w.property])) n++ })
  return n
})

function showError(prefix: string, error: unknown) {
  console.error(prefix, error)
  errorMessage.value = `${prefix}: ${String(error)}`
}

// Refresh device list
async function refreshDevices() {
  loadingDevices.value = true
  errorMessage.value = ''
  try {
    devices.value = await GetDevices()
    // Keep the current selection if still present, otherwise pick the first PAD device
    if (!devices.value.some(d => d.id === selectedDeviceId.value)) {
      const preferred = devices.value.find(d => d.type.toUpperCase().includes('PAD')) ?? devices.value[0]
      selectedDeviceId.value = preferred?.id
    }
  } catch (error) {
    showError('Failed to get devices', error)
    devices.value = []
  } finally {
    loadingDevices.value = false
  }
}

// Load buttons, wheel actions and saved mappings for the selected device.
// silent = refresh data in place without the loading state (no flicker) and
// without touching what the user has typed.
async function loadDeviceConfig(silent = false) {
  if (!selectedDevice.value) {
    buttons.value = []
    wheelActions.value = []
    return
  }

  if (!silent) loadingButtons.value = true
  try {
    const name = selectedDevice.value.name
    const [btns, wheels, saved] = await Promise.all([
      GetDeviceButtons(name),
      GetWheelActions(name),
      GetSavedMappings(name),
    ])
    buttons.value = btns
    wheelActions.value = wheels

    btns.forEach((b: main.Button) => {
      savedButtons[b.number] = saved.buttons?.[b.number] ?? ''
      if (!silent) buttonActions[b.number] = savedButtons[b.number]
    })
    wheels.forEach((w: main.WheelAction) => {
      savedWheels[w.property] = saved.wheels?.[w.property] ?? ''
      if (!silent) wheelInputs[w.property] = savedWheels[w.property]
    })
  } catch (error) {
    showError('Failed to load device configuration', error)
    if (!silent) {
      buttons.value = []
      wheelActions.value = []
    }
  } finally {
    if (!silent) loadingButtons.value = false
  }
}

// Apply action to a specific button (input is kept — it reflects the saved mapping)
async function applyButtonAction(buttonNumber: number) {
  if (!selectedDevice.value || !buttonActions[buttonNumber]) return

  applyingButton.value = buttonNumber
  errorMessage.value = ''
  try {
    await SetButtonAction(selectedDevice.value.name, buttonNumber, buttonActions[buttonNumber])
    await loadDeviceConfig(true)
  } catch (error) {
    showError(`Failed to set button ${buttonNumber}`, error)
  } finally {
    applyingButton.value = null
  }
}

// Reset a button to its default state
async function resetButton(buttonNumber: number) {
  if (!selectedDevice.value) return

  resettingButton.value = buttonNumber
  errorMessage.value = ''
  try {
    await SetButtonAction(selectedDevice.value.name, buttonNumber, '')
    buttonActions[buttonNumber] = ''
    await loadDeviceConfig(true)
  } catch (error) {
    showError(`Failed to reset button ${buttonNumber}`, error)
  } finally {
    resettingButton.value = null
  }
}

// Apply action to a wheel / touch ring direction (input is kept)
async function applyWheelAction(property: string) {
  if (!selectedDevice.value || !wheelInputs[property]) return

  applyingWheel.value = property
  errorMessage.value = ''
  try {
    await SetWheelAction(selectedDevice.value.name, property, wheelInputs[property])
    await loadDeviceConfig(true)
  } catch (error) {
    showError(`Failed to set ${property}`, error)
  } finally {
    applyingWheel.value = null
  }
}

// Reset a wheel direction to default scrolling
async function resetWheelAction(property: string) {
  if (!selectedDevice.value) return

  resettingWheel.value = property
  errorMessage.value = ''
  try {
    await SetWheelAction(selectedDevice.value.name, property, '')
    wheelInputs[property] = ''
    await loadDeviceConfig(true)
  } catch (error) {
    showError(`Failed to reset ${property}`, error)
  } finally {
    resettingWheel.value = null
  }
}

// Apply every unapplied (dirty) mapping in one go. Clearing a previously
// saved mapping counts as a change and resets that control to its default.
async function applyAll() {
  if (!selectedDevice.value || applyingAll.value) return

  applyingAll.value = true
  errorMessage.value = ''
  const name = selectedDevice.value.name
  const errors: string[] = []

  for (const b of buttons.value) {
    if (!isDirty(buttonActions[b.number], savedButtons[b.number])) continue
    try {
      await SetButtonAction(name, b.number, buttonActions[b.number] ?? '')
    } catch (error) {
      errors.push(`button ${b.number}: ${String(error)}`)
    }
  }
  for (const w of wheelActions.value) {
    if (!isDirty(wheelInputs[w.property], savedWheels[w.property])) continue
    try {
      await SetWheelAction(name, w.property, wheelInputs[w.property] ?? '')
    } catch (error) {
      errors.push(`${w.property}: ${String(error)}`)
    }
  }

  await loadDeviceConfig(true)
  if (errors.length) {
    errorMessage.value = `Some mappings failed — ${errors.join('; ')}`
  }
  applyingAll.value = false
}

// Mockup click handlers: switch section, highlight, scroll within the panel
async function focusButton(buttonNumber: number) {
  highlightedButton.value = buttonNumber
  wheelHighlighted.value = false
  activeSection.value = 'buttons'
  await nextTick()
  document.getElementById(`button-row-${buttonNumber}`)?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}

async function focusWheel() {
  wheelHighlighted.value = true
  highlightedButton.value = null
  activeSection.value = 'wheel'
  await nextTick()
  panelEl.value?.scrollTo({ top: 0, behavior: 'smooth' })
}
</script>
