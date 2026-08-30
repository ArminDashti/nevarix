<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatBytes, formatPercent } from '@/lib/utils'

const payload = ref<any>(null)
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    payload.value = await api.disk()
    error.value = ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 5000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>Disk volumes</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Mount</TableHead>
              <TableHead>FS</TableHead>
              <TableHead>Used</TableHead>
              <TableHead>Free</TableHead>
              <TableHead>Total</TableHead>
              <TableHead>%</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="disk in payload?.disks ?? []" :key="disk.mountPoint">
              <TableCell>{{ disk.mountPoint }}</TableCell>
              <TableCell>{{ disk.fstype || '—' }}</TableCell>
              <TableCell>{{ formatBytes(disk.usedBytes) }}</TableCell>
              <TableCell>{{ formatBytes(disk.freeBytes) }}</TableCell>
              <TableCell>{{ formatBytes(disk.totalBytes) }}</TableCell>
              <TableCell>{{ formatPercent(disk.usedPercent) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
