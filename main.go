package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
)

type Setting struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Default any     `json:"default"`
	Min     float32 `json:"min,omitempty"`
	Max     float32 `json:"max,omitempty"`
	Snap    float32 `json:"snap,omitempty"`
	Prompt  string  `json:"prompt,omitempty"`
}

type Metadata struct {
	Core        bool      `json:"core"`
	Id          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	Platforms   []string  `json:"platforms"`
	MinAPI      string    `json:"minAPI,omitempty"`
	Settings    []Setting `json:"settings,omitempty"`
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
	if meta.MinAPI == "" {
		meta.MinAPI = "0.0" // Update this to whatever the newest API version is.
	} else if match, _ := regexp.MatchString("^\\d+\\.\\d+$", meta.MinAPI); !match {
		fmt.Println("[WARNING] Invalid API Version: '" + meta.MinAPI + "', using '0.0'.")
		meta.MinAPI = "0.0"
	}
	return &meta, nil
}

func ProcessPermissions(perms []string) ([]byte, error) {
	ret := uint16(0)

	for _, perm := range perms {
		if perm == "localfs" {
			ret |= 0b0000000001
			continue
		}
		if perm == "rootfs" {
			ret |= 0b0000000010
			continue
		}
		if perm == "network" {
			ret |= 0b0000000100
			continue
		}
		if perm == "input" {
			ret |= 0b0000001000
			continue
		}
		if perm == "render" {
			ret |= 0b0000010000
			continue
		}
		if perm == "update" {
			ret |= 0b0000100000
			continue
		}
		if perm == "platform-specific" {
			ret |= 0b0001000000
			continue
		}
		if perm == "runtime" {
			ret |= 0b0010000000
			continue
		}
		if perm == "audio" {
			ret |= 0b0100000000
			continue
		}
		if perm == "extensions" {
			ret |= 0b1000000000
			continue
		}
		return nil, fmt.Errorf("Unknown platform: '" + perm + "'")
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, ret); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
	header := []byte("SE! EXTENSION")

	if meta.Core {
		header = append(header, 1)
	} else {
		header = append(header, 0)
	}

	header = append(header, 0) // Update as file format version changes
	majorVersion, err := strconv.Atoi(strings.Split(meta.MinAPI, ".")[0])
	if err != nil {
		return nil, err
	}
	header = append(header, byte(majorVersion))
	minorVersion, err := strconv.Atoi(strings.Split(meta.MinAPI, ".")[1])
	if err != nil {
		return nil, err
	}
	header = append(header, byte(minorVersion))

	header = append(header, append([]byte(meta.Id), 0)...)
	header = append(header, append([]byte(meta.Name), 0)...)
	header = append(header, append([]byte(meta.Description), 0)...)

	perms, err := ProcessPermissions(meta.Permissions)
	if err != nil {
		return nil, err
	}
	header = append(header, perms...)

	platforms, err := ProcessPlatforms(meta.Platforms)
	if err != nil {
		return nil, err
	}
	header = append(header, platforms)

	for _, setting := range meta.Settings {
		var defaultValue []byte = nil
		var data []byte = nil

		switch setting.Type {
		case "text":
			header = append(header, 0x63)
			data = append([]byte(setting.Prompt), 0)
			if _, ok := setting.Default.(string); ok {
				header = append(header, append([]byte(setting.Default.(string)), 0)...)
			}
			return nil, fmt.Errorf("Invalid default value for text")
		case "slider":
			header = append(header, 0x6e)
			data = make([]byte, 12)
			binary.BigEndian.PutUint32(data, math.Float32bits(float32(setting.Min)))
			binary.BigEndian.PutUint32(data, math.Float32bits(float32(setting.Max)))
			binary.BigEndian.PutUint32(data, math.Float32bits(float32(setting.Snap)))
			if _, ok := setting.Default.(float64); ok {
				defaultValue = make([]byte, 4)
				binary.BigEndian.PutUint32(defaultValue, math.Float32bits(float32(setting.Default.(float64))))
				break
			}
			return nil, fmt.Errorf("Invalid default value for slider")
		case "toggle":
			header = append(header, 0x12)
			if _, ok := setting.Default.(bool); ok {
				if setting.Default.(bool) {
					defaultValue = []byte{1}
					break
				}
				defaultValue = []byte{0}
				break
			}
			return nil, fmt.Errorf("Invalid default value for toggle")
		default:
			return nil, fmt.Errorf("Unknown setting type: '" + setting.Type + "'")
		}

		header = append(header, append([]byte(setting.Id), 0)...)
		header = append(header, append([]byte(setting.Name), 0)...)
		header = append(header, defaultValue...)
		if data != nil {
			header = append(header, data...)
		}
	}

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
		default:
			return nil, fmt.Errorf("Unknown block type: '" + blockType + "'")
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
		fileName = meta.Id + ".see"
	}

	if err = os.WriteFile(fileName, output, 0644); err != nil {
		log.Fatal(err)
	}
}
