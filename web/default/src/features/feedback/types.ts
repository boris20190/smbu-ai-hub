export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface Feedback {
  id: number
  user_id: number
  username: string
  title: string
  content: string
  contact: string
  created_time: number
}

export interface FeedbackCreatePayload {
  title: string
  content: string
  contact?: string
}

export interface FeedbackListResponse {
  items: Feedback[]
  total: number
  page: number
  page_size: number
}
