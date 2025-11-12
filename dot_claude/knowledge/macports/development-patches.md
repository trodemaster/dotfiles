<!-- Source: https://guide.macports.org/chunked/development.patches.html -->

# 4.6. Patch Files

## Introduction

Patch files are created using the Unix `diff` command and applied with the `patch` command to modify text files for bug fixes or functionality extensions.

## 4.6.1. Creating Portfile Patches

To create a Portfile patch:

1. Copy the original Portfile
2. Edit the file as desired
3. Use `diff -u` to create a unified diff patch file
4. Name the patch file descriptively (e.g., `Portfile-rrdtool.diff`)

Example patch:
```
--- Portfile.orig        2011-07-25 18:52:12.000000000 -0700
+++ Portfile    2011-07-25 18:53:35.000000000 -0700
@@ -2,7 +2,7 @@
 PortSystem          1.0
 name                foo

-version             1.3.0
+version             1.4.0
 categories          net
 maintainers         nomaintainer
 description         A network monitoring daemon.
```

## 4.6.2. Creating Source Code Patches

Guidelines for source code patches:

- Send patches to the application developer when possible
- Create one patch file per logical change
- Use descriptive filenames like `destdir-variable-fix.diff`

Steps to create a source code patch:
1. Duplicate the original file
2. Make necessary modifications
3. Use `diff -u` from the top-level source directory
4. Place the patch in `${portpath}/files`

Example patch:
```
--- src/Makefile.in.orig   2007-06-01 16:30:47.000000000 -0700
+++ src/Makefile.in       2007-06-20 10:10:59.000000000 -0700
@@ -131,23 +131,23 @@
        $(INSTALL_DATA)/gdata $(INSTALL_DATA)/perl

install-lib:
-       -mkdir -p $(INSTALL_LIB)
+       -mkdir -p $(DESTDIR)$(INSTALL_LIB)
```
