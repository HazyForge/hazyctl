# Hazyctl

Hazyctl is a utility tool designed for platform engineers, devops, or anyone who needs to manage cloud infrastructure. It provides a set of specific utilities and commands that help manage, configure, and update cloud infrastructure efficiently.

## Features

- Manage cloud secrets
- Update the tool to the latest version
- Configure cloud settings
- And more...

## Installation

To install Hazyctl, you can use the following command which will download and run the installation script:

Linux/MacOS:

```bash
curl -sSL https://raw.githubusercontent.com/hazyforge/hazyctl/master/install/install.sh | bash
```

Windows:

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; Invoke-WebRequest -Uri https://raw.githubusercontent.com/hazyforge/hazyctl/master/install/install.ps1 -OutFile install.ps1; .\install.ps1
```

Supported Features:
- Self Update Command
- `secret azure export` Secrets To Local File
- `secret azure migrate` Secrets between vaults 
