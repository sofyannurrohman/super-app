import apiClient from './axios'

export interface AIWidget {
  type: 'chart' | 'approval_list' | 'attendance_stats' | string
  title?: string
  subtitle?: string
  chart_type?: 'bar' | 'pie' | 'line' | string
  data?: Record<string, any>
  summary?: Record<string, any>
  items?: any[]
}

export interface AIResponse {
  answer: string
  provider: string
}

export async function askAI(
  message: string,
  signal?: AbortSignal
): Promise<AIResponse> {
  const { data } = await apiClient.post<AIResponse>(
    '/ai/chat',
    { message },
    {
      skipLoadingBar: true,
      headers: { 'X-Skip-Loading': 'true' },
      timeout: 60000,
      signal
    }
  )
  return data
}
