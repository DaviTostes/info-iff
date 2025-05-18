package db

type Embedding struct {
	ID    uint `gorm:"primaryKey"`
	Text  string
	Emb   string
	Chats []*Chat `gorm:"many2many:chats_embeddings"`
}

func CreateEmbedding(text, emb string) (uint, error) {
	db, err := getDb()
	if err != nil {
		return 0, err
	}

	embedding := Embedding{
		Text: text,
		Emb:  emb,
	}

	result := db.Create(&embedding)
	if result.Error != nil {
		return 0, result.Error
	}

	return embedding.ID, nil
}

func FindEmbeddings() ([]Embedding, error) {
	db, err := getDb()
	if err != nil {
		return nil, err
	}

	var embeddings []Embedding
	result := db.Find(&embeddings)

	return embeddings, result.Error
}
