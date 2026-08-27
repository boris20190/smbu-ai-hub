import { KeyRound } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { ApiKeysDialogs } from '@/features/keys/components/api-keys-dialogs'
import { ApiKeysPrimaryButtons } from '@/features/keys/components/api-keys-primary-buttons'
import { ApiKeysProvider } from '@/features/keys/components/api-keys-provider'
import { ApiKeysTable } from '@/features/keys/components/api-keys-table'
import { PanelWrapper } from '../ui/panel-wrapper'

export function ApiKeysPanel() {
  const { t } = useTranslation()

  return (
    <ApiKeysProvider>
      <PanelWrapper
        className='flex h-full flex-col'
        contentClassName='flex min-h-0 flex-1 flex-col'
        title={
          <span className='flex items-center gap-2'>
            <KeyRound className='text-muted-foreground/60 size-4' />
            {t('API Keys')}
          </span>
        }
        headerActions={<ApiKeysPrimaryButtons />}
      >
        <ApiKeysTable variant='overview' />
      </PanelWrapper>
      <ApiKeysDialogs />
    </ApiKeysProvider>
  )
}
