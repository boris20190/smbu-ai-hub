import { type FormEvent, useState } from 'react'
import { ExternalLink, Gift, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { getSelf } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useRedemption } from '@/features/wallet/hooks'
import { PanelWrapper } from '../ui/panel-wrapper'

const REDEMPTION_CODE_PURCHASE_URL = 'https://m.tb.cn/h.izZnrdB?tk=APxm5otDawQ'

export function QuotaRedemptionPanel() {
  const { t } = useTranslation()
  const [code, setCode] = useState('')
  const { redeeming, redeemCode } = useRedemption()
  const setUser = useAuthStore((state) => state.auth.setUser)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const success = await redeemCode(code)
    if (!success) return

    setCode('')
    const response = await getSelf()
    if (response.success && response.data) {
      setUser(response.data)
    }
  }

  return (
    <PanelWrapper
      className='flex h-full flex-col'
      contentClassName='flex flex-1 flex-col'
      title={
        <span className='flex items-center gap-2'>
          <Gift className='text-muted-foreground/60 size-4' />
          {t('Quota Redemption')}
        </span>
      }
    >
      <form onSubmit={handleSubmit} className='flex flex-1 flex-col gap-3'>
        <div className='space-y-2'>
          <Label
            htmlFor='overview-redemption-code'
            className='text-muted-foreground text-xs font-medium tracking-wider uppercase'
          >
            {t('Redemption Code')}
          </Label>
          <div className='grid grid-cols-[minmax(0,1fr)_auto] gap-2'>
            <Input
              id='overview-redemption-code'
              value={code}
              onChange={(event) => setCode(event.target.value)}
              placeholder={t('Enter your redemption code')}
              className='h-9 min-w-0'
            />
            <Button type='submit' disabled={redeeming} className='h-9 px-4'>
              {redeeming && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
              {t('Redeem')}
            </Button>
          </div>
        </div>
        <div className='space-y-1'>
          <p className='text-muted-foreground text-xs'>
            {t('Redeem a code to add quota to your account balance.')}
          </p>
          <a
            href={REDEMPTION_CODE_PURCHASE_URL}
            target='_blank'
            rel='noreferrer'
            className='text-primary hover:text-primary/80 inline-flex w-fit items-center gap-1 text-xs font-medium transition-colors hover:underline'
          >
            {t('Buy Redemption Code')}
            <ExternalLink className='size-3' />
          </a>
        </div>
      </form>
    </PanelWrapper>
  )
}
