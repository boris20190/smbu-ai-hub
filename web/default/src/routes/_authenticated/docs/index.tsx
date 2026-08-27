import { createFileRoute } from '@tanstack/react-router'
import { Docs } from '@/features/docs'

export const Route = createFileRoute('/_authenticated/docs/')({
  component: Docs,
})
