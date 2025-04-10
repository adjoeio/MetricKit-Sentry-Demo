package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type FullDSN struct {
	DSN         string
	HostName    string
	ProjectSlug int
}

func NewFullDSN() FullDSN {
	return FullDSN{
		DSN:         viper.GetString("sentry.dsn"),
		HostName:    viper.GetString("sentry.hostname"),
		ProjectSlug: viper.GetInt("sentry.project"),
	}
}

func (d FullDSN) String() string {
	return fmt.Sprintf("https://%s@%s/%d",
		d.DSN,
		d.HostName,
		d.ProjectSlug,
	)
}

func (d FullDSN) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

type EnvelopeHeader struct {
	EventID uuid.UUID
	DSN     FullDSN
	SentAt  time.Time
}

func NewEnvelopeHeader() EnvelopeHeader {
	return EnvelopeHeader{
		EventID: uuid.New(),
		DSN:     NewFullDSN(),
		SentAt:  time.Now().UTC(),
	}
}

func (h EnvelopeHeader) MarshalJSON() ([]byte, error) {
	m := struct {
		EventID string  `json:"event_id"`
		DSN     FullDSN `json:"dsn"`
		SentAt  string  `json:"sent_at"`
	}{
		EventID: strings.ReplaceAll(h.EventID.String(), "-", ""),
		DSN:     h.DSN,
		SentAt:  h.SentAt.UTC().Format(time.RFC3339),
	}
	return json.Marshal(&m)
}

type ItemHeader struct {
	Type        string `json:"type"`
	ContentType string `json:"content_type"`
	Length      int    `json:"length"`
}

func NewItemHeader(length int) ItemHeader {
	return ItemHeader{
		Type:        "event",
		ContentType: "application/json",
		Length:      length,
	}
}
