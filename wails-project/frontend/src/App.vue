<template>
  <UApp class="bg-white min-h-screen">
    <div class="container mx-auto p-8 max-w-6xl">
      <div class="mb-8">
        <h1 class="text-4xl font-bold mb-3 text-gray-800">Graphics Tablet Configuration</h1>
        <p class="text-gray-600">Configure your tablet buttons, touch ring and settings</p>
      </div>

      <!-- Error banner -->
      <div
        v-if="errorMessage"
        class="mb-6 flex items-start justify-between gap-4 bg-red-50 border border-red-200 rounded-lg p-4"
      >
        <div>
          <h3 class="font-medium text-red-800 mb-1">Something went wrong</h3>
          <p class="text-red-700 text-sm whitespace-pre-wrap">{{ errorMessage }}</p>
        </div>
        <UButton variant="ghost" size="sm" @click="errorMessage = ''">Dismiss</UButton>
      </div>

      <!-- Device List -->
      <UCard class="bg-white border border-gray-200 shadow-sm mb-6">
        <template #header>
          <div class="flex justify-between items-center">
            <h2 class="text-xl font-semibold text-gray-800">Select Device</h2>
            <UButton
              @click="refreshDevices"
              :loading="loadingDevices"
              variant="outline"
              size="sm"
            >
              Refresh
            </UButton>
          </div>
        </template>

        <!-- Loading State -->
        <div v-if="loadingDevices" class="p-8 text-center">
          <p class="text-gray-500">Loading devices...</p>
        </div>

        <!-- Device List -->
        <div v-else-if="devices.length" class="space-y-3">
          <div
            v-for="device in devices"
            :key="device.id"
            class="flex items-center justify-between p-4 border rounded-lg transition-colors cursor-pointer"
            :class="selectedDevice?.id === device.id
              ? 'bg-blue-50 border-blue-300'
              : 'bg-gray-50 border-gray-100 hover:bg-gray-100'"
            @click="selectDevice(device)"
          >
            <div class="flex-1">
              <h3 class="font-medium text-gray-900 mb-1">{{ device.name }}</h3>
              <div class="flex items-center gap-4 text-sm text-gray-600">
                <span>ID: {{ device.id }}</span>
                <span class="px-2 py-1 bg-blue-100 text-blue-700 rounded-md font-medium">
                  {{ device.type }}
                </span>
              </div>
            </div>
            <div v-if="selectedDevice?.id === device.id" class="ml-4">
              <span class="text-blue-600 text-sm font-medium">Selected</span>
            </div>
          </div>
        </div>

        <!-- No Devices -->
        <div v-else class="p-8 text-center">
          <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
            <h3 class="font-medium text-yellow-800 mb-2">No devices detected</h3>
            <p class="text-yellow-700 text-sm">
              Make sure your tablet is connected and xsetwacom is installed.
            </p>
            <div class="mt-4 text-xs text-yellow-600">
              <p>To install xsetwacom:</p>
              <code class="bg-yellow-100 px-2 py-1 rounded mt-1 block">
                sudo apt install xserver-xorg-input-wacom
              </code>
            </div>
          </div>
        </div>

        <template #footer v-if="devices.length">
          <p class="text-sm text-gray-500">
            Found {{ devices.length }} device{{ devices.length > 1 ? 's' : '' }}. Click a device to configure it.
          </p>
        </template>
      </UCard>

      <!-- Tablet Mockup -->
      <UCard v-if="selectedDevice" class="bg-white border border-gray-200 shadow-sm mb-6">
        <template #header>
          <h2 class="text-xl font-semibold text-gray-800">Tablet Overview</h2>
        </template>
        <div class="max-w-xl mx-auto">
          <TabletMockup
            :buttons="buttons"
            :has-wheel="wheelActions.length > 0"
            :highlighted-button="highlightedButton"
            :wheel-highlighted="wheelHighlighted"
            @select-button="focusButton"
            @select-wheel="focusWheel"
          />
        </div>
      </UCard>

      <!-- Button Configuration -->
      <UCard v-if="selectedDevice" class="bg-white border border-gray-200 shadow-sm mb-6">
        <template #header>
          <div class="flex justify-between items-center">
            <div>
              <h2 class="text-xl font-semibold text-gray-800">Button Configuration</h2>
              <p class="text-sm text-gray-600 mt-1">{{ selectedDevice.name }}</p>
            </div>
            <UButton
              @click="loadDeviceConfig"
              :loading="loadingButtons"
              variant="outline"
              size="sm"
            >
              Refresh Buttons
            </UButton>
          </div>
        </template>

        <!-- Loading Buttons -->
        <div v-if="loadingButtons" class="p-8 text-center">
          <p class="text-gray-500">Loading buttons...</p>
        </div>

        <!-- Button List -->
        <div v-else-if="buttons.length" class="space-y-4">
          <div
            v-for="button in buttons"
            :key="button.number"
            :id="`button-row-${button.number}`"
            class="flex items-center gap-4 p-4 border rounded-lg transition-colors"
            :class="highlightedButton === button.number
              ? 'bg-blue-50 border-blue-300'
              : 'bg-gray-50 border-gray-100'"
          >
            <div class="w-20 text-center">
              <span class="bg-blue-100 text-blue-800 px-3 py-1 rounded-full text-sm font-medium">
                {{ button.number }}
              </span>
            </div>
            <div class="flex-1">
              <UInput
                v-model="buttonActions[button.number]"
                :placeholder="`Enter action (e.g., ctrl+z, space, alt+tab)`"
                class="w-full"
              />
              <p class="text-xs text-gray-500 mt-1">
                Current: {{ button.currentAction || 'Not set' }}
              </p>
            </div>
            <div class="flex gap-2">
              <UButton
                @click="applyButtonAction(button.number)"
                :loading="applyingButton === button.number"
                size="sm"
                :disabled="!buttonActions[button.number]"
              >
                Apply
              </UButton>
              <UButton
                @click="resetButton(button.number)"
                :loading="resettingButton === button.number"
                variant="outline"
                size="sm"
              >
                Reset
              </UButton>
            </div>
          </div>
        </div>

        <!-- No Buttons -->
        <div v-else class="p-8 text-center">
          <div class="bg-orange-50 border border-orange-200 rounded-lg p-6">
            <h3 class="font-medium text-orange-800 mb-2">No configurable buttons found</h3>
            <p class="text-orange-700 text-sm">
              This device may not have programmable buttons, or they may not be supported by xsetwacom.
              Pad buttons are usually on the device of type <strong>PAD</strong>.
            </p>
          </div>
        </div>

        <template #footer v-if="buttons.length">
          <div class="text-sm text-gray-600 space-y-2">
            <p><strong>Tip:</strong> Common key combinations:</p>
            <div class="flex flex-wrap gap-2 text-xs">
              <code class="bg-gray-100 px-2 py-1 rounded">ctrl+z</code>
              <code class="bg-gray-100 px-2 py-1 rounded">ctrl+y</code>
              <code class="bg-gray-100 px-2 py-1 rounded">space</code>
              <code class="bg-gray-100 px-2 py-1 rounded">alt+tab</code>
              <code class="bg-gray-100 px-2 py-1 rounded">shift+ctrl+z</code>
              <code class="bg-gray-100 px-2 py-1 rounded">f1</code>
              <code class="bg-gray-100 px-2 py-1 rounded">button 3</code>
            </div>
          </div>
        </template>
      </UCard>

      <!-- Touch Ring / Wheel Configuration -->
      <UCard
        v-if="selectedDevice && wheelActions.length"
       
        class="bg-white border shadow-sm transition-colors"
        :class="wheelHighlighted ? 'border-blue-300' : 'border-gray-200'"
      >
        <template #header>
          <div>
            <h2 class="text-xl font-semibold text-gray-800">Touch Ring &amp; Wheel</h2>
            <p class="text-sm text-gray-600 mt-1">
              Assign actions to each rotation direction — e.g. <code class="bg-gray-100 px-1 rounded">ctrl+plus</code> /
              <code class="bg-gray-100 px-1 rounded">ctrl+minus</code> for zoom, or
              <code class="bg-gray-100 px-1 rounded">bracketleft</code> / <code class="bg-gray-100 px-1 rounded">bracketright</code>
              for brush size.
            </p>
          </div>
        </template>

        <div class="space-y-4">
          <div
            v-for="wheel in wheelActions"
            :key="wheel.property"
            class="flex items-center gap-4 p-4 bg-gray-50 border border-gray-100 rounded-lg"
          >
            <div class="w-56 shrink-0">
              <p class="text-sm font-medium text-gray-800">{{ wheel.label }}</p>
              <p class="text-xs text-gray-400">{{ wheel.property }}</p>
            </div>
            <div class="flex-1">
              <UInput
                v-model="wheelInputs[wheel.property]"
                placeholder="Enter action (e.g., ctrl+plus)"
                class="w-full"
              />
              <p class="text-xs text-gray-500 mt-1">
                Current: {{ wheel.currentAction || 'Not set' }}
              </p>
            </div>
            <div class="flex gap-2">
              <UButton
                @click="applyWheelAction(wheel.property)"
                :loading="applyingWheel === wheel.property"
                size="sm"
                :disabled="!wheelInputs[wheel.property]"
              >
                Apply
              </UButton>
              <UButton
                @click="resetWheelAction(wheel.property)"
                :loading="resettingWheel === wheel.property"
                variant="outline"
                size="sm"
              >
                Reset
              </UButton>
            </div>
          </div>
        </div>

        <template #footer>
          <p class="text-sm text-gray-500">
            Reset restores the default scroll behaviour for that direction.
          </p>
        </template>
      </UCard>
    </div>
  </UApp>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { GetDevices, GetDeviceButtons, GetWheelActions, SetButtonAction, SetWheelAction } from '../wailsjs/go/main/App'
