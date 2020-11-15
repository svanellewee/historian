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