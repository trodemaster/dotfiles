<!-- Source: https://guide.macports.org/chunked/development.html -->

# Chapter 4. Portfile Development

## Introduction

A port in MacPorts is a distribution of software that can be compiled and installed using MacPorts. A `Portfile` describes the steps required to build and install software, including:
- Source code retrieval
- Patch application
- Build requirements and commands

## Portfile Basics

A MacPorts Portfile is a Tcl script that typically contains:
- Keyword/value combinations
- Tcl extensions
- Potentially arbitrary Tcl code

### Key Characteristics

- Not all installation details are in the Portfile
- MacPorts base provides default installation characteristics
- Portfiles only need to specify required options

## Port Installation Phases

The main phases in a Portfile are:
- Fetch
- Extract
- Patch
- Configure
- Build
- Destroot

### Phase Behavior

- Default phases work for standard configure/make/make install applications
- Non-standard applications can:
  - Augment phases using pre-/post- declarations
  - Override phases
  - Eliminate phases

> A port consists of multiple files in a directory, usually within a category subdirectory of a ports tree.

### Note

For detailed phase descriptions, refer to the Portfile Reference chapter.
