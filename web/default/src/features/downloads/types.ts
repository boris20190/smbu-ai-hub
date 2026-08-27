export type CCSwitchMirrorAsset = {
  id: string
  name: string
  file_name: string
  size: number
  version: string
  mirrored_at: number
  download_url: string
}

export type CCSwitchMirrorPlatform = {
  key: string
  label: string
  package_type: string
  available: boolean
  assets: CCSwitchMirrorAsset[]
}

export type CCSwitchMirrorManifest = {
  repository_url: string
  platforms: CCSwitchMirrorPlatform[]
}
