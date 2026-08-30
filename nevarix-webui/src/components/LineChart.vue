<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const props = defineProps<{
  labels: string[]
  datasets: { label: string; data: number[]; color?: string }[]
}>()

const chartData = computed(() => ({
  labels: props.labels,
  datasets: props.datasets.map((d) => ({
    label: d.label,
    data: d.data,
    borderColor: d.color ?? '#5ec8ff',
    backgroundColor: (d.color ?? '#5ec8ff') + '33',
    fill: true,
    tension: 0.35,
    pointRadius: 0,
    borderWidth: 2,
  })),
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { labels: { color: '#c5d0e0' } },
  },
  scales: {
    x: { ticks: { color: '#8b97a8', maxTicksLimit: 8 }, grid: { color: '#2a3344' } },
    y: { ticks: { color: '#8b97a8' }, grid: { color: '#2a3344' } },
  },
}
</script>

<template>
  <div class="h-64 w-full">
    <Line :data="chartData" :options="options" />
  </div>
</template>
