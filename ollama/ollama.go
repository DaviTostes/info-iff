package ollama

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"iffbot/db"
	"io"
	"net/http"
	"regexp"
)

type Response struct {
	Model      string `json:"model"`
	Created_at string `json:"created_at"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
}

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

func Generate(model, userPrompt string, messages []db.Message) (string, error) {
	systemMessage := MessageChat{
		Role:    "system",
		Content: GetPremadePrompt(),
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
		Content: userPrompt,
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

	re := regexp.MustCompile(`(?s)<think>.*?</think>`)
	output := re.ReplaceAllString(resp.Message.Content, "")

	return output, err
}

func GetPremadePrompt() string {
	return fmt.Sprintf(`
você é um assistente que responde dúvidas sobre o IFF campus Itaperuna. 
você recebera um contexto e uma pergunta do usuario.
responda apenas com base neste contexto, nao invente informacoes.
`)
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
