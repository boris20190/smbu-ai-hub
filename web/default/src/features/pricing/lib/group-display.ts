import type { TFunction } from 'i18next'
import type { PricingModel } from '../types'

export const MARKETPLACE_GROUP_LABEL_KEYS = {
  default: 'Default',
} as const

export function normalizeMarketplaceGroup(group?: string | null): string {
  return group?.trim() ?? ''
}

export function getMarketplaceDisplayGroup(
  group?: string | null,
  _modelName?: string | null
): string {
  return normalizeMarketplaceGroup(group)
}

export function getMarketplaceGroupLabel(
  group?: string | null,
  t?: TFunction,
  modelName?: string | null
): string {
  const groupName = getMarketplaceDisplayGroup(group?.trim(), modelName)
  if (!groupName) return ''
  const labelKey =
    MARKETPLACE_GROUP_LABEL_KEYS[
      groupName as keyof typeof MARKETPLACE_GROUP_LABEL_KEYS
    ]

  if (!labelKey) return groupName
  return t ? t(labelKey) : labelKey
}

export function getMarketplaceDisplayGroups(
  model?: Pick<PricingModel, 'enable_groups' | 'model_name'> | null
): string[] {
  const groups = Array.isArray(model?.enable_groups) ? model.enable_groups : []
  const displayGroups = groups
    .filter((group) => group && group !== 'auto')
    .map((group) => getMarketplaceDisplayGroup(group, model?.model_name))

  return Array.from(new Set(displayGroups))
}

export function marketplaceGroupMatchesModel(
  model: Pick<PricingModel, 'enable_groups' | 'model_name'>,
  group: string
): boolean {
  const normalizedGroup = normalizeMarketplaceGroup(group)
  if (!normalizedGroup || normalizedGroup === 'all') return true
  return getMarketplaceDisplayGroups(model).includes(normalizedGroup)
}

export function getMarketplaceFilterGroups(
  usableGroup: Record<string, unknown> = {},
  _models: Array<Pick<PricingModel, 'model_name'>> = []
): string[] {
  const groups = Object.keys(usableGroup || {}).filter(
    (group) => group && group !== 'auto'
  )
  const normalizedGroups = Array.from(
    new Set(groups.map(normalizeMarketplaceGroup))
  )

  return normalizedGroups
}
