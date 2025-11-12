<!-- Source: https://guide.macports.org/chunked/internals.images.html -->

# 6.3. Port Images

MacPorts has a unique capability to allow multiple versions, revisions, and variants of the same port to be installed simultaneously, enabling users to test new port versions without removing existing working versions.

## How Port Images Work

This functionality stems from MacPorts' approach of initially installing ports to an intermediate location, rather than their final "activated" location. The process works as follows:

1. Ports are installed to an intermediate image location
2. Ports are only made available after an **activation phase**
3. During activation, files are extracted from the image repository
4. When a port is deactivated, its files are removed from activated locations (typically under `${prefix}`)
5. The port's image remains intact even after deactivation

## Benefits

The design allows for flexible port management, giving users the ability to:

- Maintain multiple versions of the same software
- Test new versions without disrupting existing installations
- Easily switch between different port configurations
- Roll back to previous versions if needed

## Viewing Port Image Location

To view an installed port's image location, you can use the appropriate `port` command to inspect the image directory.

## Navigation

- Previous: [6.2. Configuration Files](internals-configuration-files.md)
- Up: [Chapter 6. MacPorts Internals](internals.md)
- Next: [6.4. APIs and Libs](internals-apis.md)
