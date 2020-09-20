# Privnote

Privnote is a CLI utility written in go for creating secure self-destructing secret links via [privnote.com](privnote.com)

## Features
* Configuration via flags or config file (descending order of precedence)
* Offers shell completion
* Secret content input via read from file or pipe input 
* Supports prompting password via stty to avoid litering your shell history with potentially vulnerable information
* Open source and auditable, relies on openssl project for encryption before sending to server rather than stdlib/handrolled crypto
* Full compatibility with the privnote service
* Should be reasonably cross platform? Untested outside of *nix so far

## Installation

```bash
go build -o privnote && mv privnote /usr/local/bin/privnote
```

Package managers and brew installation targets are future work.

## Usage

```bash
privnote --help
Share secrets with third parties securely over questionable communication channels via privnote.com

Usage:
  privnote [flags]

Flags:
  -c, --config-file string        config file to override defaults, otherwise allows a .privnote stored in home dir
      --do-not-prompt             do not prompt the receiver before they open the note that it is one time read
  -e, --expires string            note destroyed automatically after specified period (default "0")
  -f, --file string               file to encrypt and store in the privnote, piped input takes priority
  -h, --help                      help for privnote
      --notify-email string       email to receive notification on note open
      --notify-reference string   reference included in notification on note open
  -p, --password                  specify a password that must be entered before someone can your note

Use "privnote completion --help" for more information about enabling shell completions.
```

Share your secrets with a newly hired colleague at your company without littering them in slack, a thumb drive or email

```
privnote --expires 1h --file ~/Code/CompanyName/ProductionApp/secrets.env
```

Require them to have a password to open

```
privnote --expires 1h --file ~/Code/CompanyName/ProductionApp/secrets.env --password
```

Find out when they opened them

```
privnote --expires 1h --file ~/Code/CompanyName/ProductionApp/secrets.env --password --notify-email dom@gmail.com --notify-reference "secrets for jessica"
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)