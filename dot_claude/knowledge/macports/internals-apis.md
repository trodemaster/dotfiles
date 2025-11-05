<!-- Source: https://guide.macports.org/chunked/internals.apis.html -->

# 6.4. APIs and Libs

The MacPorts system is composed of three Tcl libraries:

- MacPorts API - MacPorts public API for handling Portfiles, dependencies, and registry
- Ports API - API for Portfile parsing and execution
- pextlib - C extensions to Tcl

## 6.4.1. Ports API

The Ports API is located in `base/src/port1.0` and provides primitives for:

- Managing target registrations
- Option/default handling
- Executing target procedures
- Managing state files
- Providing Portfile Tcl extensions
- Providing UI event access

> Portfiles are executed in a Tcl interpreter as Tcl code, so every Portfile option must be a Tcl procedure.

## 6.4.2. MacPorts API

Located in `base/src/macports1.0`, the MacPorts API provides:

- Dependency support
- Dependency processing
- UI abstractions
- Registry management
- API exports for client applications

Key routines include:
- `mportinit`: Initialize MacPorts system
- `mportsearch`: Search `PortIndex` for ports
- `mportopen`: Open a Portfile
- `mportclose`: Close a Portfile
- `mportexec`: Execute a target
- `mportinfo`: Return PortInfo array
- `mportdepends`: Return port dependencies

## 6.4.3. pextlib

pextlib is a Tcl library providing C extensions to add capabilities to Tcl procedures, such as interfaces to system calls like flock(2) and mkstemp(3).
