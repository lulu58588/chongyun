package provider

// Where Codex keeps its own sign-in, and how starling's own ChatGPT sign-in
// is put there.
//
// Codex 0.162's config.toml says where (codex-rs/login/src/auth/storage.rs,
// codex-rs/config/src/types.rs): cli_auth_credentials_store "file" (the
// default) is auth.json in CODEX_HOME; "keyring" is the system keychain
// alone; "auto" is the keychain, else auth.json; "ephemeral" is Codex's
// memory. With [features] secret_auth_storage on (the default on Windows),
// keyring and auto mean Codex's encrypted auth storage instead of the
// keychain item.
//
// Nothing here reads a sign-in of Codex's: starling writes only the sign-in
// of its own OAuth flow, so that Codex itself goes by the account starling
// signed in to. Which store Codex keeps its sign-in in is told by its
// config.toml, a setting file, not a credential.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/lulu58588/starling/internal/filememo"
	"github.com/lulu58588/starling/internal/proc"
)

// codexKeychainService is the keychain item Codex keeps its sign-in in.
const codexKeychainService = "Codex Auth"

// secretAuthStorageDefault is Codex's default for [features]
// secret_auth_storage: cfg!(windows) (codex-rs/features/src/lib.rs)
var secretAuthStorageDefault = runtime.GOOS == "windows"

// keychainWait bounds a `security` run that puts up no prompt: a write.
var keychainWait = 10 * time.Second

// codexKeychainAccount is Codex's compute_store_key: the home made
// canonical where it can be (else as given), SHA-256, its first 16 hex
// digits after "cli|".
func codexKeychainAccount(home string) string {
	p := home
	if abs, err := filepath.Abs(home); err == nil {
		if real, err := filepath.EvalSymlinks(abs); err == nil {
			p = real
		}
	}
	sum := sha256.Sum256([]byte(p))
	return "cli|" + hex.EncodeToString(sum[:])[:16]
}

// codexStoreSetting is what config.toml in a Codex home says of where
// the sign-in is kept.
type codexStoreSetting struct {
	Mode    string // file, keyring, auto or ephemeral
	Secrets bool   // [features] secret_auth_storage
}

// readCodexStoreSetting reads config.toml as Codex does: no file or no
// key is "file". A file that doesn't parse, or a value Codex doesn't
// know (it would refuse to start), is an error: where the sign-in is
// isn't known.
func readCodexStoreSetting(home string) (codexStoreSetting, error) {
	path := filepath.Join(home, "config.toml")
	s, err := filememo.Read("codex credential store", path, func(b []byte) (codexStoreSetting, error) {
		var c struct {
			Store    string         `toml:"cli_auth_credentials_store"`
			Features map[string]any `toml:"features"`
		}
		if err := toml.Unmarshal(b, &c); err != nil {
			return codexStoreSetting{}, fmt.Errorf("%s doesn't read as TOML", path)
		}
		s := codexStoreSetting{Mode: c.Store, Secrets: secretAuthStorageDefault}
		if v, ok := c.Features["secret_auth_storage"].(bool); ok {
			s.Secrets = v
		}
		return s, nil
	})
	if errors.Is(err, fs.ErrNotExist) {
		return codexStoreSetting{Mode: "file", Secrets: secretAuthStorageDefault}, nil
	}
	if err != nil {
		return codexStoreSetting{}, err
	}
	switch s.Mode {
	case "":
		s.Mode = "file"
	case "file", "keyring", "auto", "ephemeral":
	default:
		return codexStoreSetting{}, fmt.Errorf("cli_auth_credentials_store = %q in %s isn't one Codex knows", s.Mode, path)
	}
	return s, nil
}

// keychainItem: the store is the keychain item starling writes to.
func (s codexStoreSetting) keychainItem() bool {
	return (s.Mode == "keyring" || s.Mode == "auto") && !s.Secrets && runtime.GOOS == "darwin"
}

