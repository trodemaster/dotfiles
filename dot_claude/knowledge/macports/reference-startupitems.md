<!-- Source: https://guide.macports.org/chunked/reference.startupitems.html -->

# 5.8. StartupItems

## Introduction

A StartupItem in MacPorts is a facility to run background daemons using Apple's `launchd` system. It helps manage programs that run continuously, such as mail servers or network listeners.

## StartupItem Attributes

### Key Attributes

- `startupitem.create`: Trigger StartupItem creation (default: no)
- `startupitem.autostart`: Automatically load after port activation (default: no)
- `startupitem.install`: Install link in `/Library` (default: yes)
- `startupitem.location`: Choose subdirectory for installation (default: LaunchDaemons)
- `startupitem.type`: StartupItem type (default: launchd)

### Logging Options

- `startupitem.logfile`: Path for logging events
- `startupitem.debug`: Enable additional debug logging
- `startupitem.logevents`: Control event logging

## Types of StartupItems

### Executable StartupItems

- Preferred method for running daemons
- Directly launches daemon process
- Uses `startupitem.executable` to specify daemon

### Script StartupItems

- Used when daemon requires a startup script
- Requires `startupitem.pidfile` for process monitoring
- Uses `startupitem.start`, `startupitem.stop`, and `startupitem.restart`

## Loading and Unloading

StartupItems can be loaded and unloaded using `launchctl` commands, which enable or disable the daemon's .plist file.

## Important Notes

- The `startupitem_type` in `macports.conf` can override StartupItem creation
- Some daemons require special handling to prevent multiple process instances
