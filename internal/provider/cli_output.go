package provider

// What an agent's own command line prints: the escape sequences it writes
// around its output, and the sign-in address it shows.

import "regexp"

// ansi is the escape sequence a command line writes around its output.
var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)

// cursorLoginURL is the sign-in address a command line prints, as it does
// when it opens a browser to sign in.
var cursorLoginURL = regexp.MustCompile(`https://\S+`)
