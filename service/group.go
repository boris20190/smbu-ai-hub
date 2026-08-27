package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

func GetUserUsableGroups(userGroup string) map[string]string {
	groupsCopy := setting.GetUserUsableGroupsCopy()
	if userGroup != "" {
		specialSettings, b := getGroupSpecialUsableGroups(userGroup)
		if b {
			// 处理特殊可用分组
			for specialGroup, desc := range specialSettings {
				if strings.HasPrefix(specialGroup, "-:") {
					// 移除分组
					groupToRemove := strings.TrimPrefix(specialGroup, "-:")
					normalized := common.NormalizeTokenGroup(groupToRemove)
					if normalized != "" {
						delete(groupsCopy, normalized)
					}
				} else if strings.HasPrefix(specialGroup, "+:") {
					// 添加分组
					groupToAdd := strings.TrimPrefix(specialGroup, "+:")
					normalized := common.NormalizeTokenGroup(groupToAdd)
					if normalized == "" {
						continue
					}
					if normalized == common.TokenGroupDefault {
						desc = common.TokenGroupDisplayName(normalized)
					}
					groupsCopy[normalized] = desc
				} else {
					// 直接添加分组
					normalized := common.NormalizeTokenGroup(specialGroup)
					if normalized == "" {
						continue
					}
					if normalized == common.TokenGroupDefault {
						desc = common.TokenGroupDisplayName(normalized)
					}
					groupsCopy[normalized] = desc
				}
			}
		}
		// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
		normalizedUserGroup := common.NormalizeTokenGroup(userGroup)
		if _, ok := groupsCopy[normalizedUserGroup]; !ok {
			groupsCopy[normalizedUserGroup] = "用户分组"
		}
	}
	return groupsCopy
}

func getGroupSpecialUsableGroups(userGroup string) (map[string]string, bool) {
	for _, candidate := range common.TokenGroupAliasCandidates(userGroup) {
		specialSettings, ok := ratio_setting.GetGroupRatioSetting().GroupSpecialUsableGroup.Get(candidate)
		if ok {
			return specialSettings, true
		}
	}
	return nil, false
}

func GroupInUserUsableGroups(userGroup, groupName string) bool {
	_, ok := GetUserUsableGroups(userGroup)[common.NormalizeTokenGroup(groupName)]
	return ok
}

// GetUserAutoGroup 根据用户分组获取自动分组设置
func GetUserAutoGroup(userGroup string) []string {
	groups := GetUserUsableGroups(userGroup)
	autoGroups := make([]string, 0)
	for _, group := range setting.GetAutoGroups() {
		if _, ok := groups[group]; ok {
			autoGroups = append(autoGroups, group)
		}
	}
	return autoGroups
}

// GetUserGroupRatio 获取用户使用某个分组的倍率
// userGroup 用户分组
// group 需要获取倍率的分组
func GetUserGroupRatio(userGroup, group string) float64 {
	ratio, ok := ratio_setting.GetGroupGroupRatio(userGroup, group)
	if ok {
		return ratio
	}
	return ratio_setting.GetGroupRatio(group)
}
