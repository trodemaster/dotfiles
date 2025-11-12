<!-- Source: https://guide.macports.org/chunked/using.variants.html -->

# 3.2. Port Variants

## Introduction

Variants are optional features for ports that can be enabled during installation. They allow port authors to provide customizable installation options that may not be useful for all users.

## Displaying Available Variants

To view available variants for a port, use the following command:

```
$ port variants <portname>
```

### Example: Apache2 Variants

For Apache2, variants might include:
- `eventmpm`: Experimental event MPM
- `openldap`: Enable LDAP support
- `preforkmpm`: Prefork MPM
- `universal`: Build for multiple architectures
- `workermpm`: Worker MPM

## Invoking Variants

### Selecting Variants

Install a port with a specific variant by using a plus sign:

```
$ port install <portname> +<variant>
```

Multiple variants can be selected simultaneously:

```
$ port install <portname> +variant1 +variant2
```

To deselect a default variant, use a minus sign:

```
$ port install <portname> -<variant>
```

### Important Notes

- MacPorts does not confirm variant selection
- Misspelled variants will not trigger warnings
- Variants can be applied to dependencies
- MacPorts remembers variants used in previous installations

## Negating Default Variants

If a port has default variants specified in its Portfile, you can disable them:

```
$ port install <portname> -<default_variant>
```
