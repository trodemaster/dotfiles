<!-- Source: https://guide.macports.org/chunked/using.binaries.html -->

# 3.4. Port Binaries

MacPorts can pre-compile ports into binaries so applications need not be compiled when installing on a target system. MacPorts supports two types of binaries: archives and packages.

## 3.4.1. Binary Archives

Binary archives can only be used on a target system running MacPorts. They allow MacPorts utilities to skip the build phase and begin installation after the destroot phase.

Key features:
- Automatically created when a port is installed
- Can be downloaded from a server
- MacPorts buildbot infrastructure creates prebuilt binary packages
- Archives uploaded to `packages.macports.org` if license allows

To manually create an archive:

```bash
$ port -d install logrotate
```

Binary archive files are stored in `${prefix}/var/macports/software/`. The archive type is set in `macports.conf` using the `portarchivetype` key, with `tbz2` as the default.

## 3.4.2. Binary Packages

Binary packages are standalone binary installers that do not require MacPorts on the target system. They are typically `.pkg` (macOS Installer packages) or `.dmg` disk images.

### Warning

When creating packages for redistribution:
- Avoid using standard MacPorts installation in `/opt/local`
- Use a custom MacPorts installation with a specific prefix

Example commands:

Create a macOS `.pkg` installer:
```bash
$ port pkg pstree
```

Create a macOS `.dmg`:
```bash
$ port dmg pstree
```

Create a metapackage with dependencies:
```bash
$ port mpkg pstree
```

Locate a port's work directory:
```bash
$ port work pstree
```
