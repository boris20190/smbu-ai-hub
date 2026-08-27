import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Code2,
  Download,
  ExternalLink,
  Laptop,
  Monitor,
  Package,
  Puzzle,
  Terminal,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Main } from '@/components/layout'
import {
  downloadCCSwitchMirrorAsset,
  getCCSwitchMirrorManifest,
} from './api'
import type { CCSwitchMirrorAsset, CCSwitchMirrorPlatform } from './types'

type ExtensionLink = {
  name: string
  publisher: string
  marketplaceId: string
  href: string
}

const vscodeExtensions: ExtensionLink[] = [
  {
    name: 'OpenCode',
    publisher: 'sst-dev',
    marketplaceId: 'sst-dev.opencode',
    href: 'https://marketplace.visualstudio.com/items?itemName=sst-dev.opencode',
  },
  {
    name: 'Codex',
    publisher: 'OpenAI',
    marketplaceId: 'openai.chatgpt',
    href: 'https://marketplace.visualstudio.com/items?itemName=openai.chatgpt',
  },
  {
    name: 'Claude Code',
    publisher: 'Anthropic',
    marketplaceId: 'anthropic.claude-code',
    href: 'https://marketplace.visualstudio.com/items?itemName=anthropic.claude-code',
  },
]

const platformIcons: Record<string, LucideIcon> = {
  windows: Monitor,
  linux: Terminal,
  mac: Laptop,
}

const platformLabels: Record<string, { label: string; packageType: string }> = {
  windows: { label: 'Windows', packageType: '.msi' },
  linux: { label: 'Linux', packageType: '.AppImage' },
  mac: { label: 'Mac', packageType: '.dmg' },
}

const preferredPlatformOrder = ['windows', 'linux', 'mac']

function formatBytes(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) return ''
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function formatMirroredAt(value: number) {
  if (!value) return ''
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value * 1000))
}

function sortPlatforms(platforms: CCSwitchMirrorPlatform[]) {
  return [...platforms].sort((a, b) => {
    const aIndex = preferredPlatformOrder.indexOf(a.key)
    const bIndex = preferredPlatformOrder.indexOf(b.key)
    return (aIndex === -1 ? 99 : aIndex) - (bIndex === -1 ? 99 : bIndex)
  })
}

function startBrowserDownload(url: string, fileName: string) {
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
}

export function CodingDownloads() {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: ['cc-switch-mirror-manifest'],
    queryFn: getCCSwitchMirrorManifest,
  })
  const platforms = sortPlatforms(data?.platforms ?? [])

  return (
    <Main>
      <div className='min-h-0 flex-1 overflow-auto px-3 py-3 sm:px-4 sm:py-6'>
        <article className='mx-auto flex w-full max-w-6xl flex-col gap-5 sm:gap-6'>
          <header className='border-border/70 bg-card rounded-lg border px-5 py-5 sm:px-6'>
            <div className='flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between'>
              <div className='space-y-2'>
                <div className='text-muted-foreground flex items-center gap-2 text-sm font-medium'>
                  <Code2 className='size-4' />
                  <span>{t('Coding')}</span>
                </div>
                <h1 className='text-2xl font-semibold tracking-tight sm:text-3xl'>
                  {t('Coding Downloads')}
                </h1>
                <p className='text-muted-foreground max-w-3xl text-sm leading-6 sm:text-base'>
                  {t(
                    'Download coding clients and open editor extensions from one place.'
                  )}
                </p>
              </div>
              {data?.repository_url ? (
                <Button
                  variant='outline'
                  render={
                    <a
                      href={data.repository_url}
                      target='_blank'
                      rel='noopener noreferrer'
                    />
                  }
                >
                  <ExternalLink className='size-4' />
                  GitHub
                </Button>
              ) : null}
            </div>
          </header>

          <section className='space-y-4'>
            <div className='flex items-center gap-2'>
              <Package className='text-primary size-5' />
              <div>
                <h2 className='text-lg font-semibold'>CC Switch</h2>
                <p className='text-muted-foreground text-sm'>
                  {t('Mirrored installers from this server.')}
                </p>
              </div>
            </div>

            <div className='grid gap-4 md:grid-cols-3'>
              {isLoading
                ? preferredPlatformOrder.map((key) => (
                    <PlatformSkeleton key={key} title={t('Loading')} />
                  ))
                : platforms.length > 0
                  ? platforms.map((platform) => (
                      <PlatformCard key={platform.key} platform={platform} />
                    ))
                  : preferredPlatformOrder.map((key) => (
                      <UnavailablePlatform key={key} platformKey={key} />
                    ))}
            </div>
          </section>

          <section className='space-y-4'>
            <div className='flex items-center gap-2'>
              <Puzzle className='text-primary size-5' />
              <div>
                <h2 className='text-lg font-semibold'>
                  {t('VS Code extensions')}
                </h2>
                <p className='text-muted-foreground text-sm'>
                  {t('Open Marketplace pages and install inside VS Code.')}
                </p>
              </div>
            </div>

            <div className='grid gap-3 md:grid-cols-3'>
              {vscodeExtensions.map((extension) => (
                <div
                  key={extension.marketplaceId}
                  className='bg-card rounded-lg border p-4'
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div className='min-w-0'>
                      <h3 className='truncate font-semibold'>
                        {extension.name}
                      </h3>
                      <p className='text-muted-foreground mt-1 text-sm'>
                        {extension.publisher}
                      </p>
                    </div>
                    <Puzzle className='text-muted-foreground size-5 shrink-0' />
                  </div>
                  <p className='text-muted-foreground mt-3 truncate font-mono text-xs'>
                    {extension.marketplaceId}
                  </p>
                  <Button
                    className='mt-4 w-full'
                    variant='outline'
                    render={
                      <a
                        href={extension.href}
                        target='_blank'
                        rel='noopener noreferrer'
                      />
                    }
                  >
                    <ExternalLink className='size-4' />
                    Marketplace
                  </Button>
                </div>
              ))}
            </div>
          </section>
        </article>
      </div>
    </Main>
  )
}

