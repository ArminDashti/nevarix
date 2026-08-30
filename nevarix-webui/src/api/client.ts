function withBase(path: string): string {
  const base = (import.meta.env.BASE_URL || '/').replace(/\/$/, '')
  if (!base || base === '/') return path
  return `${base}${path.startsWith('/') ? path : `/${path}`}`
}

async function getJson<T>(path: string): Promise<T> {
  const res = await fetch(withBase(path))
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  return (await res.json()) as T
}

export const api = {
  dashboard: () => getJson<Record<string, unknown>>('/api/v1/dashboard'),
  cpu: () => getJson<Record<string, unknown>>('/api/v1/cpu'),
  memory: () => getJson<Record<string, unknown>>('/api/v1/memory'),
  disk: () => getJson<Record<string, unknown>>('/api/v1/disk'),
  bandwidth: () => getJson<Record<string, unknown>>('/api/v1/bandwidth'),
  dockerImages: () => getJson<{ images: unknown[]; error?: string }>('/api/v1/docker/images'),
  dockerContainers: () => getJson<{ stacks: unknown[]; error?: string }>('/api/v1/docker/containers'),
  softetherSessions: () => getJson<{ sessions: SoftEtherSession[]; error?: string }>('/api/v1/softether/sessions'),
  softetherUsers: () => getJson<{ users: SoftEtherUser[]; error?: string }>('/api/v1/softether/users'),
  softetherUserSessions: (username: string) =>
    getJson<{ sessions: SoftEtherSessionLog[]; error?: string }>(
      `/api/v1/softether/users/${encodeURIComponent(username)}/sessions`,
    ),
}

export type SoftEtherSession = {
  username: string
  clientIp: string
  asn: string | null
  bandwidthBps: number
  downloadBytes: number
  uploadBytes: number
  sessionDurationSeconds: number
  connectedAt: string | null
}

export type SoftEtherUser = {
  username: string
  clientIp: string
  asn: string | null
  downloadBytes: number
  uploadBytes: number
  usageDurationSeconds: number
}

export type SoftEtherSessionLog = {
  connectedAt: string
  disconnectedAt: string | null
  clientIp: string
  asn: string | null
  downloadBytes: number
  uploadBytes: number
  durationSeconds: number
}
