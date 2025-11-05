<!-- Source: https://guide.macports.org/chunked/development.buildbot.html -->

# 4.9. MacPorts' buildbot

## Overview

The [buildbot](https://build.macports.org/) is a port build service using MacPorts Buildbot (MPBB) scripts. Key features include:

- Automatic port rebuilding after Git repository commits
- Binary archive generation for distributable ports
- Build testing across multiple macOS versions

## Build Process

When a maintainer commits changes to the MacPorts ports Git repository, the buildbot:

- Checks if port rebuilds are necessary
- Generates binary archives for distributable ports
- Supports builds across multiple macOS platforms

## Notification and Monitoring

- Build errors trigger email notifications to port maintainers
- Helps maintain port consistency across macOS versions
- Currently builds only default port variants

## Monitoring Resources

Maintainers can track builds through:

- [build.macports.org](https://build.macports.org/)
  - Waterfall view
  - Builders view

- [ports.macports.org](https://ports.macports.org/)
  - Port-specific build summaries
  - "Port Health" indicators
  - Detailed build information

The buildbot ensures ports remain functional across different macOS versions without requiring maintainers to have access to every platform.
