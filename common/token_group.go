package common

import "strings"

const (
	TokenGroupDefault = "default"
	TokenGroupAuto    = "auto"
)

func NormalizeTokenGroup(group string) string {
	return strings.TrimSpace(group)
}

func NormalizeTokenGroupOrDefault(group string) string {
	normalized := NormalizeTokenGroup(group)
	if normalized == "" {
		return TokenGroupDefault
	}
	return normalized
}

func TokenGroupDisplayName(group string) string {
	switch NormalizeTokenGroup(group) {
	case TokenGroupDefault:
		return "默认"
	default:
		return strings.TrimSpace(group)
	}
}

func TokenGroupAllowsModel(group string, modelName string) bool {
	return true
}

func TokenGroupForModel(group string, modelName string) string {
	return NormalizeTokenGroup(group)
}

func TokenGroupLookupCandidates(group string) []string {
	normalized := NormalizeTokenGroup(group)
	if normalized == "" {
		return []string{}
	}
	return []string{normalized}
}

func TokenGroupAliasCandidates(group string) []string {
	normalized := NormalizeTokenGroup(group)
	if normalized == "" {
		return []string{}
	}
	return []string{normalized}
}

func TokenGroupLookupCandidatesForModel(group string, modelName string) []string {
	normalized := NormalizeTokenGroup(group)
	if !TokenGroupAllowsModel(normalized, modelName) {
		return []string{}
	}
	return TokenGroupLookupCandidates(normalized)
}
