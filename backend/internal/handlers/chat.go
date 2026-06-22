package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ChatHandler forwards chat messages to the Groq API (OpenAI-compatible) using
// a Qwen model. It is stateless: the frontend sends the full conversation
// history plus a compact context string describing the dashboard on screen.
type ChatHandler struct {
	client *http.Client
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{client: &http.Client{Timeout: 30 * time.Second}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Messages []chatMessage `json:"messages"`
	Context  string        `json:"context"`
}

const chatSystemPrompt = "You are a helpful assistant for factory operators monitoring a live dashboard. " +
	"Answer conversationally and warmly — plain language by default, but match the technical depth of the question. " +
	"For simple questions be brief (1–2 sentences). When asked to explain or compare, give a clear structured answer. " +
	"Always cite specific widgets as [N: type 'title'] (e.g. [0: status 'Machine Status'], [2: gauge 'Temperature']) when your answer refers to them. " +
	"When referring to a specific part of a widget, cite it as [N.key] (e.g. [0.bar], [2.title]) — each widget lists its available element keys in the context."

// buildGroqBody assembles the OpenAI-compatible request payload. The system
// instruction + dashboard context are prepended as a system message. Split out
// so the assembly can be checked without a network call.
func chatTemperature() float64 {
	if v, err := strconv.ParseFloat(os.Getenv("CHAT_TEMPERATURE"), 64); err == nil {
		return v
	}
	return 0.4
}

func buildGroqBody(model string, req chatRequest) map[string]any {
	system := chatSystemPrompt
	if override := os.Getenv("CHAT_SYSTEM_PROMPT"); override != "" {
		system = override
	}
	if strings.TrimSpace(req.Context) != "" {
		system += "\n\nDashboard context:\n" + req.Context
	}
	messages := make([]chatMessage, 0, len(req.Messages)+1)
	messages = append(messages, chatMessage{Role: "system", Content: system})
	messages = append(messages, req.Messages...)
	return map[string]any{
		"model":            model,
		"messages":         messages,
		"temperature":      chatTemperature(),
		// Qwen3 is a reasoning model; keep its <think> trace out of the reply.
		"reasoning_format": "hidden",
	}
}

func (h *ChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		writeError(w, http.StatusServiceUnavailable, "chat is not configured (GROQ_API_KEY unset)")
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "messages required")
		return
	}

	model := os.Getenv("GROQ_MODEL")
	if model == "" {
		model = "qwen/qwen3-32b"
	}

	body, _ := json.Marshal(buildGroqBody(model, req))
	apiReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost,
		"https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build request")
		return
	}
	apiReq.Header.Set("content-type", "application/json")
	apiReq.Header.Set("authorization", "Bearer "+apiKey)

	resp, err := h.client.Do(apiReq)
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to reach chat service")
		return
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		writeError(w, http.StatusBadGateway, "chat service error: "+string(raw))
		return
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil || len(parsed.Choices) == 0 {
		writeError(w, http.StatusBadGateway, "invalid chat service response")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"reply": parsed.Choices[0].Message.Content})
}
