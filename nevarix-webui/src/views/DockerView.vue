<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatBytes, formatDuration, formatPercent } from '@/lib/utils'

const images = ref<any[]>([])
const stacks = ref<any[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const [imgRes, cRes] = await Promise.all([api.dockerImages(), api.dockerContainers()])
    images.value = imgRes.images ?? []
    stacks.value = cRes.stacks ?? []
    error.value = imgRes.error || cRes.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 8000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-6">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>Images</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>Tags</TableHead>
              <TableHead>Size</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="img in images" :key="img.id">
              <TableCell class="font-mono text-xs">{{ img.id }}</TableCell>
              <TableCell>{{ (img.tags || []).join(', ') }}</TableCell>
              <TableCell>{{ formatBytes(img.sizeBytes) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Card v-for="stack in stacks" :key="stack.stackName">
      <CardHeader><CardTitle>Stack: {{ stack.stackName }}</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Container</TableHead>
              <TableHead>Image</TableHead>
              <TableHead>Uptime</TableHead>
              <TableHead>RAM</TableHead>
              <TableHead>Disk</TableHead>
              <TableHead>CPU</TableHead>
              <TableHead>Internal</TableHead>
              <TableHead>Reverse proxy</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="c in stack.containers" :key="c.containerName">
              <TableCell>{{ c.containerName }}</TableCell>
              <TableCell class="max-w-[180px] truncate">{{ c.image }}</TableCell>
              <TableCell>{{ formatDuration(c.uptimeSeconds) }}</TableCell>
              <TableCell>{{ formatBytes(c.memoryBytes) }}</TableCell>
              <TableCell>{{ formatBytes(c.diskBytes) }}</TableCell>
              <TableCell>{{ formatPercent(c.cpuPercent) }}</TableCell>
              <TableCell class="font-mono text-xs">{{ c.internalEndpoint || '—' }}</TableCell>
              <TableCell>{{ c.reverseProxyRoute || '—' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
