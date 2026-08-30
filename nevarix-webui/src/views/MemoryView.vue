<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/client'
import LineChart from '@/components/LineChart.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatBytes, formatPercent } from '@/lib/utils'

const payload = ref<any>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    payload.value = await api.memory()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

const seriesLabels = computed(() =>
  (payload.value?.series ?? []).map((p: { timestamp: string }) => new Date(p.timestamp).toLocaleTimeString()),
)
const seriesData = computed(() => (payload.value?.series ?? []).map((p: { value: number }) => p.value))

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
    <div class="grid gap-4 sm:grid-cols-3">
      <Card>
        <CardHeader><CardTitle>Used</CardTitle></CardHeader>
        <CardContent class="text-xl font-semibold">{{ formatBytes(payload?.memoryUsedBytes) }}</CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Available</CardTitle></CardHeader>
        <CardContent class="text-xl font-semibold">{{ formatBytes(payload?.memoryAvailableBytes) }}</CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Total</CardTitle></CardHeader>
        <CardContent class="text-xl font-semibold">{{ formatBytes(payload?.memoryTotalBytes) }}</CardContent>
      </Card>
    </div>
    <Card>
      <CardHeader><CardTitle>Memory {{ formatPercent(payload?.memoryUsedPercent) }}</CardTitle></CardHeader>
      <CardContent>
        <LineChart :labels="seriesLabels" :datasets="[{ label: 'Used %', data: seriesData, color: '#8fd4a0' }]" />
      </CardContent>
    </Card>
  </div>
</template>
