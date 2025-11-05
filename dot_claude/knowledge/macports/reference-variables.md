<!-- Source: https://guide.macports.org/chunked/reference.variables.html -->

# 5.2. Global Variables

Global variables are variables available to any Portfile. For additional variables in MacPorts Portgroups, see portgroup(7).

All variables except `prefix` are read-only!

## Key Global Variables

### prefix
- Installation prefix
- Default: `/opt/local`
- Can be overridden on a per-port basis

### Path-Related Variables
- `libpath`: Path to MacPorts Tcl libraries
- `portpath`: Full path to the current Portfile
- `filesdir`: Path to files directory (relative to portpath)
- `filespath`: Full path to files directory
- `workpath`: Full path to work directory
- `worksrcpath`: Full path to extracted source code
- `destroot`: Full path for software installation
- `distpath`: Location for downloaded distfiles

### System and OS Variables
- `os.platform`: Operating system platform (e.g., "darwin")
- `os.arch`: Hardware architecture
- `os.version`: Host OS version number
- `os.endian`: Processor endianness
- `os.major`: Major OS version number
- `macos_version`: Full macOS version
- `macos_version_major`: Major macOS version

### Development and Build Variables
- `xcodeversion`: Installed Xcode version
- `xcodecltversion`: Installed Command Line Tools version
- `universal_possible`: Boolean for universal binary build possibility

### User and Group Variables
- `install.user`: Unix user at port installation
- `install.group`: Unix group at port installation
