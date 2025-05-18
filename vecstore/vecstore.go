package vecstore

import (
	"encoding/json"
	"iffbot/db"
	"math"
	"sort"
)

type ScoredDoc struct {
	Doc   db.Embedding
	Score float64
}

func cosineSim(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func FindTopKRelevant(
	docs []db.Embedding,
	queryEmbedding []float64,
	topK int,
	minScore float64,
) []db.Embedding {
	var scoredDocs []ScoredDoc

	for _, doc := range docs {
		var emb []float64
		err := json.Unmarshal([]byte(doc.Emb), &emb)
		if err != nil {
			continue
		}

		score := cosineSim(queryEmbedding, emb)
		if score >= minScore {
			scoredDocs = append(scoredDocs, ScoredDoc{Doc: doc, Score: score})
		}
	}

	if len(scoredDocs) < 1 {
		return []db.Embedding{}
	}

	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].Score > scoredDocs[j].Score
	})

	if topK > 0 && len(scoredDocs) > topK {
		scoredDocs = scoredDocs[:topK]
	}

	result := make([]db.Embedding, len(scoredDocs))
	for i, scored := range scoredDocs {
		result[i] = scored.Doc
	}

	return result
}