import { main } from '../wailsjs/go/models'
import TabletMockup from './components/TabletMockup.vue'

// Reactive state
const devices = ref<main.Device[]>([])
const selectedDevice = ref<main.Device | null>(null)
const buttons = ref<main.Button[]>([])
const wheelActions = ref<main.WheelAction[]>([])
const buttonActions = reactive<Record<number, string>>({})
const wheelInputs = reactive<Record<string, string>>({})
const errorMessage = ref('')

// Loading states
const loadingDevices = ref(false)
const loadingButtons = ref(false)
const applyingButton = ref<number | null>(null)
const resettingButton = ref<number | null>(null)
const applyingWheel = ref<string | null>(null)
const resettingWheel = ref<string | null>(null)

// Mockup interaction state
const highlightedButton = ref<number | null>(null)
const wheelHighlighted = ref(false)

onMounted(() => {
  refreshDevices()
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
  } catch (error) {
    showError('Failed to get devices', error)
    devices.value = []
  } finally {
    loadingDevices.value = false
  }
}

// Select a device and load its buttons + wheel actions
async function selectDevice(device: main.Device) {
  if (selectedDevice.value?.id === device.id) return
  selectedDevice.value = device
  highlightedButton.value = null
  wheelHighlighted.value = false
  await loadDeviceConfig()
}

