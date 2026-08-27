package ratio_setting

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/config"
	"github.com/QuantumNous/new-api/types"
)

var defaultGroupRatio = map[string]float64{
	common.TokenGroupDefault: 1,
}

var groupRatioMap = types.NewRWMap[string, float64]()

var defaultGroupGroupRatio = map[string]map[string]float64{}

var groupGroupRatioMap = types.NewRWMap[string, map[string]float64]()

var defaultGroupSpecialUsableGroup = map[string]map[string]string{}

type GroupRatioSetting struct {
	GroupRatio              *types.RWMap[string, float64]            `json:"group_ratio"`
	GroupGroupRatio         *types.RWMap[string, map[string]float64] `json:"group_group_ratio"`
	GroupSpecialUsableGroup *types.RWMap[string, map[string]string]  `json:"group_special_usable_group"`
}

var groupRatioSetting GroupRatioSetting

func init() {
	groupSpecialUsableGroup := types.NewRWMap[string, map[string]string]()
	groupSpecialUsableGroup.AddAll(defaultGroupSpecialUsableGroup)

	groupRatioMap.AddAll(defaultGroupRatio)
	groupGroupRatioMap.AddAll(defaultGroupGroupRatio)

	groupRatioSetting = GroupRatioSetting{
		GroupSpecialUsableGroup: groupSpecialUsableGroup,
		GroupRatio:              groupRatioMap,
		GroupGroupRatio:         groupGroupRatioMap,
	}

	config.GlobalConfig.Register("group_ratio_setting", &groupRatioSetting)
}

func GetGroupRatioSetting() *GroupRatioSetting {
	if groupRatioSetting.GroupSpecialUsableGroup == nil {
		groupRatioSetting.GroupSpecialUsableGroup = types.NewRWMap[string, map[string]string]()
		groupRatioSetting.GroupSpecialUsableGroup.AddAll(defaultGroupSpecialUsableGroup)
	}
	return &groupRatioSetting
}

func GetGroupRatioCopy() map[string]float64 {
	groupRatio := make(map[string]float64)
	for name, ratio := range groupRatioMap.ReadAll() {
		normalized := common.NormalizeTokenGroup(name)
		if normalized == "" {
			continue
		}
		groupRatio[normalized] = ratio
	}
	return groupRatio
}

func ContainsGroupRatio(name string) bool {
	for _, candidate := range common.TokenGroupAliasCandidates(name) {
		if _, ok := groupRatioMap.Get(candidate); ok {
			return true
		}
	}
	return false
}

func GroupRatio2JSONString() string {
	jsonBytes, err := common.Marshal(GetGroupRatioCopy())
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func UpdateGroupRatioByJSONString(jsonStr string) error {
	updatedGroupRatio := make(map[string]float64)
	if err := common.Unmarshal([]byte(jsonStr), &updatedGroupRatio); err != nil {
		return err
	}
	groupRatioMap.Clear()
	for name, ratio := range updatedGroupRatio {
		normalized := common.NormalizeTokenGroup(name)
		if normalized == "" {
			continue
		}
		groupRatioMap.Set(normalized, ratio)
	}
	return nil
}

func GetGroupRatio(name string) float64 {
	normalized := common.NormalizeTokenGroup(name)
	for _, candidate := range common.TokenGroupAliasCandidates(normalized) {
		ratio, ok := groupRatioMap.Get(candidate)
		if ok {
			return ratio
		}
	}
	common.SysLog("group ratio not found: " + normalized)
	return 1
}

func GetGroupGroupRatio(userGroup, usingGroup string) (float64, bool) {
	for _, userCandidate := range common.TokenGroupAliasCandidates(userGroup) {
		gp, ok := groupGroupRatioMap.Get(userCandidate)
		if !ok {
			continue
		}
		for _, usingCandidate := range common.TokenGroupAliasCandidates(usingGroup) {
			ratio, ok := gp[usingCandidate]
			if ok {
				return ratio, true
			}
		}
	}
	return -1, false
}

func GroupGroupRatio2JSONString() string {
	jsonBytes, err := common.Marshal(getGroupGroupRatioCopy())
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func UpdateGroupGroupRatioByJSONString(jsonStr string) error {
	updatedGroupGroupRatio := make(map[string]map[string]float64)
	if err := common.Unmarshal([]byte(jsonStr), &updatedGroupGroupRatio); err != nil {
		return err
	}
	groupGroupRatioMap.Clear()
	for userGroup, usingGroups := range updatedGroupGroupRatio {
		normalizedUserGroup := common.NormalizeTokenGroup(userGroup)
		if normalizedUserGroup == "" {
			continue
		}
		normalizedUsingGroups := make(map[string]float64)
		for usingGroup, ratio := range usingGroups {
			normalizedUsingGroup := common.NormalizeTokenGroup(usingGroup)
			if normalizedUsingGroup == "" {
				continue
			}
			normalizedUsingGroups[normalizedUsingGroup] = ratio
		}
		groupGroupRatioMap.Set(normalizedUserGroup, normalizedUsingGroups)
	}
	return nil
}

func CheckGroupRatio(jsonStr string) error {
	checkGroupRatio := make(map[string]float64)
	err := common.Unmarshal([]byte(jsonStr), &checkGroupRatio)
	if err != nil {
		return err
	}
	for name, ratio := range checkGroupRatio {
		if ratio < 0 {
			return errors.New("group ratio must be not less than 0: " + name)
		}
	}
	return nil
}

func getGroupGroupRatioCopy() map[string]map[string]float64 {
	groupGroupRatio := make(map[string]map[string]float64)
	for userGroup, usingGroups := range groupGroupRatioMap.ReadAll() {
		normalizedUserGroup := common.NormalizeTokenGroup(userGroup)
		if normalizedUserGroup == "" {
			continue
		}
		if _, ok := groupGroupRatio[normalizedUserGroup]; !ok {
			groupGroupRatio[normalizedUserGroup] = make(map[string]float64)
		}
		for usingGroup, ratio := range usingGroups {
			normalizedUsingGroup := common.NormalizeTokenGroup(usingGroup)
			if normalizedUsingGroup == "" {
				continue
			}
			groupGroupRatio[normalizedUserGroup][normalizedUsingGroup] = ratio
		}
	}
	return groupGroupRatio
}
