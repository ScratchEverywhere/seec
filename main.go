package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Permissions struct {
	LocalFS          bool `json:"local-fs"`
	RootFS           bool `json:"root-fs"`
	Network          bool `json:"network"`
	Input            bool `json:"input"`
	Render           bool `json:"render"`
	Update           bool `json:"update"`
	PlatformSpecific bool `json:"platform-specific"`
	Runtime          bool `json:"runtime"`
}

type Metadata struct {
	Core        bool        `json:"core"`
	Id          string      `json:"id,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Permissions Permissions `json:"permissions"`
}

func ParseJSON() (*Metadata, error) {
	data, err := os.ReadFile("meta.json")
	if err != nil {
		return nil, err
	}

	var meta Metadata
	if err = json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func main() {
	meta, err := ParseJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", *meta)
}
