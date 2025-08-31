package main

import (
	"encoding/json"
	lua "github.com/yuin/gopher-lua"
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

func ProcessPermissions(perms Permissions) byte {
	ret := byte(0)

	if perms.LocalFS {
		ret |= 0b00000001
	}
	if perms.RootFS {
		ret |= 0b00000010
	}
	if perms.Network {
		ret |= 0b00000100
	}
	if perms.Input {
		ret |= 0b00001000
	}
	if perms.Render {
		ret |= 0b00010000
	}
	if perms.Update {
		ret |= 0b00100000
	}
	if perms.PlatformSpecific {
		ret |= 0b01000000
	}
	if perms.Runtime {
		ret |= 0b10000000
	}

	return ret
}

func CreateHeader(meta *Metadata) []byte {
	var header []byte

	if meta.Core {
		header = []byte("SE! CORE.EXT")
	} else {
		header = []byte("SE! EXTENSION")

		header = append(header, append([]byte(meta.Id), 0)...)
	}

	header = append(header, append([]byte(meta.Name), 0)...)
	header = append(header, append([]byte(meta.Description), 0)...)
	header = append(header, ProcessPermissions(meta.Permissions))

	return header
}

func CompileLua(source string) ([]byte, error) {
	L := lua.NewState()
	defer L.Close()

	lfunc, err := L.LoadString(source)
	if err != nil {
		return nil, err
	}

	return DumpLua(lfunc.Proto), nil
}

func main() {
	meta, err := ParseJSON()
	if err != nil {
		log.Fatal(err)
	}

	output := CreateHeader(meta)

	data, err := os.ReadFile("main.lua")
	if err != nil {
		log.Fatal(err)
	}

	luaBytecode, err := CompileLua(string(data))
	if err != nil {
		log.Fatal(err)
	}

	output = append(output, luaBytecode...)

	var fileName string
	if meta.Core {
		fileName = meta.Id + ".sece"
	} else {
		fileName = meta.Id + ".see"
	}

	if err = os.WriteFile(fileName, output, 0644); err != nil {
		log.Fatal(err)
	}
}
