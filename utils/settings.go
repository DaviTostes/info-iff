package utils

import (
	"encoding/json"
	"os"
)

type Settings struct {
	Bot_token   string `json:"bot_token"`
	Url         string `json:"url"`
	Model       string `json:"model"`
	Db          string `json:"db"`
}

func GetSettings() (Settings, error) {
	file, err := os.ReadFile("settings.json")
	if err != nil {
		return Settings{}, err
	}

	var settings Settings
	err = json.Unmarshal(file, &settings)
	if err != nil {
		return Settings{}, err
	}

	return settings, nil
}
