<!-- Source: https://guide.macports.org/chunked/reference.portgroup.html -->

# 5.10. PortGroups

## Introduction

PortGroups are include files for portfiles that provide common definitions and behaviors for groups of portfiles. They help reduce redundancy and simplify port creation. PortGroups are located in:

`${prefix}/var/macports/sources/rsync.macports.org/macports/release/tarballs/ports/_resources/port1.0/group/`

## Available PortGroups

The document describes several PortGroups:

1. GitHub PortGroup
2. GNUstep PortGroup
3. Golang PortGroup
4. Java PortGroup
5. Perl5 PortGroup
6. Python PortGroup
7. Ruby PortGroup
8. Xcode PortGroup

## Key Features

Each PortGroup provides specific functionality:

- Simplifies package configuration
- Handles common build and installation tasks
- Provides default settings and variables
- Supports multiple versions of programming languages
- Automates dependency management

## Example Usage

A typical PortGroup usage might look like:

```tcl
PortGroup           github 1.0
github.setup        author project version
```

## Detailed Documentation

The document provides extensive details for each PortGroup, including:
- Specific keywords
- Configuration options
- Default behaviors
- Example implementations

## Conclusion

PortGroups are a powerful mechanism in MacPorts for standardizing and simplifying port creation across different types of software packages.
