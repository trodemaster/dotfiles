<!-- Source: https://guide.macports.org/chunked/project.update-policies.html -->

# 7.4. Port Update Policies

## Introduction

Port maintainers typically have commit privileges to the Git repository for updating their own ports. However, The MacPorts Project does not restrict these privileges, so it's recommended to inform a port's maintainer before making updates.

## 7.4.1. Non-Maintainer Port Updates

Guidelines for updating ports you do not maintain:

1. If the maintainer is `<nomaintainer@macports.org>`, you can freely update or take maintainership.

2. If the maintainer includes `<openmaintainer@macports.org>`:
   - Minor updates are allowed without prior contact
   - Committers must thoroughly investigate changes
   - Verify the port builds and works correctly

3. For maintained ports without open maintenance:
   - Create patch files
   - Attach patches to a Trac ticket
   - Assign or Cc the maintainer
   - Wait for a response

Exceptions to requiring maintainer permission include:
- No maintainer response within 72 hours
- Port is officially abandoned
- Critical port affecting many users is broken

## 7.4.2. Port Abandonment

A port may be considered abandoned if:
- Bug not acknowledged for over three weeks
- Tickets resolved with no maintainer input
- Maintainer contact information is invalid

To initiate port abandonment:
1. File a Trac ticket with `[Port Abandoned] portname`
2. Assign to the maintainer
3. Set ticket type to Defect
4. Reference unacknowledged tickets
5. Indicate the specific port

The ticket remains open for 72 hours to allow the maintainer a final opportunity to respond.
