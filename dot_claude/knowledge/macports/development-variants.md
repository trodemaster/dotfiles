<!-- Source: https://guide.macports.org/chunked/development.variants.html -->

# 4.4. Port Variants

## Introduction

Variants are a way for port authors to provide options that can be invoked at install time. They are declared in the global section of a Portfile using the "variant" keyword.

## 4.4.1. Example Variants

Variants commonly modify dependencies, configure arguments, and build arguments. Here are some example variants:

```tcl
variant fastcgi description {Add fastcgi binary} {
    configure.args-append \
            --enable-fastcgi \
            --enable-force-cgi-redirect \
            --enable-memory-limit
}

variant gmp description {Add GNU MP functions} {
    depends_lib-append port:gmp
    configure.args-append --with-gmp=${prefix}
}

variant sqlite description {Build sqlite support} {
    depends_lib-append \
        port:sqlite3
    configure.args-delete \
        --without-sqlite \
        --without-pdo-sqlite
    configure.args-append \
        --with-sqlite \
        --with-pdo-sqlite=${prefix} \
        --enable-sqlite-utf8
}
```

> **Note**: Variant names may only contain A-Z, a-z, and underscore characters.

Another example of a variant modifying configuration:

```tcl
variant x11 description {Builds port as an X11 program with Lucid widgets} {
    configure.args-delete   --without-x
    configure.args-append   --with-x-toolkit=lucid \
                            --without-carbon \
                            --with-xpm \
                            --with-jpeg \
                            --with-tiff \
                            --with-gif \
                            --with-png
    depends_lib-append      lib:libX11:XFree86 \
                            lib:libXpm:XFree86 \
                            port:jpeg \
                            port:tiff \
                            port:libungif \
                            port:libpng
}
```

## 4.4.2. Variant Actions in a Phase

For complex variant actions, use `variant_isset` to modify phase behavior.
