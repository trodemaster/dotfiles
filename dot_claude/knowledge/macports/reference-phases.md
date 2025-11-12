<!-- Source: https://guide.macports.org/chunked/reference.phases.html -->

# 5.3. Port Phases

## Introduction

The MacPorts port installation process involves several distinct phases:

1. **Fetch**: Download distribution files from specified sites
2. **Checksum**: Verify downloaded files' integrity
3. **Extract**: Unpack distribution files
4. **Patch**: Apply optional patch files to source code
5. **Configure**: Prepare source code for compilation
6. **Build**: Compile the source code
7. **Test**: Run optional test suites
8. **Destroot**: Install files to an intermediate location
9. **Install**: Archive installed files
10. **Activate**: Move files to final installed locations

## Key Characteristics

### Fetch Phase
- Downloads files from `master_sites`
- Supports multiple mirror sites
- Can fetch from various version control systems (Git, SVN, etc.)

### Destroot Phase
Critical features:
- "Stages" installation in an intermediate location
- Enables clean port uninstalls
- Allows multiple port versions on the same host

### Configuration
MacPorts provides flexible configuration options through keywords like:
- `configure.cmd`
- `configure.args`
- `configure.env`
- Compiler selection
- Architecture support

### Build and Test
- Supports parallel builds
- Optional test phase
- Configurable build environments

## Notable Details

- Uses standard GNU coding practices
- Supports multiple architectures
- Provides extensive customization through Portfile keywords

The document provides an in-depth technical reference for understanding MacPorts' port installation process and configuration mechanisms.
