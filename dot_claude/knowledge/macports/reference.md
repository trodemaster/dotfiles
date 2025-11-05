<!-- Source: https://guide.macports.org/chunked/reference.html -->

# Chapter 5. Portfile Reference

## 5.1. Global Keywords

This section describes the global keywords used in MacPorts Portfiles to specify port information and configuration.

### PortSystem

The first non-comment line of a Portfile, defining the Portfile interpreter version:

```tcl
PortSystem          1.0
```

### name

Specifies the port's name, using only alphanumeric characters, underscores, dashes, and periods:

```tcl
name                foo
```

### version

Indicates the software version, typically in dotted decimal format:

```tcl
version             1.23.45
```

### revision

An optional integer (default 0) incremented when a port is updated independently of the software version:

```tcl
revision            1
```

### epoch

An optional integer used when a port is updated to a version that appears less than the previous version:

```tcl
epoch               1
```

### categories

Specifies the categories the ported software falls under:

```tcl
categories          net security
```

### maintainers

Lists the port's maintainers using GitHub usernames or email addresses:

```tcl
maintainers         @neverpanic \
                    jdoe \
                    example.org:julesverne
```

### description and long_description

Provide short and detailed descriptions of the software:

```tcl
description         a classic shooter arcade game
long_description    ${name} is a classic shooter arcade game derived from \
                    alien-munchers. Not suitable for children under two.
```

### homepage

The software's primary web site:

```tcl
homepage            https://www.example.org/apps/
```

### platforms

Lists the platforms where the port is expected to work:

```tcl
platforms           macosx freebsd
```

### Other Keywords

The document also covers additional keywords like:
- `supported_archs`
- `license`
- `use_xcode`
- `known_fail`
- `macosx_deployment_target`
- `installs_libs`
- `add_users`

Each keyword serves a specific purpose in configuring and describing a MacPorts port.
