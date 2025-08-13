// Copyright (c) 2024 Miguel Mamani
// Email: miguel.coder.per@gmail.com
// License: MIT

package entities

import (
	"time"

	"github.com/google/uuid"
)

type HLSPlaylist struct {
	MasterPlaylist string       `json:"master_playlist"`
	Variants       []HLSVariant `json:"variants"`
	BaseURL        string       `json:"base_url"`
}

type HLSVariant struct {
	Quality     string `json:"quality"`
	Bandwidth   int    `json:"bandwidth"`
	Resolution  string `json:"resolution"`
	PlaylistURL string `json:"playlist_url"`
}

type StreamingSession struct {
	UserID    uuid.UUID `json:"user_id"`
	VideoID   uuid.UUID `json:"video_id"`
	Quality   string    `json:"quality"`
	Position  int       `json:"position"` // en segundos
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
