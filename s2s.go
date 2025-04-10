package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type SentryURL struct {
	DSN         string
	HostName    string
	ProjectSlug int
}

func NewSentryURL() SentryURL {
	return SentryURL{
		DSN:         viper.GetString("sentry.dsn"),
		HostName:    viper.GetString("sentry.hostname"),
		ProjectSlug: viper.GetInt("sentry.project"),
	}
}

func (u SentryURL) String() string {
	return fmt.Sprintf("https://%s@%s/api/%d/envelope/",
		u.DSN,
		u.HostName,
		u.ProjectSlug,
	)
}
