import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '@/views/DashboardView.vue'
import CpuView from '@/views/CpuView.vue'
import MemoryView from '@/views/MemoryView.vue'
import DiskView from '@/views/DiskView.vue'
import BandwidthView from '@/views/BandwidthView.vue'
import DockerView from '@/views/DockerView.vue'
import SoftEtherOnlineView from '@/views/SoftEtherOnlineView.vue'
import SoftEtherUsersView from '@/views/SoftEtherUsersView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'dashboard', component: DashboardView },
    { path: '/cpu', name: 'cpu', component: CpuView },
    { path: '/memory', name: 'memory', component: MemoryView },
    { path: '/disk', name: 'disk', component: DiskView },
    { path: '/bandwidth', name: 'bandwidth', component: BandwidthView },
    { path: '/docker', name: 'docker', component: DockerView },
    { path: '/softether/online-users', name: 'softether-online', component: SoftEtherOnlineView },
    { path: '/softether/users-stat', name: 'softether-users', component: SoftEtherUsersView },
  ],
})

export default router
