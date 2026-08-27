package setting

import (
	"sync"

	"github.com/QuantumNous/new-api/common"
)

var userUsableGroups = map[string]string{
	common.TokenGroupDefault: "默认",
}
var userUsableGroupsMutex sync.RWMutex

func normalizeUserUsableGroups(groups map[string]string) map[string]string {
	normalizedGroups := make(map[string]string)
	for group, desc := range groups {
		normalized := common.NormalizeTokenGroup(group)
		if normalized == "" {
			continue
		}
		switch normalized {
		case common.TokenGroupDefault:
			normalizedGroups[normalized] = common.TokenGroupDisplayName(normalized)
		default:
			normalizedGroups[normalized] = desc
		}
	}
	return normalizedGroups
}

func GetUserUsableGroupsCopy() map[string]string {
	userUsableGroupsMutex.RLock()
	defer userUsableGroupsMutex.RUnlock()

	normalizedGroups := normalizeUserUsableGroups(userUsableGroups)
	copyUserUsableGroups := make(map[string]string)
	for k, v := range normalizedGroups {
		copyUserUsableGroups[k] = v
	}
	return copyUserUsableGroups
}

func UserUsableGroups2JSONString() string {
	userUsableGroupsMutex.RLock()
	defer userUsableGroupsMutex.RUnlock()

	jsonBytes, err := common.Marshal(normalizeUserUsableGroups(userUsableGroups))
	if err != nil {
		common.SysLog("error marshalling user groups: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateUserUsableGroupsByJSONString(jsonStr string) error {
	userUsableGroupsMutex.Lock()
	defer userUsableGroupsMutex.Unlock()

	updatedGroups := make(map[string]string)
	if err := common.Unmarshal([]byte(jsonStr), &updatedGroups); err != nil {
		return err
	}
	userUsableGroups = normalizeUserUsableGroups(updatedGroups)
	return nil
}

func GetUsableGroupDescription(groupName string) string {
	userUsableGroupsMutex.RLock()
	defer userUsableGroupsMutex.RUnlock()

	normalized := common.NormalizeTokenGroup(groupName)
	if desc, ok := userUsableGroups[normalized]; ok {
		return desc
	}
	return common.TokenGroupDisplayName(normalized)
}
