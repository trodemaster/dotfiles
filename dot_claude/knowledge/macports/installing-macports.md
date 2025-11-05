<!-- Source: https://guide.macports.org/chunked/installing.macports.html -->

# 2.2. Install MacPorts

## Overview

MacPorts can be installed on macOS using several methods:
1. macOS Package Install (Recommended)
2. Source Install
3. Git Install
4. Multiple MacPorts Copies

## 2.2.1. macOS Package Install

### Steps:
1. Download the appropriate `.pkg` installer for your macOS version from GitHub
   - Versions available for macOS 10.12 Sierra through 15 Sequoia
   - Latest version: MacPorts-2.11.5

2. Double-click the package installer
3. Confirm installation by running `port` in a new terminal window

## 2.2.2. Source Install

### Steps:
1. Download and extract MacPorts 2.11.5 tarball
2. Run installation commands in terminal
3. Set up shell environment

## 2.2.3. Git Install

For developers wanting the latest version:

1. Check out MacPorts source from Git
2. Build and install MacPorts
3. (Optional) Configure MacPorts to use port information from Git
4. Set up environment variables

## 2.2.4. Install Multiple MacPorts Copies

- Use `--prefix` option to install in different locations
- Change applications directory to avoid conflicts
- Use `--without-startupitems` to prevent launch conflicts

**Note:** Requires careful configuration to manage multiple instances.
