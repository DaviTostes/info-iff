package main

import (
	"encoding/json"
	"fmt"
	"iffbot/db"
	"iffbot/models"
	"iffbot/ollama"
	"iffbot/utils"
	"iffbot/webhook"
	"io"
	"log"
	"net/http"
)

func main() {
	if err := db.Migrate(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/bot-info", HandleGetMe)
	http.HandleFunc("/webhook-info", webhook.HandleGetWebhookInfo)
	http.HandleFunc("/webhook-set", webhook.HandleSetWebhook)
	http.HandleFunc("/webhook", webhook.HandleWebhook)
	http.HandleFunc("/embedding", ollama.HandleEmbedding)

	log.Fatal(http.ListenAndServe(":20512", nil))
}

func HandleGetMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	settings, err := utils.GetSettings()
	if err != nil {
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	url := fmt.Sprint(settings.Url, "/getMe")

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

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}
