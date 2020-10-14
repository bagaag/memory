# Memory

Memory is a command line application that captures, stores and manages the data
that describes human experience. Information is organized into people, places,
things and events. A network of relationships between these entries relays human
experience in a way that traditional journaling methods cannot.

Though intended to capture and preserve memories, other uses include fiction 
or non-fiction development, genealogical research, contact management and 
historical study.

Example output:

```bash
[matt@bodhi ~]$ memory
Welcome. You have 2 entries under management. Type 'help' for assistance.
memory> help
NAME:
   memory - A CLI tool to collect and browse the elements of human experience.

USAGE:
   memory [global options] command [command options] [arguments...]

VERSION:
   1.0

DESCRIPTION:
   memory is a tool to collect, browse and manage entries. Each entry
represents either an Event, Person, Place, Thing or Note. Each entry has a
unique name and entries can link to other entries using an entry name in
brackets, as in [Linked Entry]. When editing an entry, your favorite text
editor is loaded with a markdown file containing YAML frontmatter defining
the entry's attributes. Frontmatter is surrounded by three hyphens above
and below, and everything below the frontmatter is the entry's description,
which can be formatted with markdown and contain links. Files such as
images and documents can be attached to entries. Use the --help argument to
explore the commands available and use `memory command --help` to get
command-specific help. Commands can be used in interactive mode (at the
`memory>` prompt) or directly from your shell.

AUTHOR:
   Matt Wiseley <wiseley@gmail.com>

COMMANDS:
   add       adds a new entry
   delete    deletes an entry
   detail    displays details of an entry
   edit      edits an entry
   file      list file details and associated commands
   files     displays a list of attachments associated with an entry
   get       prints the editable form of an entry
   links     displays links to and from an entry
   ls        lists entries
   put       adds or updates an entry from a file
   rebuild   rebuilds the search index and internal database from entry files
   rename    renames an entry
   seeds     displays links to entries that don't exist yet
   tags      displays summary of entry tags
   timeline  displays a chronological list of dated entries
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --home value   directory path where data and settings are read from and saved to
   --help, -h     show help
   --version, -v  print the version
memory> _
```

Binary releases are provided for Linux and MacOS Darwin. If you're interested 
in a Windows version, let me know. Currently, one of the libraries I'm using 
for terminal control doesn't compile for Windows.

By default, preferences, entries and attachments are stored in ~/.memory. You 
can override this with the --home argument.

Feedback is welcome. I'm currently working on a web interface.