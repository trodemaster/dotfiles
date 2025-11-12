<!-- Source: https://guide.macports.org/chunked/development.subport.html -->

# 4.5. Subports

## Introduction

MacPorts subports allow declaring multiple related ports in a single Portfile, which is especially useful when ports share variables or keywords.

## 4.5.1. Subport Declarations

The subport declaration syntax is:

```
subport name body
```

This causes MacPorts to define an additional port with the specified name, using keywords from the global Portfile section and the subport-specific body.

## 4.5.2. Subport Examples

Example Portfile demonstrating subports:

```
Portfile                   1.0

name                       example

depends_lib                aaa
configure.args             --bbb

subport example-sub1 {
    depends_lib-append     ccc
    configure.args         --ddd
}

subport example-sub2 {
    depends_lib-append     eee
    configure.args-append  --fff
}
```

This is equivalent to creating three separate Portfiles, each with its own configuration.

## 4.5.3. Subport Tips

Key points about subports:
- Each subport has independent phases (fetch, configure, build, etc.)
- Subports share the same `dist_subdir` by default
- Distfiles are fetched only once and reused by subsequent subports
