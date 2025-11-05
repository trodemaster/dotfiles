<!-- Source: https://guide.macports.org/chunked/development.practices.html -->

# 4.8. Portfile Best Practices

## Introduction

This section provides guidelines for creating Portfiles that install smoothly and maintain consistency between ports.

## 4.8.1. Port Style

Portfiles should be formatted like a table of keys and values. Key guidelines include:

- Use soft tabs (spaces instead of tabs)
- Include a modeline at the top:
```
# -*- coding: utf-8; mode: tcl; tab-width: 4; indent-tabs-mode: nil; c-basic-offset: 4 -*- vim:fenc=utf-8:ft=tcl:et:sw=4:ts=4:sts=4
```
- Left column should use single words
- Separate columns with spaces in multiples of four
- Use line continuation with backslashes for multiple items
- Place braces on the same line for variants and blocks

## 4.8.2. Don't Overwrite Config Files

Avoid overwriting user-modified configuration files during upgrades:

```tcl
post-destroot {
    file rename ${destroot}${prefix}/etc/apcupsd/apcupsd.conf \
                ${destroot}${prefix}/etc/apcupsd/apcupsd.conf.sample
}

post-activate {
    if {![file exists ${prefix}/etc/apcupsd/apcupsd.conf]} {
        file copy ${prefix}/etc/apcupsd/apcupsd.conf.sample \
                  ${prefix}/etc/apcupsd/apcupsd.conf
    }
}
```

## 4.8.6. Renaming or Replacing a Port

When a port needs to be replaced:

1. Add `replaced_by` with the new port name
2. Increase version/revision
3. Clear distfiles
4. Disable livecheck
5. Add a pre-configure error block

### Shortcut Method Using PortGroup

```tcl
PortSystem          1.0
PortGroup           obsolete 1.0

name                skrooge-devel
replaced_by         skrooge
```

## 4.8.7. Removing a Port

Consider the implications before removing a port from the MacPorts tree.
