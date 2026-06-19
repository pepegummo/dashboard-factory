# AI Chat Improvements — Design Spec

**Date:** 2026-06-19
**Status:** Approved

## Goal

Make the in-app AI chat (ExplorePanel) feel warmer and more useful by improving
its tone and giving it memory of the current conversation session.

## Scope

Two files changed, ~10 lines of net new code:

1. `backend/internal/handlers/chat.go` — rewrite `chatSystemPrompt`
2. `frontend/src/components/ExplorePanel.vue` — add invisible conversation history

## Design

### 1. System Prompt Rewrite (`chat.go`)

Replace the current `chatSystemPrompt` constant with a tone that is:

- **Warm and conversational** — not stiff or purely technical
- **Adaptive** — 1–2 sentences for simple questions; structured answer when
  asked to explain or compare
- **Plain-language by default** — matches technical depth to the question
- **Widget-aware** — still cites widgets as `[N]` when referring to them

New prompt (exact wording tunable, this is the intent):

> "You are a helpful assistant for factory operators monitoring a live
> dashboard. Answer conversationally and warmly — plain language by default,
> but match the technical depth of the question. For simple questions be brief
> (1–2 sentences). When asked to explain or compare, give a clear structured
> answer. Always cite specific widgets as [N] when your answer refers to them."

The dashboard context block appended below the system prompt is unchanged.

### 2. Invisible Conversation History (`ExplorePanel.vue`)

**Problem:** every call to `api.chat()` currently sends only the single latest
message, so the AI has no memory of previous turns.

**Solution:** accumulate turns in a local `messages` array and send the full
array on every call. The UI stays the same — only the latest reply is shown in
`answer`.

#### State changes

| Before | After |
|--------|-------|
| `answer` ref (last reply) | unchanged |
| `question` ref (current input) | unchanged |
| sends `[{ role:'user', content:q }]` | sends full `messages` array |
| — | `messages` ref: growing `{role,content}[]` |

#### Submit flow

1. Push `{ role: 'user', content: q }` onto `messages`
2. Call `api.chat(messages.value, props.context)`
3. Push `{ role: 'assistant', content: reply }` onto `messages`
4. Set `answer.value = reply` (unchanged — what the UI renders)

#### Reset trigger

Watch `props.context`. When it changes (user switches dashboard or machine
page), reset both `messages.value = []` and `answer.value = ''`. This prevents
the AI from referencing a previous machine's data after navigation.

#### Why invisible history (not a visible thread)

- UI redesign was out of scope — user confirmed option B
- The AI still benefits from context ("what about the pressure?" after asking
  about temperature resolves correctly)
- Keeps the component small; a visible thread can be added later if needed

## What Is Not In Scope

- Machine `type` / `status` fields in the context string (user said context is fine as-is)
- Visible conversation thread / chat bubbles
- Backend session storage
- System prompt in an external `.md` file (user confirmed prompt rarely changes)
- Alert threshold interpretation

## Files Changed

| File | Change |
|------|--------|
| `backend/internal/handlers/chat.go` | Rewrite `chatSystemPrompt` constant |
| `frontend/src/components/ExplorePanel.vue` | Add `messages` ref, update `submit()`, add `watch(props.context)` reset |
