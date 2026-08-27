import { api } from '@/lib/api'
import type {
  ApiResponse,
  Feedback,
  FeedbackCreatePayload,
  FeedbackListResponse,
} from './types'

export async function submitFeedback(
  payload: FeedbackCreatePayload
): Promise<ApiResponse<Feedback>> {
  const res = await api.post('/api/feedback', payload)
  return res.data
}

export async function getFeedbacks(params: {
  p?: number
  page_size?: number
}): Promise<ApiResponse<FeedbackListResponse>> {
  const { p = 1, page_size = 10 } = params
  const res = await api.get('/api/feedback', {
    params: { p, page_size },
  })
  return res.data
}
