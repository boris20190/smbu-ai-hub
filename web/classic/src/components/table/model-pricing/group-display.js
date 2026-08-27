/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

export const MARKETPLACE_GROUP_LABELS = {
  offline: '非联网',
  online: '联网',
};

export const MARKETPLACE_NON_WEB_GROUP = 'offline';
export const MARKETPLACE_WEB_GROUP = 'online';

const MARKETPLACE_GROUP_ALIASES = {
  default: MARKETPLACE_NON_WEB_GROUP,
  vip: MARKETPLACE_WEB_GROUP,
};

export const normalizeMarketplaceGroup = (group) => {
  const trimmed = typeof group === 'string' ? group.trim() : '';
  return MARKETPLACE_GROUP_ALIASES[trimmed] || trimmed;
};

export const isMarketplaceOnlineModelName = (modelName) => {
  if (typeof modelName !== 'string') return false;
  return modelName.trim().toLowerCase().endsWith('online');
};

export const isMarketplaceOnlineModel = (model) =>
  isMarketplaceOnlineModelName(
    typeof model === 'string' ? model : model?.model_name,
  );

export const getMarketplaceDisplayGroup = (group, modelName) => {
  const groupName = normalizeMarketplaceGroup(group);
  if (
    groupName === MARKETPLACE_NON_WEB_GROUP &&
    isMarketplaceOnlineModelName(modelName)
  ) {
    return MARKETPLACE_WEB_GROUP;
  }
  return groupName;
};

export const getMarketplaceGroupLabel = (group, t, modelName) => {
  const displayGroup = getMarketplaceDisplayGroup(group, modelName);
  const label = MARKETPLACE_GROUP_LABELS[displayGroup];
  if (!label) return group;
  return t ? t(label) : label;
};

export const getMarketplaceDisplayGroups = (model) => {
  const groups = Array.isArray(model?.enable_groups) ? model.enable_groups : [];
  const displayGroups = groups
    .filter((group) => group && group !== 'auto')
    .map((group) => getMarketplaceDisplayGroup(group, model?.model_name));

  if (
    isMarketplaceOnlineModel(model) &&
    !displayGroups.includes(MARKETPLACE_WEB_GROUP)
  ) {
    displayGroups.unshift(MARKETPLACE_WEB_GROUP);
  }

  return Array.from(new Set(displayGroups));
};

export const marketplaceGroupMatchesModel = (model, group) => {
  const normalizedGroup = normalizeMarketplaceGroup(group);
  if (!normalizedGroup || normalizedGroup === 'all') return true;
  return getMarketplaceDisplayGroups(model).includes(normalizedGroup);
};

export const getMarketplacePricingGroup = (model, selectedGroup) => {
  if (!selectedGroup || selectedGroup === 'all') return selectedGroup;
  return normalizeMarketplaceGroup(selectedGroup);
};

export const getMarketplaceFilterGroups = (usableGroup = {}, models = []) => {
  const groups = Object.keys(usableGroup || {}).filter(
    (group) => group && group !== 'auto',
  );
  const normalizedGroups = Array.from(
    new Set(groups.map((group) => normalizeMarketplaceGroup(group))),
  );

  if (
    models.some((model) => isMarketplaceOnlineModel(model)) &&
    !normalizedGroups.includes(MARKETPLACE_WEB_GROUP)
  ) {
    normalizedGroups.push(MARKETPLACE_WEB_GROUP);
  }

  return normalizedGroups;
};

export const getMarketplaceBillingLabel = (t) =>
  t ? t('按量计费') : '按量计费';
