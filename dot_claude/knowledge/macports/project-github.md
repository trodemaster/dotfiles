<!-- Source: https://guide.macports.org/chunked/project.github.html -->

# 7.2. Using Git and GitHub

The MacPorts project uses [Git](https://git-scm.com/) for version control, with master repositories hosted on [GitHub](https://github.com/).

## Public Repositories

MacPorts maintains public repositories for:
- MacPorts system
- MacPorts ports
- MacPorts guide

For those new to Git, the book [Pro Git](https://git-scm.com/book/en/v2) is recommended as a free online resource.

## Contributing via Pull Request

### Steps to Submit a Pull Request:

1. Make changes in your own Git repository:
   - Fork the appropriate repository
   - Create a branch
   - Make changes
   - Test changes (Portfile changes should pass `port lint`)
   - Commit with standard commit messages
   - Rebase to minimize commits

2. Push changes to your GitHub repository

3. Create a pull request to the MacPorts repository

4. Await review and potential feedback

## Additional Notes

- The MacPorts team welcomes improvements to:
  - New ports
  - Port fixes
  - Base software
  - Documentation
  - Website

- If no response within a few days, contact `macports-dev@lists.macports.org`

The project encourages community contributions and appreciates collaborative improvements.