// Load buttons and touch ring actions for the selected device
async function loadDeviceConfig() {
  if (!selectedDevice.value) return

  loadingButtons.value = true
  try {
    const [btns, wheels] = await Promise.all([
      GetDeviceButtons(selectedDevice.value.name),
      GetWheelActions(selectedDevice.value.name),
    ])
    buttons.value = btns
    wheelActions.value = wheels

    // Initialize inputs, keeping anything the user has typed
    btns.forEach(b => {
      if (!(b.number in buttonActions)) buttonActions[b.number] = ''
    })
    wheels.forEach(w => {
      if (!(w.property in wheelInputs)) wheelInputs[w.property] = ''
    })
  } catch (error) {
    showError('Failed to load device configuration', error)
    buttons.value = []
    wheelActions.value = []
  } finally {
    loadingButtons.value = false
  }
}

// Apply action to a specific button
async function applyButtonAction(buttonNumber: number) {
  if (!selectedDevice.value || !buttonActions[buttonNumber]) return

  applyingButton.value = buttonNumber
  errorMessage.value = ''
  try {
    await SetButtonAction(selectedDevice.value.name, buttonNumber, buttonActions[buttonNumber])
    buttonActions[buttonNumber] = ''
    await loadDeviceConfig()
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
    await loadDeviceConfig()
  } catch (error) {
    showError(`Failed to reset button ${buttonNumber}`, error)
  } finally {
    resettingButton.value = null
  }
}

// Apply action to a wheel / touch ring direction
async function applyWheelAction(property: string) {
  if (!selectedDevice.value || !wheelInputs[property]) return

  applyingWheel.value = property
  errorMessage.value = ''
  try {
    await SetWheelAction(selectedDevice.value.name, property, wheelInputs[property])
    wheelInputs[property] = ''
    await loadDeviceConfig()
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
    await loadDeviceConfig()
  } catch (error) {
    showError(`Failed to reset ${property}`, error)
  } finally {
    resettingWheel.value = null
  }
}

// Mockup click handlers: highlight and scroll to the matching config row
async function focusButton(buttonNumber: number) {
  highlightedButton.value = buttonNumber
  wheelHighlighted.value = false
  await nextTick()
  document.getElementById(`button-row-${buttonNumber}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

async function focusWheel() {
  wheelHighlighted.value = true
  highlightedButton.value = null
  await nextTick()
  window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' })
}
</script>
