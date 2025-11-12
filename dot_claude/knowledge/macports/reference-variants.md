<!-- Source: https://guide.macports.org/chunked/reference.variants.html -->

# 5.5. Variants

## Introduction

MacPorts variants are conditional modifications of port installation behavior. There are two main types:

1. User-selected variants
2. Platform variants

## 5.5.1. User-Selected Variants

User-selected variants allow users to enable specific port options during installation. Key characteristics include:

- Variant names can only contain letters, numbers, and underscores
- Can define dependencies and conflicts with other variants
- Can be specified using the `variant` keyword

### Example Variant Declaration

```tcl
variant gnome requires glib {
    configure.args-append   --with-gnome
    depends_lib-append      port:gnome-session
}
```

### Default Variants

The `default_variants` keyword allows port authors to specify variants enabled by default:

```tcl
default_variants    +ssl +tcpd
```

Users can suppress default variants by using a "-" prefix.

### Universal Variant

- Enabled by default on macOS
- Can be overridden or suppressed if needed

## 5.5.2. User-Selected Variant Descriptions

Variant descriptions should be:
- Short and clear
- Capitalized
- Without trailing punctuation
- Descriptive but concise

Example:
```tcl
variant bar description {Add IMAP support} {}
```

## 5.5.3. Platform Variants

Platform variants allow customization based on:
- Operating system
- OS version
- Hardware architecture

### Example Platform Variants

```tcl
platform darwin 10 {
    configure.env-append LIBS=-lresolv
}

platform darwin i386 {
    configure.args-append --disable-mmx
}
```

**Note**: Multiple platform statements are needed to specify different combinations of OS and architecture.
