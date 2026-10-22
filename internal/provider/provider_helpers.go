package provider

// Small helpers several files of this package share: what a provider id may
// be called, how two providers are told apart, and how another app's
// SQLite database is opened without writing to it.

import (
	"database/sql"
	"encoding/json"
	"errors"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lulu58588/chongyun/internal/plugin"

	// opens another app's SQLite database (openReadOnly below)
	_ "github.com/lulu58588/chongyun/internal/sqlitedb"
)

// freeID is id, or id-2, id-3… whichever no provider has.
func freeID(id string) string {
	taken := map[string]bool{"magpie": true}
	for _, id := range accountIDs {
		taken[id] = true
	}
	for _, pp := range plugin.Cached() { // signed in or not, as a built-in's
		taken[PluginID(pp.ID)] = true
	}
	for _, p := range All() {
		taken[p.ID] = true
	}
	for _, p := range load().Providers {
		taken[p.ID] = true
	}
	for n, try := 2, id; ; n++ {
		if !taken[try] {
			return try
		}
		try = id + "-" + itoa(n)
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// sameProvider: the same key, host, and headers are the same account; a is
// the provider magpie has, which may hold the key among its others.
func sameProvider(a, b Provider) bool {
	if !maps.Equal(cleanHeaders(a.Headers), cleanHeaders(b.Headers)) {
		return false
	}
	has := a.Key == b.Key
	for _, k := range a.Keys {
		has = has || k.Key == b.Key
	}
	if !has {
		return false
	}
	hosts := map[string]bool{}
	for _, u := range []string{a.Chat, a.Responses, a.Anthropic, a.Gemini, a.Decide} {
		if h := hostOf(u); h != "" {
			hosts[h] = true
		}
	}
	for _, u := range []string{b.Chat, b.Responses, b.Anthropic, b.Gemini} {
		if hosts[hostOf(u)] {
			return true
		}
	}
	return false
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// sqlitePath is a file's path as a SQLite URI wants it: slashed, and on
// Windows with a slash before the drive (file:///C:/…) — without it the
// drive letter is read as a host, and the open fails.
func sqlitePath(path string) string {
	p := filepath.ToSlash(path)
	if len(p) >= 2 && p[1] == ':' {
		p = "/" + p
	}
	return p
}

// openReadOnly opens another app's SQLite database without writing to it,
// waiting out the app's own writes.
func openReadOnly(path string) (*sql.DB, error) {
	u := url.URL{Scheme: "file", Path: sqlitePath(path), RawQuery: "mode=ro&_pragma=busy_timeout(3000)"}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, errors.New("can't open " + path + ": " + err.Error())
	}
	return db, nil
}

// OpenReadOnly opens another app's SQLite database without writing to it.
func OpenReadOnly(path string) (*sql.DB, error) { return openReadOnly(path) }
