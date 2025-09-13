# SEE (Scratch Everywhere! Extension)

IDK how to properly format one of these and it's a pretty simple format so...

The first part of a `.see` file is the magic string `SE! EXTENSION` (not
null-terminated.) It's followed by a null-terminated string representing the ID
of the extension. Then there's a null-terminated string representing the name of
the extension. There's a null-terminated string representing the description of
the extension after that. The next byte is used to figure out the permissions
needed by the extension. After the permissions is a byte representing supported
platforms. Following the supported platforms is any amount of bytes representing
types of blocks, this section is null-terminated. After the types is the
corresponding IDs in the same order, each one should be null-terminated and
there should be the same amount of IDs as types. Followed by the Lua bytecode of
the extension.

# SECE (Scratch Everywhere! Core Extension)

`.sece` files start with the magic string `SE! CORE.EXT` (not null-terminated.)
The magic string is followed by a null-terminated string representing the name
of the extension. The name is followed by a null-terminated representing the
description of the extension. The next byte is used to figure out the
permissions needed by the extension. After the permissions is a byte
representing supported platforms. Following the supported platforms is any
amount of bytes representing types of blocks, this section is null-terminated.
After the types is the corresponding IDs in the same order, each one should be
null-terminated and there should be the same amount of IDs as types. Followed by
the Lua bytecode of the extension.

# Permissions

- 1st bit represents if the extension needs local file system access. This is
  what should be used the majority of the time.
- 2nd bit represents if the extension needs root file system access. This is not
  recommended for most extensions.
- 3rd bit represents if the extension needs internet access, only HTTP and HTTPS
  is supported.
- 4th bit represents if the extension needs access to controller inputs.
- 5th bit represents if the extension needs access to rendering.
- 6th bit represents if the extension has update functions that need to be
  called each frame.
- 7th bit represents if the extension needs access to platform specific APIs
  (e.g. Amiibo or account information.)
- 8th bit represents if the extension needs access to the runtime (e.g. sprite
  position, costume, or variables.)

# Block Types

- `0x1`: Command
- `0x2`: Hat
- `0x3`: Event
- `0x4`: Reporter
- `0x5`: Boolean

# Supported Platforms Map

Just `&` these together.

- `0b00000001`: 3DS
- `0b00000010`: Wii U
- `0b00000100`: Wii
- `0b00001000`: GameCube
- `0b00010000`: Switch
- `0b00100000`: PC
- `0b01000000`: Vita
