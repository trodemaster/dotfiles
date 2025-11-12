<!-- Source: https://guide.macports.org/chunked/using.html -->

# Chapter 3. Using MacPorts

## 3.1. The port Command

The `port` command is the main utility for interacting with MacPorts, used to update Portfiles, the MacPorts infrastructure, and install and manage ports.

### Key Actions

#### 3.1.1. port help
Displays brief information about specified actions or general usage.

#### 3.1.2. port selfupdate
Updates the local ports tree and checks for MacPorts upgrades. Recommended for regularly updating available software packages.

#### 3.1.7. port search
Allows finding ports by partial name or description matches. Key flags include:
- `--name`: Search only port names
- `--line`: Compact output
- `--glob`: Wildcard search (default)
- `--regex`: Regular expression search

Example:
```
$ port search php*
```

#### 3.1.12. port install
Installs a specified port.

Tips for troubleshooting:
- Use `-v` or `-d` for verbose/debug output
- Clean and retry if installation fails
- Check `main.log` for detailed error information

#### 3.1.15. port uninstall
Removes an installed port.

Considerations:
- Cannot uninstall ports with active dependencies
- Use `--follow-dependents` to recursively uninstall
- Use `-f` to force uninstallation (not recommended)

### Other Useful Actions

- `sync`: Update ports tree
- `diagnose`: Check for environment issues
- `info`: Get port details
- `deps`: List port dependencies
- `variants`: Show available port variations
- `upgrade`: Update installed ports
- `clean`: Remove intermediate build files

## Conclusion

The `port` command provides comprehensive management of software packages through MacPorts, offering extensive options for searching, installing, updating, and maintaining ports.
