package main

import (
	"bytes"
	"encoding/binary"
	"github.com/yuin/gopher-lua"
)

func dumpString(buf *bytes.Buffer, str string) {
	binary.Write(buf, binary.LittleEndian, int32(len(str)+1))
	buf.WriteString(str)
	buf.WriteByte(0)
}

func dumpInt(buf *bytes.Buffer, i int) {
	binary.Write(buf, binary.LittleEndian, int32(i))
}

func dumpNumber(buf *bytes.Buffer, num lua.LValue) {
	binary.Write(buf, binary.LittleEndian, lua.LVAsNumber(num))
}

func dumpHeader(buf *bytes.Buffer) {
	binary.Write(buf, binary.BigEndian, uint32(0x1B4C7561))
	buf.WriteByte(0x51)
	buf.WriteByte(0)
	buf.WriteByte(1)
	buf.WriteByte(4)
	buf.WriteByte(4)
	buf.WriteByte(4)
	buf.WriteByte(8)
	buf.WriteByte(0)
}

func dumpFunction(buf *bytes.Buffer, f *lua.FunctionProto) {
	if f.SourceName != "" {
		dumpString(buf, f.SourceName)
	} else {
		buf.WriteByte(0)
	}
	dumpInt(buf, f.LineDefined)
	dumpInt(buf, f.LastLineDefined)
	buf.WriteByte(f.NumUpvalues)
	buf.WriteByte(f.NumParameters)
	buf.WriteByte(f.IsVarArg)
	buf.WriteByte(f.NumUsedRegisters)

	dumpCode(buf, f.Code)
	dumpConstants(buf, f.Constants)
	dumpPrototypes(buf, f.FunctionPrototypes)

	// DUMPING ZEROS IN PLACE OF DEBUG FIELDS
	dumpInt(buf, 0) // No line info
	dumpInt(buf, 0) // No locals
	dumpInt(buf, 0) // No upvalue names
}

func dumpCode(buf *bytes.Buffer, code []uint32) {
	dumpInt(buf, len(code))
	for _, inst := range code {
		binary.Write(buf, binary.LittleEndian, inst)
	}
}

func dumpPrototypes(buf *bytes.Buffer, protos []*lua.FunctionProto) {
	dumpInt(buf, len(protos))
	for _, proto := range protos {
		dumpFunction(buf, proto)
	}
}

func dumpConstants(buf *bytes.Buffer, consts []lua.LValue) {
	dumpInt(buf, len(consts))
	for _, cons := range consts {
		switch cons.Type() {
		case lua.LTNil:
			buf.WriteByte(0)
		case lua.LTBool:
			buf.WriteByte(1)
			if lua.LVAsBool(cons) {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		case lua.LTNumber:
			buf.WriteByte(3)
			dumpNumber(buf, cons)
		case lua.LTString:
			buf.WriteByte(4)
			dumpString(buf, lua.LVAsString(cons))
		}
	}
}

func DumpLua(proto *lua.FunctionProto) []byte {
	buf := new(bytes.Buffer)
	dumpHeader(buf)
	dumpFunction(buf, proto)
	return buf.Bytes()
}
