# Memory

This application is in early development. See the commit history for progress
updates.

Memory is a command line application that captures, stores and manages the data
that describes human experience. Information is organized into people, places,
things and events. A network of relationships between them relays human
experience in a way that traditional journaling methods cannot.

Though intended to capture and preserve memories, other uses include fiction 
or non-fiction development, genealogical research, contact management and 
historical study.

Example commands:

```bash
# Print usage help
$ memory 

# Add a Note (-n == --name, -d == --description, -t == --tags)
$ memory add-note -n "My first note" -d "Long description" -t "tag1,tag2"

# List 10 most recent entries
$ memory ls

# List 20 most recent entries
$ memory ls --limit 20

# List Notes sorted by Name (-t == --types)
$ memory ls -t Note --sort-name

# List Notes with a certain tag (tags are case-insensitive)
$ memory ls -t Note --tag todo

```
