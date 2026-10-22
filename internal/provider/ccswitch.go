package provider

// CC Switch's own folder is only ever read: where it keeps the skills it
// installs, and where each of them came from. Nothing of a sign-in is taken
// from it.

import (
	"os"
	"path/filepath"
)

func ccSwitchPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".cc-switch")
	if db := filepath.Join(dir, "cc-switch.db"); fileExists(db) {
		return db
	}
	if js := filepath.Join(dir, "config.json"); fileExists(js) {
		return js
	}
	return filepath.Join(dir, "cc-switch.db")
}

// CCSwitchSkillsDir is the folder CC Switch keeps the skills it installs in.
func CCSwitchSkillsDir() string { return filepath.Join(filepath.Dir(ccSwitchPath()), "skills") }

// CCSwitchSkillOrigin is the GitHub repository (owner/name) and branch CC
// Switch installed the skill in its folder dir from, as its database
// records it.
func CCSwitchSkillOrigin(dir string) (repo, branch string, ok bool) {
	p := ccSwitchPath()
	if filepath.Ext(p) != ".db" || !fileExists(p) {
		return "", "", false
	}
	db, err := openReadOnly(p)
	if err != nil {
		return "", "", false
	}
	defer db.Close()
	rows, err := db.Query(`SELECT directory, COALESCE(repo_owner,''), COALESCE(repo_name,''), COALESCE(repo_branch,'') FROM skills`)
	if err != nil {
		return "", "", false
	}
	defer rows.Close()
	for rows.Next() {
		var d, owner, name, ref string
		if rows.Scan(&d, &owner, &name, &ref) != nil || owner == "" || name == "" {
			continue
		}
		if d == dir || filepath.Base(filepath.FromSlash(d)) == dir {
			return owner + "/" + name, ref, true
		}
	}
	return "", "", false
}
