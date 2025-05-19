package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iffbot/db"
	"iffbot/models"
	"iffbot/ollama"
	"iffbot/utils"
	"iffbot/vecstore"
	"io"
	"net/http"
	"strings"
)

func HandleGetWebhookInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	settings, err := utils.GetSettings()
	if err != nil {
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprint(settings.Url, "/getWebhookInfo")

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to make request", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	var result models.Result
	err = json.Unmarshal(body, &result)
	if err != nil {
		http.Error(w, "Failed to read result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if resp.StatusCode != 200 {
		http.Error(w, "", resp.StatusCode)
		json.NewEncoder(w).Encode(models.Error{Description: result.Description})
		return
	}

	json.NewEncoder(w).Encode(result)
}

func HandleSetWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	settings, err := utils.GetSettings()
	if err != nil {
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprint(settings.Url, "/setWebhook")

	bodyReq, _ := json.Marshal(map[string]string{
		"url": "https://bot.mediumblue.space/webhook",
	})

	payload := bytes.NewBuffer(bodyReq)

	resp, err := http.Post(url, "application/json", payload)
	if err != nil {
		http.Error(w, "Failed to make request", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != 200 {
		http.Error(w, "", resp.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	result := models.Result{
		Ok: true,
	}

	json.NewEncoder(w).Encode(result)
}

type WebhookData struct {
	Message struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID           int64  `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int64  `json:"date"`
		Text string `json:"text"`
	} `json:"message"`
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	settings, err := utils.GetSettings()
	if err != nil {
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	var webhookData WebhookData
	err = json.Unmarshal(body, &webhookData)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	chat, err := db.CreateOrGetChat(uint(webhookData.Message.Chat.ID))
	if err != nil {
		http.Error(w, "Faile to create chat", http.StatusInternalServerError)
		return
	}

	embeddings, err := db.FindEmbeddings()
	if err != nil {
		http.Error(w, "Faile to get embeddings", http.StatusInternalServerError)
		return
	}

	var diferEmbeddings []db.Embedding
	for _, emb := range embeddings {
		inChat := false

		for _, chatEmb := range chat.Embeddings {
			if emb.ID == chatEmb.ID {
				inChat = true
			}
		}

		if !inChat {
			diferEmbeddings = append(diferEmbeddings, emb)
		}
	}

	userEmb, err := ollama.GetEmbedding(webhookData.Message.Text)
	if err != nil {
		return
	}

	topDocs := vecstore.FindTopKRelevant(
		diferEmbeddings,
		webhookData.Message.Text,
		userEmb,
		6,
		0.70,
	)

	var context strings.Builder
	for _, d := range topDocs {
		context.WriteString(d.Text + "\n")
		db.AddChatEmbedding(chat.ID, d.ID)
	}
	if err != nil {
		http.Error(w, "Failed to generate completion", http.StatusInternalServerError)
		return
	}

	userPrompt := fmt.Sprintf(
		"Contexto Inicio\n%sContexto Fim\nPergunta\n%s",
		context.String(),
		webhookData.Message.Text,
	)

	completion, err := ollama.Generate(settings.Model, userPrompt, chat.Messages)
	if err != nil {
		http.Error(w, "Failed to generate completion", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprint(settings.Url, "/sendMessage")

	bodyReq, _ := json.Marshal(map[string]any{
		"chat_id": webhookData.Message.Chat.ID,
		"text":    completion,
	})

	payload := bytes.NewBuffer(bodyReq)

	_, err = http.Post(url, "application/json", payload)
	if err != nil {
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	db.CreateMessage(chat.ID, userPrompt, "user")
	db.CreateMessage(chat.ID, completion, "assistant")

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(models.Result{Ok: true})
}
