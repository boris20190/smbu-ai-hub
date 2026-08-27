import { createFileRoute } from '@tanstack/react-router'
import { Feedback } from '@/features/feedback'

export const Route = createFileRoute('/_authenticated/feedback/')({
  component: Feedback,
})
