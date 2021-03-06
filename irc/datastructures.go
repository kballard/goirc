package irc

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type User struct {
	// Nick, User, and Host will only be present if the user is of the form
	// nick[!user]@host Notably, a server host sender will have all three fields
	// as the empty string.
	Nick, User, Host string
	Raw              string
}

var userRegex = regexp.MustCompile("^([a-zA-Z[-`{-}][a-zA-Z0-9[-`{-}\\-]+)(?:!([^@]+))@(.+)$")

// should be called with a string like nick!user@host
func parseUser(raw string) User {
	user := User{Raw: raw}
	if matches := userRegex.FindStringSubmatch(raw); matches != nil {
		user.Nick = matches[1]
		user.User = matches[2]
		user.Host = matches[3]
	}
	return user
}

// Returns the user's nickname, or the raw string if there is no nickname.
func (u User) String() string {
	if u.Nick != "" {
		return u.Nick
	}
	return u.Raw
}

// Returns the user@host string, or just host if no user, or "" if this isn't a
// user.
func (u User) Ident() string {
	if u.Host != "" && u.User != "" {
		return fmt.Sprintf("%s@%s", u.User, u.Host)
	} else if u.Host != "" {
		return u.Host
	}
	return ""
}

type Line struct {
	Src     User
	Command string
	Args    []string
	Raw     string
	Time    time.Time

	// Dst is only filled in for the special commands such as ACTION, CTCP, and
	// CTCPReply. It denotes the target the PRIVMSG/NOTICE was sent to.
	Dst string

	me User
}

func parseLine(input string) (line Line) {
	line.Raw = input
	line.Time = time.Now()
	// quick sanity check
	if len(input) == 0 || input[0] == ' ' {
		return
	}
	// split input, first into "prefix :suffix", and then tokenize prefix
	comps := strings.SplitN(input, " :", 2)
	input = comps[0]
	words := strings.FieldsFunc(input, func(r rune) bool { return r == ' ' })
	if len(words) == 0 {
		// where's my prefix/command?
		return
	} else if words[0][0] == ':' {
		// it has the expected sender prefix
		line.Src = parseUser(words[0][1:])
		words = words[1:]
	}
	if len(words) == 0 {
		// where's my command?
		return
	}
	line.Command = words[0]
	words = words[1:]
	if len(comps) > 1 {
		words = append(words, comps[1])
	}
	line.Args = words
	return
}

// SrcIsMe returns if the Src is the same as Me.
func (l *Line) SrcIsMe() bool {
	return l.Src.Nick == l.me.Nick
}
