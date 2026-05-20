// Package i18n provides internationalization support for the Modbus simulator.
// This package handles localization of user-facing messages using go-i18n.
// It supports English and Chinese languages through embedded locale files.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	// bundle is the i18n bundle that holds all locale data.
	// It is initialized by calling Init or MustInit.
	bundle *i18n.Bundle
)

// Init initializes the i18n bundle with the specified default language.
// The defaultLang parameter should be a language tag like "en" or "zh".
// This function loads all locale files embedded in the locales directory.
// Returns an error if loading fails.
func Init(defaultLang string) error {
	bundle = i18n.NewBundle(language.Make(defaultLang))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := "locales/" + entry.Name()
		data, err := localeFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read locale file %s: %w", path, err)
		}
		_, err = bundle.ParseMessageFileBytes(data, entry.Name())
		if err != nil {
			return fmt.Errorf("failed to parse locale file %s: %w", path, err)
		}
	}

	return nil
}

// T translates a message by its ID with optional template data.
// If the message ID is not found, it returns the message ID itself.
// The templateData parameter can be used for variable substitution
// in messages that contain placeholders like {{.Name}}.
func T(messageID string, templateData map[string]interface{}) string {
	if bundle == nil {
		return messageID
	}

	localizer := i18n.NewLocalizer(bundle, "")

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// MustInit initializes i18n and panics on failure.
// This is a convenience wrapper for use in main functions where
// failure to initialize i18n should be a fatal error.
func MustInit(defaultLang string) {
	if err := Init(defaultLang); err != nil {
		panic(fmt.Sprintf("i18n init failed: %v", err))
	}
}
