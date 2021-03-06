package config

import (
	"errors"
	"strings"
	"time"
)

const PrivateKey = "2e2337a4b6870c09ddb7ef926c4dc382765beed14ace12066ca5de7490cf8e4828f9ee6111b2c0e58c13481ae674ef6e12ba2df1a015442e474c7a63fcd06dc6"

type Tags struct {
	Values []string
}

func (tags *Tags) Set(value string) error {
	tags.Values = strings.Split(value, ",")

	for i, tag := range tags.Values {
		tag = strings.TrimSpace(tag)

		if len(tag) == 0 {
			return errors.New("incorrect NMEA tags")
		}

		tag = strings.ToUpper(tag)
		tags.Values[i] = tag
	}

	return nil
}

func (tags *Tags) String() string {
	return strings.Join(tags.Values, ",")
}

func ParseTags(value string) Tags {
	tags := &Tags{}
	tags.Set(value)
	return *tags
}

type Sd struct {
	// sd card settings
	SCid          string
	SdPart        int
	SdSys         string
	SdDev         string
	DirMedia      string
	DirZeroConfig string
	Build         int64

	// flush cache settings
	FileWatch      string
	TimeoutChanges time.Duration

	// alive checker settings
	FileLive       string
	FileIsAlive    string
	FileWallpaper  string
	DirLogs        string
	TimeoutIsAlive time.Duration

	// csv handler
	DataPort    string
	CaptureTags Tags
}
