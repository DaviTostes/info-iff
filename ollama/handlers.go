package ollama

import (
	"encoding/json"
	"iffbot/db"
	"io"
	"net/http"
)

func HandleEmbedding(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileB, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	embedding, err := GetEmbedding(string(fileB))
	if err != nil {
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	jsonEmb, err := json.Marshal(embedding)
	if err != nil {
		http.Error(w, "Failed to jsonify embedding", http.StatusInternalServerError)
		return
	}

	result, err := db.CreateEmbedding(string(fileB), string(jsonEmb))
	if err != nil {
		http.Error(w, "Failed to create embedding", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(result)
}
