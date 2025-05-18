package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iffbot/db"
	"io"
	"net/http"
)

type ResponseChat struct {
	Model      string      `json:"model"`
	Created_at string      `json:"created_at"`
	Message    MessageChat `json:"message"`
	Done       bool        `json:"done"`
}

type MessageChat struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func Generate(model, context, userMessage string, messages []db.Message) (string, error) {
	systemMessage := MessageChat{
		Role:    "system",
		Content: GetPremadePrompt(context),
	}

	chatMessages := []MessageChat{systemMessage}
	for _, m := range messages {
		chatMessages = append(chatMessages, MessageChat{
			Role:    m.Sender,
			Content: m.Content,
		})
	}

	chatMessages = append(chatMessages, MessageChat{
		Role:    "user",
		Content: userMessage,
	})

	fmt.Println(chatMessages)

	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"messages": chatMessages,
		"stream":   false,
	})
	payload := bytes.NewBuffer(body)

	req, err := http.Post("http://localhost:11434/api/chat", "application/json", payload)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	if req.StatusCode != 200 {
		return "", errors.New(req.Status)
	}

	body, err = io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	var resp ResponseChat
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}

	return resp.Message.Content, err
}

func GetPremadePrompt(context string) string {
	return fmt.Sprintf(`
Você é InfoIFF, um assistente virtual acolhedor e simpático que responde dúvidas sobre o Instituto Federal Fluminense (IFF) — especialmente o campus Itaperuna.

Fale como se estivesse em um chat no WhatsApp: com leveza, clareza e vontade real de ajudar.

---

**Contexto útil para as respostas:**
%s

---

**Regras de comportamento:**
- Fale de forma natural, como uma pessoa prestativa e educada.
- Seja direto ao ponto, sem enrolar, mas sempre gentil.
- Nao invente informacoes, apenas responda com as que estao no contexto
- Se o contexto for insuficiente para responder a pergunta, tente perguntar para o usuario por mais informacoes
- Se o contexto for extenso, filtre e use apenas o que for relevante para responder.
- Mantenha as respostas ate 3 paragrafos, nao crie textos muito extensos.

Seu objetivo é ser claro, útil e acolhedor — sem parecer um robô e sem exageros.
Mantenha o foco da conversa no IFF, nao va para outros assuntos mesmo se o usuario requisitar.

---
`, context)
}

func GetEmbedding(text string) ([]float64, error) {
	body, err := json.Marshal(map[string]string{
		"model":  "nomic-embed-text",
		"prompt": text,
	})
	if err != nil {
		return nil, err
	}

	payload := bytes.NewBuffer(body)

	req, err := http.Post("http://localhost:11434/api/embeddings", "application/json", payload)
	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	respBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Embedding []float64 `json:"embedding"`
	}
	err = json.Unmarshal(respBody, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Embedding, nil
}
