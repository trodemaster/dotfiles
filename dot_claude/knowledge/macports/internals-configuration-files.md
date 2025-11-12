<!-- Source: https://guide.macports.org/chunked/internals.configuration-files.html -->

# 6.2. Configuration Files

MacPorts has three main configuration files located in `${prefix}/etc/macports`:

## Overview
The configuration files are typically not modified by general users and contain options for advanced users and port developers.

## Configuration File Format
- Uses simple key/value pairs
- Separated by space or tab
- Lines starting with '#' are comments
- Empty lines are ignored

## 6.2.1. macports.conf

### Key Configuration Options
- `prefix`: Directory where ports are installed (default: `/opt/local`)
- `portdbpath`: Working data directory
- `build_arch`: Machine architecture for building
- `applications_dir`: Location for macOS application bundles
- `buildfromsource`: Controls source vs. pre-built archive usage
- `portarchivetype`: Archive format for port images

### Notable Settings
- `buildmakejobs`: Number of simultaneous make jobs
- `portautoclean`: Automatically clean after port installation
- `universal_archs`: Architectures for universal binaries

## 6.2.2. sources.conf
- Enables rsync synchronization of ports tree
- Default source: `rsync://rsync.macports.org/macports/release/tarballs/ports.tar`
- Supports local repository configuration

## 6.2.3. variants.conf
- Optional file for global variant specifications
- Ignored if a specified variant is not supported by a Portfile