function PlatformCard({ platform }: { platform: CCSwitchMirrorPlatform }) {
  const { t } = useTranslation()
  const Icon = platformIcons[platform.key] ?? Package
  const assets = platform.assets ?? []

  return (
    <div className='bg-card flex min-h-72 flex-col rounded-lg border p-5'>
      <div className='flex flex-col items-center text-center'>
        <div className='bg-muted flex size-20 items-center justify-center rounded-lg'>
          <Icon className='size-10' />
        </div>
        <h3 className='mt-4 text-xl font-semibold'>{platform.label}</h3>
        <p className='text-muted-foreground mt-1 text-sm'>
          {platform.package_type}
        </p>
      </div>

      <div className='mt-5 flex flex-1 flex-col gap-3'>
        {assets.length > 0 ? (
          assets.map((asset) => (
            <AssetDownloadLink key={asset.id} asset={asset} />
          ))
        ) : (
          <button
            type='button'
            disabled
            className='bg-muted text-muted-foreground flex h-10 w-full items-center justify-center gap-2 rounded-lg px-3 text-sm font-medium'
          >
            <Download className='size-4' />
            {t('Not mirrored yet')}
          </button>
        )}
      </div>
    </div>
  )
}

function AssetDownloadLink({ asset }: { asset: CCSwitchMirrorAsset }) {
  const { t } = useTranslation()
  const [isDownloading, setIsDownloading] = useState(false)
  const size = formatBytes(asset.size)
  const mirroredAt = formatMirroredAt(asset.mirrored_at)

  async function handleDownload() {
    setIsDownloading(true)
    try {
      const downloadUrl = await downloadCCSwitchMirrorAsset(asset)
      startBrowserDownload(downloadUrl, asset.file_name)
    } catch {
      toast.error(t('Download failed'))
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <button
      type='button'
      onClick={handleDownload}
      disabled={isDownloading}
      className='border-border hover:border-primary/60 hover:bg-muted/50 flex min-h-16 w-full items-center gap-3 rounded-lg border px-3 py-2 transition-colors'
    >
      <span className='bg-primary text-primary-foreground flex size-8 shrink-0 items-center justify-center rounded-md'>
        <Download className='size-4' />
      </span>
      <span className='min-w-0 flex-1'>
        <span className='block truncate text-sm font-medium'>
          {isDownloading ? t('Downloading...') : asset.file_name}
        </span>
        <span className='text-muted-foreground mt-1 flex flex-wrap gap-x-2 gap-y-1 text-xs'>
          {asset.version ? <span>{asset.version}</span> : null}
          {size ? <span>{size}</span> : null}
          {mirroredAt ? (
            <span>
              {t('Mirrored at')} {mirroredAt}
            </span>
          ) : null}
        </span>
      </span>
    </button>
  )
}

function PlatformSkeleton({ title }: { title: string }) {
  return (
    <div className='bg-card flex min-h-72 flex-col rounded-lg border p-5'>
      <div className='bg-muted mx-auto size-20 animate-pulse rounded-lg' />
      <div className='bg-muted mx-auto mt-4 h-6 w-28 animate-pulse rounded' />
      <div className='bg-muted mx-auto mt-2 h-4 w-16 animate-pulse rounded' />
      <div className='bg-muted mt-5 h-10 animate-pulse rounded-lg' />
      <span className='sr-only'>{title}</span>
    </div>
  )
}

function UnavailablePlatform({ platformKey }: { platformKey: string }) {
  const meta = platformLabels[platformKey] ?? {
    label: platformKey,
    packageType: '',
  }
  const platform: CCSwitchMirrorPlatform = {
    key: platformKey,
    label: meta.label,
    package_type: meta.packageType,
    available: false,
    assets: [],
  }
  return <PlatformCard platform={platform} />
}
