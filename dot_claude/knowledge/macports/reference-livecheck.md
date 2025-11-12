<!-- Source: https://guide.macports.org/chunked/reference.livecheck.html -->

# 5.9. Livecheck / Distcheck

## Overview

Livecheck and distcheck are tools primarily useful for port maintainers to check for software updates and availability of distribution files.

## Livecheck Options

### livecheck.type
- Specifies the type of update check to perform
- Default options include:
  - `sourceforge`
  - `googlecode`
  - `freecode`
  - `moddate`
  - `regex`
  - `md5`
  - `none`

### Example Configuration
```tcl
livecheck.type      regex
livecheck.url       ${homepage}
livecheck.regex     "Generally Available (\d+(?:\.\d+)*)"
```

### Other Livecheck Keywords
- `livecheck.name`: Project name for checks
- `livecheck.distname`: File release name
- `livecheck.version`: Project version for checks
- `livecheck.url`: URL to query for updates
- `livecheck.regex`: Regular expression for version parsing
- `livecheck.md5`: Checksum for comparison

## Distcheck Options

### distcheck.check
- Controls distfile availability checks
- Options:
  - `moddate`: Check if Portfile is older than distfile
  - `none`: Disable checking

### Example
```tcl
distcheck.check     none
```

## Key Features
- Helps maintainers track software updates
- Provides flexible checking mechanisms
- Supports various source repositories and checking methods
