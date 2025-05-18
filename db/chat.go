package db

import (
	"errors"

	"gorm.io/gorm"
)

type Chat struct {
	ID         uint  `gorm:"primaryKey"`
	Created    int64 `gorm:"autoCreateTime"`
	TelID      uint
	Messages   []Message
	Embeddings []*Embedding `gorm:"many2many:chats_embeddings"`
}

func CreateOrGetChat(telID uint) (Chat, error) {
	db, err := getDb()
	if err != nil {
		return Chat{}, err
	}

	var chat Chat

	result := db.Preload("Messages").Preload("Embeddings").First(&chat, "tel_id = ?", telID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		chat = Chat{TelID: telID}

		if err := db.Create(&chat).Error; err != nil {
			return Chat{}, err
		}

		if err := db.Preload("Messages").Preload("Embeddings").First(&chat, chat.ID).Error; err != nil {
			return Chat{}, err
		}
	} else if result.Error != nil {
		return Chat{}, result.Error
	}

	return chat, nil
}

func AddChatEmbedding(chatID, embeddingID uint) error {
	db, err := getDb()
	if err != nil {
		return err
	}

	chat := Chat{
		ID: chatID,
	}

	result := db.First(&chat)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	chat.Embeddings = append(chat.Embeddings, &Embedding{ID: embeddingID})

	return db.Save(&chat).Error
}
