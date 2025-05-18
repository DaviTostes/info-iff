package models

type Result struct {
	Ok          bool           `json:"ok"`
	Description string         `json:"description"`
	Result      map[string]any `json:"result"`
}

type Error struct {
	Description string `json:"error"`
}
