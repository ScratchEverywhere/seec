# SEE (Scratch Everywhere! Extension)

IDK how to properly format one of these and it's a pretty simple format so...

The first part of a `.see` file is the magic string `SE! EXTENSION` (not
null-terminated.) It's followed by a null-terminated string representing the ID
of the extension. Then there's a null-terminated string representing the name of
the extension. There's a null-terminated string representing the description of
the extension after that. The next byte is used to figure out the permissions
needed by the extension. Followed by the Lua bytecode of the extension.

# SECE (Scratch Everywhere! Core Extension)

`.sece` files start with the magic string `SE! CORE.EXT` (not null-terminated.)
The magic string is followed by a null-terminated string representing the ID of
the extension. The ID is followed by a null-terminated representing the
description of the extension. The next byte is used to figure out the
permissions needed by the extension. Followed by the Lua bytecode of the
extension.

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
