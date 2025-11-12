<!-- Source: https://guide.macports.org/chunked/using.common-tasks.html -->

# 3.3. Common Tasks

## Introduction

This section covers common operations for managing a MacPorts installation. The guide recommends reading this section as a primer for using MacPorts, with more details available in the port command documentation.

## Key Tasks

### 3.3.1. Updating Your Ports Tree

Use the `selfupdate` command to update the local ports tree:

```bash
$ port selfupdate
```

### 3.3.2. Show Ports Which Need Updating

Use `port outdated` to list ports with newer versions available:

```bash
$ port outdated
```

### 3.3.3. Upgrading Outdated Ports

To upgrade all outdated ports:

```bash
$ sudo port upgrade outdated
```

To upgrade a specific port:

```bash
$ sudo port upgrade [portname]
```

### 3.3.4. Removing Inactive Versions

To list inactive ports:

```bash
$ port installed
```

To remove all inactive ports:

```bash
$ sudo port uninstall inactive
```

### 3.3.5. Finding Port Dependencies

To find ports depending on a specific port:

```bash
$ port dependents [portname]
```

### 3.3.6. Finding Leaves

To identify ports that are no longer needed:

```bash
$ port leaves
```

To uninstall all leaves:

```bash
$ sudo port uninstall leaves
```

### 3.3.7. Managing Requested Ports

To mark a port as manually installed:

```bash
$ sudo port setrequested [portname]
```

The guide recommends periodically reviewing and cleaning up unnecessary ports to keep the MacPorts installation lean.

## Recommendation

Use the `port_cutleaves` port for an interactive method of managing unnecessary ports.
