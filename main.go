package main

import "github.com/dombo/privnote/cmd"

/*
privnote
	--destructs (immediately|1hr|24hr|7d|1w|30d|1m) // automatically destruct after this time
	--password "SomePassword" // require a password to open the note
	--notification // receive an email notification at this address when the note is read
	--do-not-confirm // do not prompt the reader to confirm before showing and destroying the note

Configure by flags, ENV_VAR, config file - in descending order of precedence
*/

func main() {
	cmd.Execute()
}