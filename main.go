package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
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
	Platforms   []string    `json:"platforms"`
}

func ParseJSON(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
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

func ProcessPlatforms(platforms []string) (byte, error) {
	ret := byte(0)

	for _, platform := range platforms {
		if platform == "3ds" {
			ret |= 0b00000001
			continue
		}
		if platform == "wiiu" {
			ret |= 0b00000010
			continue
		}
		if platform == "wii" {
			ret |= 0b00000100
			continue
		}
		if platform == "gamecube" {
			ret |= 0b00001000
			continue
		}
		if platform == "switch" {
			ret |= 0b00010000
			continue
		}
		if platform == "pc" {
			ret |= 0b00100000
			continue
		}
		if platform == "vita" {
			ret |= 0b01000000
			continue
		}
		return 0, fmt.Errorf("Unknown platform: '" + platform + "'")
	}

	return ret, nil
}

func ProcessBlockInfo(source string) (map[string]string, error) {
	reader := strings.NewReader(source)
	stmts, err := parse.Parse(reader, "main.lua")
	if err != nil {
		return nil, err
	}

	blocks := make(map[string]string)

	for _, statement := range stmts {
		var block string

		assignStmt, ok := statement.(*ast.AssignStmt)
		if ok {
			if len(assignStmt.Lhs) != 1 {
				continue
			}

			attrGetExpr, ok := assignStmt.Lhs[0].(*ast.AttrGetExpr)
			if !ok {
				continue
			}

			object, ok := attrGetExpr.Object.(*ast.IdentExpr)
			if !ok || object.Value != "blocks" {
				continue
			}

			key, ok := attrGetExpr.Key.(*ast.StringExpr)
			if !ok {
				continue
			}

			block = key.Value
		} else {
			funcDefStmt, ok := statement.(*ast.FuncDefStmt)
			if ok {
				attrGetExpr, ok := funcDefStmt.Name.Func.(*ast.AttrGetExpr)
				if !ok {
					continue
				}

				object, ok := attrGetExpr.Object.(*ast.IdentExpr)
				if !ok || object.Value != "blocks" {
					continue
				}

				key, ok := attrGetExpr.Key.(*ast.StringExpr)
				if !ok {
					continue
				}

				block = key.Value
			}
		}

		if statement.Line()-1 <= 0 {
			return nil, fmt.Errorf("Invalid Block Function at Line: 1")
		}

		commentLine := strings.Split(source, "\n")[statement.Line()-2]
		if match, _ := regexp.MatchString("^-- type: (command|hat|event|reporter|boolean|bool)$", commentLine); !match {
			return nil, fmt.Errorf("Invalid Block Function at Line: " + strconv.Itoa(statement.Line()))
		}

		blocks[block] = commentLine[9:]
	}

	return blocks, nil
}

func CreateHeader(meta *Metadata, blocks map[string]string) ([]byte, error) {
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

	platforms, err := ProcessPlatforms(meta.Platforms)
	if err != nil {
		return nil, err
	}
	header = append(header, platforms)

	var blockTypes []byte
	var blockIds []byte

	for blockId, blockType := range blocks {
		blockIds = append(blockIds, append([]byte(blockId), 0)...)

		switch blockType {
		case "command":
			blockTypes = append(blockTypes, 0x1)
		case "hat":
			blockTypes = append(blockTypes, 0x2)
		case "event":
			blockTypes = append(blockTypes, 0x3)
		case "reporter":
			blockTypes = append(blockTypes, 0x4)
		case "boolean":
		case "bool":
			blockTypes = append(blockTypes, 0x5)
		}
	}

	header = append(header, append(append(blockTypes, 0), blockIds...)...)

	return header, nil
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
	sourcePath := pflag.StringP("source", "i", "main.lua", "Path to the Lua source code of the extension.")
	outputPath := pflag.StringP("output", "o", "", "Path to put the outputed .see or .sece.")
	metaPath := pflag.StringP("meta", "m", "meta.json", "Path to the JSON file containing metadata.")
	pflag.Parse()

	if pflag.Lookup("source").Changed {
		if _, err := os.Stat(*sourcePath); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Error: '%s' does not exist.\n", *sourcePath)
			pflag.Usage()
			os.Exit(1)
		} else if err != nil {
			panic(err)
		}
	}

	if pflag.Lookup("meta").Changed {
		if _, err := os.Stat(*metaPath); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Error: '%s' does not exist.\n", *metaPath)
			pflag.Usage()
			os.Exit(1)
		} else if err != nil {
			panic(err)
		}
	}

	meta, err := ParseJSON(*metaPath)
	if err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile(*sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	blocks, err := ProcessBlockInfo(string(data))
	if err != nil {
		log.Fatal(err)
	}

	output, err := CreateHeader(meta, blocks)
	if err != nil {
		log.Fatal(err)
	}

	luaBytecode, err := CompileLua(string(data))
	if err != nil {
		log.Fatal(err)
	}

	output = append(output, luaBytecode...)

	var fileName string
	if pflag.Lookup("output").Changed {
		fileName = *outputPath
	} else {
		fileName = meta.Id
		if meta.Core {
			fileName += ".sece"
		} else {
			fileName += ".see"
		}
	}

	if err = os.WriteFile(fileName, output, 0644); err != nil {
		log.Fatal(err)
	}
}
