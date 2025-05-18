package db

type Message struct {
	ID      uint  `gorm:"primaryKey"`
	Created int64 `gorm:"autoCreateTime"`
	Sender  string
	Content string
	ChatID  uint
}

func CreateMessage(chatID uint, content, sender string) (uint, error) {
	db, err := getDb()
	if err != nil {
		return 0, err
	}

	message := Message{
		ChatID:  chatID,
		Content: content,
		Sender:  sender,
	}

	result := db.Create(&message)
	if result.Error != nil {
		return 0, result.Error
	}

	return message.ID, nil
}
