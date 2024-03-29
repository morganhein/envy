# envy

### Intended usage
- initial configuration 
- maintenance of an environment and backups
- distributing an environment

In a nutshell, envy performs two major duties:
1. Installs packages and dependencies requested by a platform-agnostic config file
2. Syncs dotfiles and other configuration files from a repository, and then symlinks those files into a home/config directory

Think of the first duty as a "universal package manager" and the second "automatic git and management of dotfiles".
Combined they bootstrap a configured environment. The intent is for the configuration to handle various OS and distros agnostically.

#### The main goals
- ease of use - lower the friction to usage
- cross-platform, specifically *nix/darwin
- minimal dependencies, currently your package manager, bash, and the "which" command
- should do "the right thing" whenever presented with options

#### Specific non-goals:
- Speed (Not like it's really slow either.)
- Efficiency (no state tracking, that's the package manager's job)

For you terraform nerds:
Think of this as a terraform for... you. It builds your environment for you.

For everyone else, it's a combination of a dotfile manager and a very basic universal package manager. Since it can potentially support nearly any *nix style environment (and maybe even Windows), that means it is the lowest common denominator of all of these environments. Not to speak ill of my own sofware, but that means it's very limited and not very smart. However, it makes creating the same set of personal environments to be applied in many different contexts/operating systems very easy and fast. Not to mention it's just awesome.

This is heavily inspired by other dotfile and home directory managers. Specifically, homely and homemaker.

## Installation
TODO: add installation line. The intent is a single curl call that both downloads the binary for the appropriate architecture/environment, but also run a personal configuration to start the envy process
```bash
curl <someUrl> - runs script - dls and installs binary - optionally also runs passed in configuration
```

### As a library
This package was designed with the intent of being consumed as a library. You can use it in go by:
```bash
go get github.com/morganhein/envy
```

## Overview of how this works
The intent is for an individual to write and maintain a single file that declares all the steps required to set up and bootstrap an environment in a specific way. If the user desires, they can also use this to manage their configuration and misc dotfiles.

The originally supported environments are: 
- arch linux, using pacman and yay
- fedora and friends, using dnf
- alpine linux, using apk
- ubuntu and friends, using apt
- mac using brew

This adds another attack vector for you, the user, because now there is automation in your setup pipeline. You must audit both
this software, and any scripts/commands you want run during setup, to ensure they meet your security requirements. This software, by itself, 
does nothing. However, given a malicious configuration file, it is very possible for you to install something with a malicious intent.

More information on the security posture of envy can be found below, under security. (TODO: LINK MAYBE?)

Once a user writes a configuration file, they can apply that config to their current environment. It will attempt to install the software and link any configuration/dotfiles that the user requested.

Once a system is up and running, there are facilities for maintaining and updating the dotfiles and ASH configuration.

### Usage:

For more usage information, read [USAGE.md](USAGE.md)

need functions for:
- running config(s)
    - interactive mode
    - run without applying links
    - run without installing packages
- updating links and dotfiles

`envy run [configuration file] --installers=gvm,brew <task>`

### Stretch-goals
1. Make the config declarative, so that if a section/app is removed, then envy reconciles that difference and removes it.. maybe
2. Have this tool spit out an actual shell script called "envy.sh" that is an auditable set of actions that match what this tool would do given a specific installation target/task
3. Make pretty and interactive with https://github.com/rivo/tview