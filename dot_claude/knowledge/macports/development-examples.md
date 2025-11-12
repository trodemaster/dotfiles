<!-- Source: https://guide.macports.org/chunked/development.examples.html -->

# 4.3. Example Portfiles

## Introduction

This section explores various Portfile examples, demonstrating different techniques for creating and modifying port installation processes.

## 4.3.1. A Basic Portfile

A basic Portfile for the `rrdtool` package includes:
- Metadata like name, version, categories
- Source download information
- Checksums
- Dependencies
- Configuration arguments

```tcl
PortSystem          1.0

name                rrdtool
version             1.2.23
categories          net
platforms           darwin
license             GPL-2+
maintainers         julesverne
description         Round Robin Database
long_description    RRDtool is a system to store and display time-series data
homepage            https://people.ee.ethz.ch/~oetiker/webtools/rrdtool/
master_sites        https://oss.oetiker.ch/rrdtool/pub/

# Additional configuration details...
```

## 4.3.2. Augmenting Phases Using pre- / post-

Demonstrate adding extra steps to the installation phase:

```tcl
post-destroot {
    file mkdir ${destroot}${prefix}/share/doc/${name}/examples
    file copy ${worksrcpath}/examples/ \
        ${destroot}${prefix}/share/doc/${name}/examples
}
```

## 4.3.3. Overriding Phases

Completely replace the default installation phase:

```tcl
destroot {
    xinstall -m 755 -d ${destroot}${prefix}/share/doc/${name}
    xinstall -m 755 ${worksrcpath}/README ${destroot}${prefix}/share/doc/${name}
}
```

## 4.3.4. Eliminating Phases

Remove a specific build phase:

```tcl
build {}
```

## 4.3.5. Creating a StartupItem

Define a startup item in the global Portfile section:

```tcl
startupitem.create      yes
startupitem.name        nmicmpd
startupitem.executable  "${prefix}/bin/nmicmpd"
```
