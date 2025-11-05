<!-- Source: https://guide.macports.org/chunked/installing.shell.html -->

# 2.5. MacPorts and the Shell

MacPorts requires setting certain environment variables in the shell. When installed via the macOS package installer, a "postflight" script automatically configures shell configuration files.

## 2.5.1. The Postflight Script

The script sets three key variables:

### Required: PATH Variable
```
export PATH=/opt/local/bin:/opt/local/sbin:$PATH
```

### Optional: MANPATH Variable
```
export MANPATH=/opt/local/share/man:$MANPATH
```

### Optional: DISPLAY Variable
- Only set on macOS versions earlier than 10.5
```
export DISPLAY=:0.0
```

## 2.5.2. Verify the Configuration File

To check environment settings, use the `env` command in the terminal. Note that changes take effect only after opening a new terminal session.

## 2.5.3. Optional Editor Variables

MacPorts checks editor variables in this order:
1. `MP_EDITOR`
2. `VISUAL`
3. `EDITOR`

Example configurations:

```bash
# Use nano
export EDITOR=/usr/bin/nano

# Use BBEdit
export EDITOR=/Applications/BBEdit.app/Contents/Helpers/bbedit_tool

# Keep command-line editor as default, use BBEdit for portfiles
export EDITOR=/usr/bin/vi
export MP_EDITOR=/Applications/BBEdit.app/Contents/Helpers/bbedit_tool
```