// where names the store keyring and auto mean here.
func (s codexStoreSetting) where() string {
	switch {
	case s.Secrets:
		return "Codex's encrypted auth storage (features.secret_auth_storage)"
	case runtime.GOOS == "darwin":
		return "the macOS keychain"
	}
	return "the system keychain"
}

// codexStoreError: Codex keeps its sign-in where starling cannot write it,
// so the sign-in starling made is not the one Codex goes by.
type codexStoreError struct {
	Setting codexStoreSetting
	Config  string
	Cause   error // nil: a store starling doesn't write to on this system
}

func (e *codexStoreError) Error() string {
	if e.Setting.Mode == "ephemeral" {
		return fmt.Sprintf("Codex keeps its sign-in in its own memory (cli_auth_credentials_store = \"ephemeral\" in %s), which no other program can write", e.Config)
	}
	if e.Cause != nil {
		return fmt.Sprintf("Codex keeps its sign-in in %s (cli_auth_credentials_store = %q in %s), and %v", e.Setting.where(), e.Setting.Mode, e.Config, e.Cause)
	}
	return fmt.Sprintf("Codex keeps its sign-in in %s (cli_auth_credentials_store = %q in %s), which starling can't write", e.Setting.where(), e.Setting.Mode, e.Config)
}

// codexStore is where a Codex sign-in is written.
type codexStore struct {
	home     string
	keychain bool
}

func (s codexStore) file() string { return filepath.Join(s.home, "auth.json") }

// write puts a sign-in where Codex looks for it: rotated tokens, every
// other field as it was. The keychain item takes it on one line, as Codex
// writes it; `security -w` prints one with a newline as hex (#70).
func (s codexStore) write(b []byte) error {
	if !s.keychain {
		// the user's own sign-in, which Codex CLI and the Codex app
		// read: written beside it, flushed and renamed in, never
		// truncated in place (#2239, #1505)
		return writePrivate(s.file(), append(bytes.TrimSpace(b), '\n'))
	}
	var one bytes.Buffer
	if err := json.Compact(&one, b); err != nil {
		return errors.New("the Codex sign-in to save isn't JSON")
	}
	return codexKeychainPut(s.home, one.Bytes())
}

// codexKeychainPut writes the sign-in into the keychain item Codex keeps
// it in, on a Mac.
func codexKeychainPut(home string, secret []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), keychainWait)
	defer cancel()
	account := codexKeychainAccount(home)
	if err := proc.CommandContext(ctx, "security", "add-generic-password", "-U", "-s", codexKeychainService, "-a", account, "-w", string(secret)).Run(); err != nil {
		return errors.New("the keychain didn't take the sign-in")
	}
	return nil
}

// saveCodexLogin makes b the sign-in Codex goes by (a switch, or the
// first sign-in in starling), in the store its config.toml names. It
// writes only: no sign-in of Codex's is read, so none is taken for
// absent, and none is written over by accident — starling's own OAuth
// sign-in is what it puts there.
func saveCodexLogin(home string, b []byte) error {
	s, err := readCodexStoreSetting(home)
	if err != nil {
		return fmt.Errorf("where Codex keeps its sign-in isn't known (%v); nothing was changed", err)
	}
	st := codexStore{home: home}
	cfg := filepath.Join(home, "config.toml")
	switch {
	case s.Mode == "ephemeral":
		return &codexStoreError{Setting: s, Config: cfg}
	case s.keychainItem():
		st.keychain = true
		werr := st.write(b)
		if werr == nil || s.Mode == "keyring" {
			return werr
		}
		// auto, and the keychain wouldn't take it: Codex saves to the
		// file then, and loads it
		st.keychain = false
	case s.Mode == "keyring":
		return &codexStoreError{Setting: s, Config: cfg}
	case s.Mode == "auto":
		// a store starling can't write to: only a Codex on its file
		// fallback goes by auth.json
		if _, err := os.Stat(st.file()); err != nil {
			return &codexStoreError{Setting: s, Config: cfg}
		}
	}
	return st.write(b)
}
