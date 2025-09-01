package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

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

func ProcessBlockInfo(source string) ([]byte, error) {
	reader := strings.NewReader(source)
	stmts, err := parse.Parse(reader, "main.lua")
	if err != nil {
		return nil, err
	}

	var types []byte
	var names []byte

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

		switch commentLine[9:] {
		case "command":
			types = append(types, 0x1)
		case "hat":
			types = append(types, 0x2)
		case "event":
			types = append(types, 0x3)
		case "reporter":
			types = append(types, 0x4)
		case "boolean":
		case "bool":
			types = append(types, 0x5)
		}

		names = append(names, append([]byte(block), 0)...)
	}

	return append(append(types, 0), names...), nil
}

func CreateHeader(meta *Metadata, luaSource string) ([]byte, error) {
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

	blockInfo, err := ProcessBlockInfo(luaSource)
	if err != nil {
		return nil, err
	}
	header = append(header, blockInfo...)

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
	meta, err := ParseJSON()
	if err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile("main.lua")
	if err != nil {
		log.Fatal(err)
	}

	output, err := CreateHeader(meta, string(data))
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
