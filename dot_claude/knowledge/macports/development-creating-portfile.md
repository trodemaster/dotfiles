<!-- Source: https://guide.macports.org/chunked/development.creating-portfile.html -->

# 4.2. Creating a Portfile

This section provides a detailed guide for creating a Portfile for a standard open source application. Here are the key components:

## 1. Modeline (Optional)
```
# -*- coding: utf-8; mode: tcl; tab-width: 4; indent-tabs-mode: nil; c-basic-offset: 4 -*- vim:fenc=utf-8:ft=tcl:et:sw=4:ts=4:sts=4
```

## 2. Required Components

### PortSystem
```tcl
PortSystem          1.0
```

### Basic Port Information
- `name`: Port name
- `version`: Port version
- `categories`: Port categories (first should match ports tree directory)
- `platforms`: Supported platforms
- `maintainers`: People responsible for port maintenance
- `description`: Short port description
- `long_description`: Detailed port description
- `homepage`: Project website
- `master_sites`: Download URLs

### Checksums
- Use `rmd160` and `sha256` for security
- Include file size

### Dependencies
```tcl
depends_lib         port:perl5.8 \
                    port:tcl \
                    port:zlib
```

### Optional Configuration
```tcl
configure.args      --enable-perl-site-install \
                    --mandir=${prefix}/share/man
```

## Key Notes
- Maintainers can be listed with GitHub usernames or obfuscated email addresses
- Checksums can be verified by running MacPorts commands
- Dependencies ensure required ports are installed before the target port

The guide provides a comprehensive template for creating a standard Portfile for MacPorts.
