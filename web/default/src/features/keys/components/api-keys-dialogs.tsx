import { useState } from 'react'
import { ApiKeysDeleteDialog } from './api-keys-delete-dialog'
import { ApiKeysMutateDrawer } from './api-keys-mutate-drawer'
import { useApiKeys } from './api-keys-provider'
import { CCSwitchDialog } from './dialogs/cc-switch-dialog'

export function ApiKeysDialogs() {
  const { open, setOpen, currentRow, resolvedKey } = useApiKeys()
  const [lastOpen, setLastOpen] = useState(open)
  const [lastMutateSide, setLastMutateSide] = useState<'left' | 'right'>(
    'right'
  )

  if (open !== lastOpen) {
    setLastOpen(open)
    if (open === 'create') {
      setLastMutateSide('left')
    } else if (open === 'update') {
      setLastMutateSide('right')
    }
  }

  const mutateSide =
    open === 'create' ? 'left' : open === 'update' ? 'right' : lastMutateSide

  return (
    <>
      <ApiKeysMutateDrawer
        open={open === 'create' || open === 'update'}
        onOpenChange={(isOpen) => !isOpen && setOpen(null)}
        currentRow={open === 'update' ? currentRow || undefined : undefined}
        side={mutateSide}
      />
      <ApiKeysDeleteDialog />
      <CCSwitchDialog
        open={open === 'cc-switch'}
        onOpenChange={(isOpen) => !isOpen && setOpen(null)}
        tokenKey={resolvedKey}
      />
    </>
  )
}
