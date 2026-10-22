package settings

import (
	"os"
	"strings"

	"github.com/lulu58588/chongyun/internal/filememo"
)

// AccountDiscoveryOff is whether the agents' sign-ins are left unread
// (#1535): the user saved them off (NoAccountDiscovery), MAGPIE_NO_ACCOUNTS
// is set to anything but "", "0" or "false" for this process alone, or the
// user never turned them on at all — this build reads an agent's sign-in
// only when asked (AccountDiscovery, or MAGPIE_ACCOUNT_DISCOVERY for this
// process alone). A switch saved off is never turned on by the
// environment.
func AccountDiscoveryOff() bool {
	if noAccountsEnv() {
		return true
	}
	fileMu.RLock()
	defer fileMu.RUnlock()
	// the one setting, parsed once for each version of the file: it is
	// asked for each agent and each plugin every time the providers are
	// built, a whole parse of settings.json each (#2502)
	st, err := filememo.Read("settings:no-accounts", Path(), func(b []byte) (accountDiscovery, error) {
		s, err := parse(b)
		if err != nil {
			return accountDiscovery{}, err
		}
		return accountDiscovery{off: s.NoAccountDiscovery, on: s.AccountDiscovery}, nil
	})
	if err != nil {
		s := load()
		st = accountDiscovery{off: s.NoAccountDiscovery, on: s.AccountDiscovery}
	}
	if st.off {
		return true
	}
	if accountDiscoveryOnEnv() {
		return false
	}
	return !st.on
}

// accountDiscovery is what the settings say about reading the agents'
// sign-ins: off when the user saved them off, on when the user turned them
// on.
type accountDiscovery struct {
	off bool
	on  bool
}

// AccountDiscoveryOffIn is AccountDiscoveryOff with the settings already
// read, for a caller asking it many times in one look.
func AccountDiscoveryOffIn(s Settings) bool {
	if noAccountsEnv() {
		return true
	}
	if s.NoAccountDiscovery {
		return true
	}
	if accountDiscoveryOnEnv() {
		return false
	}
	return !s.AccountDiscovery
}

// accountDiscoveryOnEnv is whether the agents' sign-ins are read for this
// process alone: MAGPIE_ACCOUNT_DISCOVERY set to anything but "", "0" or
// "false".
func accountDiscoveryOnEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MAGPIE_ACCOUNT_DISCOVERY")))
	return v != "" && v != "0" && v != "false"
}

func noAccountsEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("MAGPIE_NO_ACCOUNTS")))
	return v != "" && v != "0" && v != "false"
}
