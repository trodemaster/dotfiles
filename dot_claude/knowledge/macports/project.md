<!-- Source: https://guide.macports.org/chunked/project.html -->

# Chapter 7. MacPorts Project

## 7.1. Using Trac for Tickets

### 7.1.1. Before Filing a New Ticket

Before creating a ticket, follow these guidelines:

- **Clean and try again**: Run `sudo port clean` if a build fails
- **Check the Problem Hotlist** for existing solutions
- **Search existing tickets** to avoid duplicates
- **Verify the issue is MacPorts-related**
- For 'port upgrade' problems, try:
  - `port uninstall foo`
  - Reinstall the port
  - Use `port upgrade outdated` for safest upgrades

### 7.1.2. Creating Trac Tickets

To create a ticket:
- Login with a GitHub account
- Click "New Ticket"
- Follow ticket guidelines

### 7.1.3. Trac Ticket Guidelines

#### Key Ticket Fields

| Field | Recommended Content |
|-------|---------------------|
| **Summary** | `$port $version [$variants]`: short problem summary |
| **Description** | Detailed problem explanation |
| **Type** | Choose: defect, enhancement, update, submissions, request |
| **Priority** | Normal or low (high reserved for developers) |
| **Component** | Select appropriate area (base, ports, guide, etc.) |

**Important Tips**:
- Use preview button
- Format code/logs with `{{{triple brackets}}}`
- Attach large logs instead of pasting
- Link GitHub commits using special formatting

Detailed guidelines are provided in the original document for each ticket field.
