import { onUnmounted, ref } from 'vue'
import { api } from '@/services/api'
import type { TelemetryReading } from '@/types'

export interface HistoryPoint {
  t: number
  v: number
}

const MAX_HISTORY = 30
const POLL_MS = 2000

/**
 * Polls /api/machines/:id/telemetry on an interval and keeps a rolling
 * history per numeric metric for line-chart widgets.
 */
export function useTelemetry() {
  const reading = ref<TelemetryReading['readings'] | null>(null)
  const history = ref<Record<string, HistoryPoint[]>>({})
  let timer: ReturnType<typeof setInterval> | null = null
  let currentMachineId: string | null = null

  async function tick() {
    const id = currentMachineId
    if (!id) return
    try {
      const data = await api.getTelemetry(id)
      if (id !== currentMachineId) return // stale response after switching machines
      reading.value = data.readings
      const t = Date.now()
      for (const [key, value] of Object.entries(data.readings)) {
        if (typeof value !== 'number') continue
        const arr = history.value[key] ?? (history.value[key] = [])
        arr.push({ t, v: value })
        if (arr.length > MAX_HISTORY) arr.shift()
      }
    } catch {
      // ignore transient errors; keep last known reading
    }
  }

  function start(machineId: string) {
    stop()
    currentMachineId = machineId
    reading.value = null
    history.value = {}
    void tick()
    timer = setInterval(tick, POLL_MS)
  }

  function stop() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    currentMachineId = null
  }

  onUnmounted(stop)

  return { reading, history, start, stop }
}
