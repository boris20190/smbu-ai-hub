import { useCallback, useEffect, useState } from 'react'
import { Loader2, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useCountdown } from '@/hooks/use-countdown'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { EmailVerificationCaptcha } from '../../types'
import {
  sendEmailVerification,
  bindEmail,
  getEmailVerificationCaptcha,
} from '../../api'

// ============================================================================
// Email Bind Dialog Component
// ============================================================================

interface EmailBindDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentEmail?: string
  onSuccess: () => void
}

export function EmailBindDialog({
  open,
  onOpenChange,
  currentEmail,
  onSuccess,
}: EmailBindDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [captchaLoading, setCaptchaLoading] = useState(false)
  const [captchaChallenge, setCaptchaChallenge] =
    useState<EmailVerificationCaptcha | null>(null)
  const [captchaCode, setCaptchaCode] = useState('')
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const {
    secondsLeft,
    isActive,
    start: startCountdown,
    reset: resetCountdown,
  } = useCountdown({
    initialSeconds: 60,
  })

  const refreshCaptcha = useCallback(async () => {
    if (!open) return

    try {
      setCaptchaLoading(true)
      const response = await getEmailVerificationCaptcha()
      if (response.success && response.data?.captcha_id) {
        setCaptchaChallenge(response.data)
        setCaptchaCode('')
      } else {
        setCaptchaChallenge(null)
        toast.error(response.message || t('Failed to load captcha'))
      }
    } catch (_error) {
      setCaptchaChallenge(null)
      toast.error(t('Failed to load captcha'))
    } finally {
      setCaptchaLoading(false)
    }
  }, [open, t])

  useEffect(() => {
    if (open) {
      void refreshCaptcha()
    }
  }, [open, refreshCaptcha])

  const handleSendCode = async () => {
    if (!email || !email.includes('@')) {
      toast.error(t('Please enter a valid email address'))
      return
    }
    if (!captchaChallenge?.captcha_id) {
      toast.error(t('Failed to load captcha'))
      await refreshCaptcha()
      return
    }
    if (!captchaCode.trim()) {
      toast.error(t('Please enter the captcha code'))
      return
    }

    try {
      setSendingCode(true)
      const response = await sendEmailVerification(email, undefined, {
        captchaId: captchaChallenge.captcha_id,
        captchaCode,
      })

      if (response.success) {
        toast.success(t('Verification code sent! Please check your email.'))
        startCountdown()
        await refreshCaptcha()
      } else {
        toast.error(response.message || t('Failed to send verification code'))
        await refreshCaptcha()
      }
    } catch (_error) {
      toast.error(t('Failed to send verification code'))
      await refreshCaptcha()
    } finally {
      setSendingCode(false)
    }
  }

  const handleBind = async () => {
    if (!email || !code) {
      toast.error(t('Please enter email and verification code'))
      return
    }

    try {
      setLoading(true)
      const response = await bindEmail(email, code)

      if (response.success) {
        toast.success(t('Email bound successfully!'))
        onOpenChange(false)
        onSuccess()
        // Reset form
        setEmail('')
        setCode('')
        setCaptchaCode('')
        setCaptchaChallenge(null)
        resetCountdown()
      } else {
        toast.error(response.message || t('Failed to bind email'))
      }
    } catch (_error) {
      toast.error(t('Failed to bind email'))
    } finally {
      setLoading(false)
    }
  }

  const handleOpenChange = (open: boolean) => {
    if (!loading) {
      onOpenChange(open)
      if (!open) {
        // Reset form when closing
        setEmail('')
        setCode('')
        setCaptchaCode('')
        setCaptchaChallenge(null)
        resetCountdown()
      }
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{t('Bind Email')}</DialogTitle>
          <DialogDescription>
            {currentEmail
              ? t('Current email: {{email}}. Enter a new email to change.', {
                  email: currentEmail,
                })
              : t('Bind an email address to your account.')}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-4'>
          <div className='space-y-2'>
            <Label htmlFor='email'>{t('Email Address')}</Label>
            <Input
              id='email'
              type='email'
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t('Enter your email')}
              disabled={loading}
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='email-bind-captcha'>{t('Manual captcha')}</Label>
            <div className='flex gap-2'>
              <div className='flex min-h-10 flex-1 items-center overflow-hidden rounded-md border bg-muted/30'>
                {captchaChallenge?.image ? (
                  <img
                    src={captchaChallenge.image}
                    alt={t('Manual captcha')}
                    className='h-10 w-full object-contain'
                  />
                ) : (
                  <div className='text-muted-foreground flex w-full items-center justify-center text-sm'>
                    {captchaLoading ? (
                      <Loader2 className='h-4 w-4 animate-spin' />
                    ) : (
                      t('Captcha unavailable')
                    )}
                  </div>
                )}
              </div>
              <Button
                type='button'
                variant='outline'
                size='icon'
                aria-label={t('Refresh captcha')}
                disabled={captchaLoading || loading}
                onClick={() => {
                  void refreshCaptcha()
                }}
              >
                {captchaLoading ? (
                  <Loader2 className='h-4 w-4 animate-spin' />
                ) : (
                  <RefreshCw className='h-4 w-4' />
                )}
              </Button>
            </div>
            <Input
              id='email-bind-captcha'
              value={captchaCode}
              onChange={(e) => setCaptchaCode(e.target.value.toUpperCase())}
              placeholder={t('Enter captcha code')}
              disabled={loading}
              autoComplete='off'
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='code'>{t('Verification Code')}</Label>
            <div className='flex gap-2'>
              <Input
                id='code'
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={t('Enter code')}
                disabled={loading}
                maxLength={6}
              />
              <Button
                type='button'
                variant='outline'
                onClick={handleSendCode}
                disabled={
                  sendingCode ||
                  isActive ||
                  captchaLoading ||
                  !email ||
                  !captchaChallenge?.captcha_id ||
                  !captchaCode.trim()
                }
              >
                {isActive
                  ? `${secondsLeft}s`
                  : sendingCode
                    ? t('Sending...')
                    : t('Send')}
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => handleOpenChange(false)}
            disabled={loading}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='button'
            onClick={handleBind}
            disabled={loading || !email || !code}
          >
            {loading && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {loading ? t('Binding...') : t('Bind Email')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
