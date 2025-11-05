<!-- Source: https://guide.macports.org/chunked/installing.macports.uninstalling.html -->

# 2.4. Uninstall MacPorts

Uninstalling MacPorts is a significant step that should be carefully considered. The guide recommends consulting the [macports-users](https://lists.macports.org/mailman/listinfo/macports-users) mailing list if you're unsure.

## 2.4.1. Uninstall All Ports

If the `port` command is working, uninstall all installed ports:

```bash
$ port uninstall installed
```

This leaves behind configuration files, databases, and MacPorts base software.

## 2.4.2. Remove Users and Groups

Remove the MacPorts user and group with admin privileges:

```bash
$ sudo dscl . -delete /Users/macports
$ sudo dscl . -delete /Groups/macports
```

## 2.4.3. Remove the Rest of MacPorts

To completely remove MacPorts, run:

```bash
$ sudo rm -rf \
    /opt/local \
    /Applications/MacPorts \
    /Library/LaunchDaemons/org.macports.* \
    /private/var/db/receipts/org.macports.*
```

**Note**: For macOS 10.15 Catalina or later with SIP enabled, remove the `macports` user first.

The guide emphasizes checking with the community before taking this drastic step and notes that not all paths may exist on your system.
