# seec (Scratch Everywhere! Extension Compiler)

## Usage

See the example for how to structure `meta.json` and `main.lua`.

```
Usage of seec:
  -m, --meta string     Path to the JSON file containing metadata. (default "meta.json")
  -o, --output string   Path to put the outputed .see or .sece.
  -i, --source string   Path to the Lua source code of the extension. (default "main.lua")
```

## File Format

See [FORMAT.md](FORMAT.md) for information about `.see` and `.sece` files.

## Installation

> [!NOTE]
> seec is on the AUR as the `seec` package.

Simply download the static executable from the releases page.

## Todo List

- [x] Parse `meta.json`
- [x] Create Header (part before lua bytecode)
- [x] Parse Command Line Arguments
- [x] Compile Lua into bytecode
- [x] Update the formats to add information about the types of blocks
