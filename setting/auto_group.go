package setting

import (
	"github.com/QuantumNous/new-api/common"
)

var autoGroups = []string{
	common.TokenGroupDefault,
}

var DefaultUseAutoGroup = false

func ContainsAutoGroup(group string) bool {
	group = common.NormalizeTokenGroup(group)
	for _, autoGroup := range autoGroups {
		if common.NormalizeTokenGroup(autoGroup) == group {
			return true
		}
	}
	return false
}

func UpdateAutoGroupsByJsonString(jsonString string) error {
	updatedGroups := make([]string, 0)
	if err := common.Unmarshal([]byte(jsonString), &updatedGroups); err != nil {
		return err
	}
	autoGroups = make([]string, 0, len(updatedGroups))
	seen := make(map[string]bool)
	for _, group := range updatedGroups {
		normalized := common.NormalizeTokenGroup(group)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		autoGroups = append(autoGroups, normalized)
	}
	return nil
}

func AutoGroups2JsonString() string {
	jsonBytes, err := common.Marshal(GetAutoGroups())
	if err != nil {
		return "[]"
	}
	return string(jsonBytes)
}

func GetAutoGroups() []string {
	groups := make([]string, 0, len(autoGroups))
	seen := make(map[string]bool)
	for _, group := range autoGroups {
		normalized := common.NormalizeTokenGroup(group)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		groups = append(groups, normalized)
	}
	return groups
}
