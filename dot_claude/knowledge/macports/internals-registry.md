<!-- Source: https://guide.macports.org/chunked/internals.registry.html -->

# 6.5. The MacPorts Registry

## Overview

The MacPorts registry is a system for tracking and managing installed ports, providing utilities with information about:

- Installed ports
- Dependencies
- Port images
- File conflicts

## Registry Files

### SQLite Registry
- Located at `${portdbpath}/registry`
- Default path: `${prefix}/var/macports/registry`
- Main file: `registry.db`
- Temporary directory: `portfiles`

### Legacy Flat File Registry
- Located at `${portdbpath}/receipts`
- Default path: `${prefix}/var/macports/receipts`
- Tracked files:
  - `file_map.db`
  - `dep_map.bz2`

## Registry API

The registry provides a public API in the `registry1.0` Tcl package with key functions:

### Entry Management
- `registry::new_entry`: Create a new registry entry
- `registry::open_entry`: Open an existing entry
- `registry::entry_exists`: Check if an entry exists
- `registry::write_entry`: Write a receipt
- `registry::delete_entry`: Delete a receipt

### Property Handling
- `registry::property_store`: Store a property
- `registry::property_retrieve`: Retrieve a property

### Port Information
- `registry::installed`: Get installed ports
- `registry::location`: Find port installation location

### File and Dependency Tracking
- `registry::file_registered`: Check file ownership
- `registry::port_registered`: List files for a port
- `registry::list_depends`: List port dependencies
- `registry::list_dependents`: List ports depending on a port

The API allows comprehensive management of port installations, dependencies, and file tracking.
