import { useMemo } from 'react'
import { useAuthStore } from '@/stores/auth-store'
import { getCurrencyLabel, isCurrencyDisplayEnabled } from '@/lib/currency'
import { formatNumber, formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { useStatus } from '@/hooks/use-status'
import { StaggerContainer, StaggerItem } from '@/components/page-transition'
import { useSummaryCardsConfig } from '@/features/dashboard/hooks/use-dashboard-config'
import { StatCard } from '../ui/stat-card'

interface SummaryCardsProps {
  embedded?: boolean
  compact?: boolean
}

export function SummaryCards(props: SummaryCardsProps = {}) {
  const user = useAuthStore((state) => state.auth.user)
  const { status, loading } = useStatus()

  const summaryValues = useMemo(() => {
    const remainQuota = Number(user?.quota ?? 0)
    const usedQuota = Number(user?.used_quota ?? 0)
    const requestCount = Number(user?.request_count ?? 0)

    return {
      remainDisplay: formatQuota(remainQuota),
      usedDisplay: formatQuota(usedQuota),
      requestCountDisplay: formatNumber(requestCount),
    }
  }, [user])

  const currencyEnabledFromStore = isCurrencyDisplayEnabled()
  const statusCurrencyFlag =
    typeof status?.display_in_currency === 'boolean'
      ? Boolean(status.display_in_currency)
      : undefined
  const currencyEnabled =
    statusCurrencyFlag !== undefined
      ? statusCurrencyFlag
      : currencyEnabledFromStore
  const currencyLabel = currencyEnabled ? getCurrencyLabel() : 'Tokens'

  const items = useSummaryCardsConfig({
    ...summaryValues,
    currencyEnabled,
    currencyLabel,
  }).map((config, index) => ({
    title: config.title,
    value: config.value,
    desc: config.description,
    icon: config.icon,
    isBalance: index === 0,
  }))

  const content = (
    <StaggerContainer className='divide-border/60 grid h-full grid-cols-3 divide-x'>
      {items.map((it) => (
        <StaggerItem
          key={it.title}
          className={cn(
            'px-3 py-3 sm:px-5 sm:py-4',
            props.compact && 'sm:px-4 sm:py-3'
          )}
        >
          <StatCard
            title={it.title}
            value={it.value}
            description={it.desc}
            icon={it.icon}
            loading={loading}
          />
        </StaggerItem>
      ))}
    </StaggerContainer>
  )

  if (props.embedded) return content

  return <div className='overflow-hidden rounded-lg border'>{content}</div>
}
