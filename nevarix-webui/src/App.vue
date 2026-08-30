<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/cpu', label: 'CPU' },
  { to: '/memory', label: 'Memory' },
  { to: '/disk', label: 'Disk' },
  { to: '/bandwidth', label: 'Bandwidth' },
  { to: '/docker', label: 'Docker' },
  { to: '/softether/online-users', label: 'SE Online' },
  { to: '/softether/users-stat', label: 'SE Users' },
]
</script>

<template>
  <div class="mx-auto flex min-h-screen max-w-7xl flex-col px-4 py-6">
    <header class="mb-8 flex flex-wrap items-end justify-between gap-4 border-b border-border pb-4">
      <div>
        <p class="text-xs font-medium uppercase tracking-[0.25em] text-primary">Nevarix</p>
        <h1 class="mt-1 text-2xl font-semibold tracking-tight">Host monitor</h1>
      </div>
      <nav class="flex flex-wrap gap-1">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          active-class="!bg-primary !text-primary-foreground"
        >
          {{ link.label }}
        </RouterLink>
      </nav>
    </header>
    <main class="flex-1">
      <RouterView />
    </main>
  </div>
</template>
