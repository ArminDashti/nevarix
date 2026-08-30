<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api, type SoftEtherSessionLog, type SoftEtherUser } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatBytes, formatDuration } from '@/lib/utils'

const users = ref<SoftEtherUser[]>([])
const selected = ref<string | null>(null)
const logs = ref<SoftEtherSessionLog[]>([])
const error = ref('')
let timer: number | undefined

async function load() {
  try {
    const res = await api.softetherUsers()
    users.value = res.users ?? []
    error.value = res.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  }
}

async function openUser(username: string) {
  selected.value = username
  try {
    const res = await api.softetherUserSessions(username)
    logs.value = res.sessions ?? []
    if (res.error) error.value = res.error
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load sessions'
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 10000)
})
onUnmounted(() => {
  if (timer) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-muted-foreground">{{ error }}</p>
    <Card>
      <CardHeader><CardTitle>SoftEther user stats</CardTitle></CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Username</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>ASN</TableHead>
              <TableHead>Download</TableHead>
              <TableHead>Upload</TableHead>
              <TableHead>Time of usage</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="users.length === 0">
              <TableCell colspan="6" class="text-muted-foreground">No user stats yet</TableCell>
            </TableRow>
            <TableRow
              v-for="u in users"
              :key="u.username"
              class="cursor-pointer"
              @click="openUser(u.username)"
            >
              <TableCell>{{ u.username }}</TableCell>
              <TableCell>{{ u.clientIp || '—' }}</TableCell>
              <TableCell>{{ u.asn || '—' }}</TableCell>
              <TableCell>{{ formatBytes(u.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(u.uploadBytes) }}</TableCell>
              <TableCell>{{ formatDuration(u.usageDurationSeconds) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Card v-if="selected">
      <CardHeader>
        <CardTitle>Session logs — {{ selected }}</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Connected at</TableHead>
              <TableHead>Disconnected at</TableHead>
              <TableHead>IP</TableHead>
              <TableHead>ASN</TableHead>
              <TableHead>Download</TableHead>
              <TableHead>Upload</TableHead>
              <TableHead>Time</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="logs.length === 0">
              <TableCell colspan="7" class="text-muted-foreground">No sessions</TableCell>
            </TableRow>
            <TableRow v-for="(log, i) in logs" :key="i">
              <TableCell>{{ new Date(log.connectedAt).toLocaleString() }}</TableCell>
              <TableCell>{{ log.disconnectedAt ? new Date(log.disconnectedAt).toLocaleString() : '—' }}</TableCell>
              <TableCell>{{ log.clientIp || '—' }}</TableCell>
              <TableCell>{{ log.asn || '—' }}</TableCell>
              <TableCell>{{ formatBytes(log.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(log.uploadBytes) }}</TableCell>
              <TableCell>{{ formatDuration(log.durationSeconds) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
