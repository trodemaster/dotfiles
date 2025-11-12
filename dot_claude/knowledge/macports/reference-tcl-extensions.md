<!-- Source: https://guide.macports.org/chunked/reference.tcl-extensions.html -->

# 5.7. Tcl Extensions & Useful Tcl Commands

A MacPorts Portfile is a Tcl script that can use various Tcl commands and MacPorts-specific extensions for port development.

## File Operations

### Standard Tcl File Commands
- `file copy`: Copy a file
- `file rename`: Rename a file
- `file delete [-force]`: Remove a file or directory
- `file mkdir`: Create a directory

### MacPorts File Macros
- `copy`: Shorthand for `file copy`
- `move`: Handles case-sensitive renames
- `delete`: Shorthand for `file delete -force`
- `touch`: Mimics BSD touch command
- `ln`: Mimics BSD ln command

### xinstall Command
A versatile file installation command with options for:
- Copying files
- Creating directories
- Setting ownership and permissions

Example:
```tcl
xinstall -m 640 ${worksrcpath}/README ${destroot}${prefix}/share/doc/${name}
```

## String Manipulation

### strsed
Perform string manipulations using regular expressions:
- Replace first or all instances of a pattern
- Uses sed-like syntax

### reinplace
Replace text in-place within files:
- Supports different locales
- Can modify multiple files
- Allows extended regular expressions

## User and Group Management

### User Commands
- `adduser`: Create a new local user
- `existsuser`: Check if a user exists
- `nextuid`: Get next available user ID

### Group Commands
- `addgroup`: Create a new local group
- `existsgroup`: Check if a group exists
- `nextgid`: Get next available group ID

**Note**: The document suggests these commands should be used carefully and only when necessary.
