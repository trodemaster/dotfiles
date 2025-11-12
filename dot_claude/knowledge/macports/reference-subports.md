<!-- Source: https://guide.macports.org/chunked/reference.subports.html -->

# 5.6. Subports

## Overview

The subport declaration allows MacPorts to define additional ports with specific configurations. Here's how it works:

### Syntax
```
subport name body
```

### Example Portfile

```tcl
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

### Important Note

> Because MacPorts treats each subport as a separate port declaration, each subport will have its own, independent phases.

Key characteristics:
- Subports have independent build phases
- Subports share the global declaration part
- Distfiles are fetched only once and reused by subsequent subports

This mechanism allows for flexible port configurations within a single Portfile, enabling developers to create multiple related port variants efficiently.
