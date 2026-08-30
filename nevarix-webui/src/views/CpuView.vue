<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/client'
import LineChart from '@/components/LineChart.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatPercent } from '@/lib/utils'

const payload = ref<any>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    payload.value = await api.cpu()
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
    <Card>
      <CardHeader><CardTitle>CPU {{ formatPercent(payload?.cpuPercent) }}</CardTitle></CardHeader>
      <CardContent>
        <LineChart :labels="seriesLabels" :datasets="[{ label: 'CPU %', data: seriesData, color: '#5ec8ff' }]" />
      </CardContent>
    </Card>
    <Card>
      <CardHeader><CardTitle>Per core</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Core</TableHead>
              <TableHead>Usage</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="core in payload?.cores ?? []" :key="core.coreIndex">
              <TableCell>{{ core.coreIndex }}</TableCell>
              <TableCell>{{ formatPercent(core.usagePercent) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
