<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api, type SoftEtherSession } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatBytes, formatBps, formatDuration } from '@/lib/utils'

const sessions = ref<SoftEtherSession[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await api.softetherSessions()
    sessions.value = res.sessions ?? []
    error.value = res.error || ''
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
    <p v-if="error" class="text-sm text-muted-foreground">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>SoftEther online users</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>ASN</TableHead>
              <TableHead>BW</TableHead>
              <TableHead>Download</TableHead>
              <TableHead>Upload</TableHead>
              <TableHead>Time</TableHead>
              <TableHead>Connected at</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="sessions.length === 0">
              <TableCell colspan="8" class="text-muted-foreground">No online sessions</TableCell>
            </TableRow>
            <TableRow v-for="(s, i) in sessions" :key="i">
              <TableCell>{{ s.username }}</TableCell>
              <TableCell>{{ s.clientIp || '—' }}</TableCell>
              <TableCell>{{ s.asn || '—' }}</TableCell>
              <TableCell>{{ formatBps(s.bandwidthBps) }}</TableCell>
              <TableCell>{{ formatBytes(s.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(s.uploadBytes) }}</TableCell>
              <TableCell>{{ formatDuration(s.sessionDurationSeconds) }}</TableCell>
              <TableCell>{{ s.connectedAt ? new Date(s.connectedAt).toLocaleString() : '—' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
