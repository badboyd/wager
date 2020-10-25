package config

import (
	"bytes"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Schema of configurations
type Schema struct {
	// Service configuration
	Service struct {
		Port int `json:"port"`
	} `json:"service"`
	// Database configuration
	Database struct {
		Host     string `json:"host"`
		Database string `json:"database"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
}

// Load configuration
func Load() (*Schema, error) {
	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()
	v.SetConfigType("yaml")

	if err := v.ReadConfig(bytes.NewBuffer([]byte(defaultValue))); err != nil {
		return nil, err
	}

	cfg := Schema{}
	err := v.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
		c.TagName = "json"
	})
	return &cfg, err
}
