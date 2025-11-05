<!-- Source: https://guide.macports.org/chunked/reference.dependencies.html -->

# 5.4. Dependencies

## Introduction

Free and open source software is highly modular, and MacPorts ports often require other ports to be installed beforehand. These prerequisites are called a port's "dependencies".

## Dependency Types

MacPorts supports several dependency keywords related to port install phases:

### Fetch Dependencies
- `depends_fetch`: Dependencies needed to download distfiles
- Checked before fetch, checksum, extract, patch, configure, build, destroot, install, and package phases

### Extract Dependencies
- `depends_extract`: Dependencies needed to unpack distfiles
- Checked before extract, patch, configure, build, destroot, install, and package phases

### Build Dependencies
- `depends_build`: Dependencies needed when building software
- Checked before configure, build, destroot, install, and package phases

### Library Dependencies
- `depends_lib`: Dependencies needed at build and runtime
- Checked before configure, build, destroot, install, and package phases

### Test Dependencies
- `depends_test`: Dependencies needed for testing
- Checked before test phase when `test.run yes`

### Run Dependencies
- `depends_run`: Dependencies needed when running software
- Checked before destroot, install, and package phases

## Port and File Dependencies

Two types of dependencies exist:
1. Port dependencies (preferred)
2. File dependencies

### Port Dependencies Example
```
depends_lib         port:rrdtool port:apache2
depends_build       port:libtool
depends_run         port:apache2 port:php5
```

### File Dependencies
Three types:
- `bin`: Programs in bin directories
- `lib`: Libraries in lib directories
- `path`: Specific file paths

File dependency example:
```
depends_lib         lib:libX11.6:xorg
depends_build       bin:glibtool:libtool
depends_run         path:lib/libltdl.a:libtool
```

Note: The specified port is only installed if the library, binary, or file is not found.
