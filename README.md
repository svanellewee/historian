# Historian

This aims to be a more powerful way of storing your bash (zsh?) history.

Current goals:

- Per directory command history
  - A global command history dump still needs to be implemented
  - A dump for specific days needs to be added.
- A transferrable history database that can stay with you long after you've moved laptops.
- Something that will not get confused by multiple tmux/terminal sessions.
- Allow annotation additions to history entries (ie These commands were part of JIRA ticket this or that)
- Searchable using a proper search system (Bleve?)
  - I'd like to get rid of `history | grep ...`

## Installation

```sh
go get github.com/svanellewee/historian
```

## How to use this

Add something like this to your bashrc

```sh
# Taken from https://www.digitalocean.com/community/tutorials/how-to-use-bash-history-commands-and-expansions-on-a-linux-vps
HISTSIZE=5000
HISTFILESIZE=10000
function history-store() {
    echo "$(history 1)" >> "${OUTPUT_FILE}"
    historian insert "$(history 1)"
}
alias hlist="historian last"

# The secret sauce that makes things work...
export PROMPT_COMMAND="history-store"
```

### Last

To see the last 10 commands used inside *the current directory*, run 

```sh
historian last 10
```

### Today

To see all the commands you ran and at what times and where for `today` just ask:

```sh
historian today
```

This will give you a sorted list, so you can see what you did. Might be useful for timesheet-y type applications. (Mmmm perhaps I need a `sprint` command as well)

### Search

Go's regex are used here. Examples:

- Where did you mispell `PATH`?
```sh
historian search PAHT 
```

- If you write nothing it matches all...probably a bad idea

```sh
historian search 
```

- You need to find all calls with `bind` and/or `shutdown`

```sh
historian search 'bind|shutdown' 
```

- You wrote this dumb history app called `historian` and now need to write examples for the docs:

```sh
historian search 'historian search'
```

# Why write another bash history?

My motivation is purely to learn go, and to have something useful out of it. If you end up using it too please drop me a line, I'd love to hear your experience or if there's any improvements in useability or code I can make. I'll be making updates as I go.
