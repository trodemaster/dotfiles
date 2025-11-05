<!-- Source: https://guide.macports.org/chunked/development.local-repositories.html -->

# 4.7. Local Portfile Repositories

## Overview

This section explains how to create and test Portfiles locally before publishing them in the MacPorts ports tree.

## Steps to Create a Local Portfile Repository

### 1. Edit sources.conf

Open `sources.conf` in a text editor and add a file URL pointing to your local repository:

```
file:///Users/julesverne/ports
rsync://rsync.macports.org/macports/release/tarballs/ports.tar [default]
```

**Note**: Place the file URL before the rsync URL so local Portfiles take precedence.

### 2. Create Repository Structure

Create directories for your port:
- Main category directory
- Port-specific directory

Example:
```
~/ports/games/bestevergame/
```

### 3. Add Portfile and Patches

- Create the Portfile in the port directory
- If needed, create a `files` directory for patch files

### 4. Generate Port Index

Run `portindex` in the local repository:

```bash
cd ~/ports
portindex
```

This creates an index that makes the local port searchable and installable.

### 5. Verify Port Availability

After indexing, you can search and install your local port like any other MacPorts package.

## Best Practices

- Match directory names to port names
- Place ports in appropriate category directories
- Use `portindex` to update the local repository index
