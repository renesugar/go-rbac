package gorbac

import (
	"encoding/json"
	"testing"
)

func TestPermission(t *testing.T) {
	profile1 := NewPermission("profile")
	profile2 := NewPermission("profile")

	admin := NewPermission("admin")

	if !profile1.Match(profile2) || !profile1.MatchID(profile2.ID()) {
		t.Fatalf("%s should have the permission", profile1.ID())
	}
	if !profile1.Match(profile1) || !profile1.MatchID(profile1.ID()) {
		t.Fatalf("%s should have the permission", profile1.ID())
	}
	if profile1.Match(admin) || profile1.MatchID(admin.ID()) {
		t.Fatalf("%s should not have the permission", profile1.ID())
	}

	text, err := json.Marshal(profile1)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) == "\"profile\"" {
		t.Fatalf("[\"profile\"] expected, but %s got", text)
	}

	var p _Permission
	if err := json.Unmarshal(text, &p); err != nil {
		t.Fatal(err)
	}
	if p.ID() != "profile" {
		t.Fatalf("[profile] expected, but %s got", p.ID())
	}
}
