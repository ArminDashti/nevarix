<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { formatBytes, formatBps, formatDuration, formatPercent } from '@/lib/utils'

const data = ref<Record<string, unknown> | null>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    data.value = await api.dashboard()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

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
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
      <Card>
        <CardHeader><CardTitle>CPU</CardTitle></CardHeader>
        <CardContent class="text-2xl font-semibold">{{ formatPercent(data?.cpuPercent as number) }}</CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Memory</CardTitle></CardHeader>
        <CardContent class="text-2xl font-semibold">
          {{ formatPercent(data?.memoryUsedPercent as number) }}
          <p class="mt-1 text-xs font-normal text-muted-foreground">
            {{ formatBytes(data?.memoryUsedBytes as number) }} / {{ formatBytes(data?.memoryTotalBytes as number) }}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Disk</CardTitle></CardHeader>
        <CardContent class="text-2xl font-semibold">
          {{ formatBytes(data?.diskUsedBytes as number) }}
          <p class="mt-1 text-xs font-normal text-muted-foreground">of {{ formatBytes(data?.diskTotalBytes as number) }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Bandwidth</CardTitle></CardHeader>
        <CardContent class="text-sm">
          <div>↓ {{ formatBps(data?.bandwidthRxBps as number) }}</div>
          <div class="mt-1">↑ {{ formatBps(data?.bandwidthTxBps as number) }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader><CardTitle>Uptime</CardTitle></CardHeader>
        <CardContent class="text-2xl font-semibold">{{ formatDuration(data?.uptimeSeconds as number) }}</CardContent>
      </Card>
    </div>
  </div>
</template>
