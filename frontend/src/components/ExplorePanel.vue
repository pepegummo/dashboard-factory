<script setup lang="ts">
import { ref, watch } from 'vue'
import type { ChatMessage } from '@/types'
import { api, apiErrorMessage } from '@/services/api'

const props = defineProps<{ context: string; resetKey?: string; prefill?: string }>()
const emit = defineEmits<{
  highlight: [indices: number[]]
  focus: [el: { widgetIndex: number; key: string } | null]
}>()

const question = ref('')
const messages = ref<ChatMessage[]>([])

watch(
  () => props.prefill,
  (v) => { if (v) question.value = v },
)
const answer = ref('')
const loading = ref(false)
const errorMsg = ref('')

watch(
  () => props.resetKey,
  () => { messages.value = []; answer.value = ''; errorMsg.value = '' },
)

function clear() {
  messages.value = []
  answer.value = ''
  errorMsg.value = ''
  emit('highlight', [])
  emit('focus', null)
}

async function submit() {
  const q = question.value.trim()
  if (!q || loading.value) return
  loading.value = true
  errorMsg.value = ''
  question.value = ''
  messages.value.push({ role: 'user', content: q })
  try {
    const { reply } = await api.chat(messages.value.slice(-20), props.context)
    messages.value.push({ role: 'assistant', content: reply })
    answer.value = reply
    const indices = [...reply.matchAll(/\[(\d+)[^\]]*\]/g)].map((m) => parseInt(m[1]))
    emit('highlight', [...new Set(indices)])
    const elMatch = reply.match(/\[(\d+)\.(\w+)\]/)
    emit('focus', elMatch ? { widgetIndex: parseInt(elMatch[1]), key: elMatch[2] } : null)
  } catch (err) {
    messages.value.pop() // remove the user message that failed
    errorMsg.value = apiErrorMessage(err, 'Failed to get answer')
  } finally {
    loading.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    submit()
  }
}
</script>

<template>
  <div class="explore-panel">
    <div class="explore-answer">
      <p v-if="!answer && !errorMsg && !loading" class="explore-placeholder">
        Ask a question about this dashboard to get started.
      </p>
      <div v-else-if="loading" class="text-center py-4">
        <div class="spinner-border spinner-border-sm text-info" role="status"></div>
      </div>
      <p v-else-if="errorMsg" class="text-danger small">{{ errorMsg }}</p>
      <p v-else class="explore-text">{{ answer }}</p>
    </div>

    <div class="explore-input-row">
      <button v-if="messages.length" class="explore-clear-btn" title="Clear chat" @click="clear">
        <i class="bi bi-trash"></i>
      </button>
      <textarea
        v-model="question"
        class="explore-textarea"
        placeholder="Ask about the dashboard…"
        rows="2"
        :disabled="loading"
        @keydown="onKeydown"
      ></textarea>
      <button class="explore-send-btn" :disabled="!question.trim() || loading" @click="submit">
        <i class="bi bi-send-fill"></i>
      </button>
    </div>
  </div>
</template>

<style scoped>
.explore-panel {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--db-card-bg);
  border: 1px solid var(--db-border);
  border-radius: 0.5rem;
}

.explore-answer {
  max-height: 45vh;
  overflow-y: auto;
  padding: 1rem;
}

.explore-placeholder {
  color: var(--db-text-muted);
  font-size: 0.85rem;
  margin: 0;
}

.explore-text {
  color: var(--db-text);
  font-size: 0.875rem;
  line-height: 1.6;
  margin: 0;
  white-space: pre-wrap;
}

.explore-input-row {
  display: flex;
  gap: 0.5rem;
  align-items: flex-end;
  padding: 0.75rem;
  border-top: 1px solid var(--db-border);
}

.explore-textarea {
  flex: 1;
  resize: none;
  background: var(--db-bg);
  border: 1px solid var(--db-border);
  border-radius: 0.375rem;
  color: var(--db-text);
  font-size: 0.875rem;
  padding: 0.5rem 0.625rem;
  outline: none;
  transition: border-color 0.15s;
}

.explore-textarea:focus {
  border-color: var(--db-accent);
}

.explore-textarea::placeholder {
  color: var(--db-text-muted);
}

.explore-send-btn {
  background: var(--db-accent);
  border: none;
  border-radius: 0.375rem;
  color: #000;
  padding: 0.5rem 0.75rem;
  font-size: 1rem;
  cursor: pointer;
  flex-shrink: 0;
  transition: opacity 0.15s;
}

.explore-send-btn:disabled {
  opacity: 0.4;
  cursor: default;
}

.explore-clear-btn {
  background: none;
  border: 1px solid var(--db-border);
  border-radius: 0.375rem;
  color: var(--db-text-muted);
  padding: 0.5rem 0.625rem;
  font-size: 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
  transition: color 0.15s;
}

.explore-clear-btn:hover {
  color: var(--bs-danger);
}
</style>
