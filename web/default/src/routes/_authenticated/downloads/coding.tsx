import { createFileRoute } from '@tanstack/react-router'
import { CodingDownloads } from '@/features/downloads/coding'

export const Route = createFileRoute('/_authenticated/downloads/coding')({
  component: CodingDownloads,
})
