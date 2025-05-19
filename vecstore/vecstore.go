package vecstore

import (
	"encoding/json"
	"iffbot/db"
	"math"
	"sort"
	"strings"
)

type ScoredDoc struct {
	Doc   db.Embedding
	Score float64
}

func cosineSim(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	if len(a) != len(b) {
		return 0
	}

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
	query string,
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

	scoredDocs = heuristicRerank(scoredDocs, query)

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

func heuristicRerank(results []ScoredDoc, query string) []ScoredDoc {
	queryTerms := strings.Fields(strings.ToLower(query))

	for i := range results {
		text := strings.ToLower(results[i].Doc.Text)
		for _, term := range queryTerms {
			if strings.Contains(text, term) {
				results[i].Score += 0.05
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results
}
