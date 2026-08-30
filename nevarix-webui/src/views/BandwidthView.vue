<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/client'
import LineChart from '@/components/LineChart.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatBps } from '@/lib/utils'

const payload = ref<any>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    payload.value = await api.bandwidth()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

const seriesLabels = computed(() =>
  (payload.value?.series ?? []).map((p: { timestamp: string }) => new Date(p.timestamp).toLocaleTimeString()),
)
const rx = computed(() => (payload.value?.series ?? []).map((p: { rxBps: number }) => p.rxBps))
const tx = computed(() => (payload.value?.series ?? []).map((p: { txBps: number }) => p.txBps))

onMounted(() => {
  load()
  timer = window.setInterval(load, 3000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <div class="grid gap-4 sm:grid-cols-2">
      <Card>
        <CardHeader><CardTitle>Download</CardTitle></CardHeader>
        <CardContent class="text-xl font-semibold">{{ formatBps(payload?.bandwidthRxBps) }}</CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Upload</CardTitle></CardHeader>
        <CardContent class="text-xl font-semibold">{{ formatBps(payload?.bandwidthTxBps) }}</CardContent>
      </Card>
    </div>
    <Card>
      <CardHeader><CardTitle>Bandwidth</CardTitle></CardHeader>
      <CardContent>
        <LineChart
          :labels="seriesLabels"
          :datasets="[
            { label: 'RX B/s', data: rx, color: '#5ec8ff' },
            { label: 'TX B/s', data: tx, color: '#f0b45c' },
          ]"
        />
      </CardContent>
    </Card>
  </div>
</template>
