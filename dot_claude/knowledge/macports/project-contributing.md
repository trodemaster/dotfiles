<!-- Source: https://guide.macports.org/chunked/project.contributing.html -->

# 7.3. Contributing to MacPorts

## Introduction

Contributing to MacPorts involves submitting new ports, enhancing existing ports, and potentially becoming a port maintainer. The project strongly prefers contributions through GitHub pull requests over Trac tickets.

## 7.3.1. New Ports

To contribute a new port:

1. Run port lint for the new port
2. Submit via:
   - GitHub pull request (preferred method)
   - Trac ticket with:
     - Type: submission
     - Component: ports
     - Port field: new port name
     - Attach Portfile and patch files

If no response within a few days, email `macports-dev@lists.macports.org` with a review request.

## 7.3.2. Port Enhancements

For port updates or improvements:

1. Create a Portfile patch
2. Run port lint
3. Submit via:
   - GitHub pull request (preferred)
   - Trac ticket with:
     - Type: enhancement/defect/update
     - Component: ports
     - Port field: port name
     - Cc: port maintainer

## 7.3.3. Becoming a Port Maintainer

Requirements:
- Email address and GitHub account
- Portfile copy
- MacPorts Trac account
- Interest in the software

Steps to become a maintainer:
1. Check if port is unmaintained
2. Copy port directory
3. Verify port information
4. Edit Portfile to add maintainer details
5. Create patch
6. Submit via GitHub pull request or Trac ticket
7. Follow up if no response

Maintainers are expected to review pull requests and tickets for their ports and seek help from the community when needed.
