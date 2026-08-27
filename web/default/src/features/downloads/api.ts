import { api } from '@/lib/api'
import type { CCSwitchMirrorAsset, CCSwitchMirrorManifest } from './types'

type ManifestResponse = {
  success: boolean
  message?: string
  data?: CCSwitchMirrorManifest
}

type DownloadTicketResponse = {
  success: boolean
  message?: string
  data?: {
    download_url: string
    expires_at: number
  }
}

export async function getCCSwitchMirrorManifest(): Promise<CCSwitchMirrorManifest> {
  const res = await api.get<ManifestResponse>(
    '/api/downloads/cc-switch/manifest'
  )
  if (!res.data.success || !res.data.data) {
    return {
      repository_url: 'https://github.com/farion1231/cc-switch',
      platforms: [],
    }
  }
  return res.data.data
}

export async function downloadCCSwitchMirrorAsset(
  asset: CCSwitchMirrorAsset
): Promise<string> {
  const res = await api.post<DownloadTicketResponse>(
    `${asset.download_url}/ticket`,
    undefined,
    { skipBusinessError: true } as Record<string, unknown>
  )
  if (!res.data.success || !res.data.data?.download_url) {
    throw new Error(res.data.message || 'Download failed')
  }
  return res.data.data.download_url
}
