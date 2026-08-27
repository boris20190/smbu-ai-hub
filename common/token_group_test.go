package common

import (
	"reflect"
	"testing"
)

func TestNormalizeTokenGroupPreservesDefaultAndCustomGroups(t *testing.T) {
	tests := []struct {
		name  string
		group string
		want  string
	}{
		{name: "empty inherits user group", group: "", want: ""},
		{name: "blank inherits user group", group: "  ", want: ""},
		{name: "default stays default", group: TokenGroupDefault, want: TokenGroupDefault},
		{name: "custom group", group: "custom", want: "custom"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizeTokenGroup(test.group); got != test.want {
				t.Fatalf("NormalizeTokenGroup(%q) = %q, want %q", test.group, got, test.want)
			}
		})
	}
}

func TestNormalizeTokenGroupOrDefaultUsesDefault(t *testing.T) {
	if got := NormalizeTokenGroupOrDefault(""); got != TokenGroupDefault {
		t.Fatalf("NormalizeTokenGroupOrDefault empty = %q, want %q", got, TokenGroupDefault)
	}
	if got := NormalizeTokenGroupOrDefault("custom"); got != "custom" {
		t.Fatalf("NormalizeTokenGroupOrDefault custom = %q, want custom", got)
	}
}

func TestTokenGroupDoesNotPartitionModelsByName(t *testing.T) {
	if !TokenGroupAllowsModel(TokenGroupDefault, "gpt-5") {
		t.Fatal("default group should allow models")
	}
	if !TokenGroupAllowsModel(TokenGroupDefault, "custom-model") {
		t.Fatal("default group should no longer partition models by naming convention")
	}
	if got := TokenGroupForModel(TokenGroupDefault, "custom-model"); got != TokenGroupDefault {
		t.Fatalf("TokenGroupForModel(default, custom model) = %q, want %q", got, TokenGroupDefault)
	}
}

func TestTokenGroupLookupCandidatesAreExact(t *testing.T) {
	tests := []struct {
		name  string
		group string
		want  []string
	}{
		{name: "empty", group: "", want: []string{}},
		{name: "default", group: TokenGroupDefault, want: []string{TokenGroupDefault}},
		{name: "custom", group: "custom", want: []string{"custom"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := TokenGroupLookupCandidatesForModel(test.group, "gpt-5")
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("TokenGroupLookupCandidatesForModel(%q) = %#v, want %#v", test.group, got, test.want)
			}
		})
	}
}
