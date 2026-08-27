import { useCallback, useEffect, useState } from 'react'
import i18next from 'i18next'
import { toast } from 'sonner'
import { useCountdown } from '@/hooks/use-countdown'
import { getEmailVerificationCaptcha, sendEmailVerification } from '../api'
import { EMAIL_VERIFICATION_COUNTDOWN } from '../constants'
import type { EmailVerificationCaptcha } from '../types'

interface UseEmailVerificationOptions {
  enabled?: boolean
  turnstileToken?: string
  validateTurnstile?: () => boolean
}

/**
 * Hook for managing email verification code sending
 */
export function useEmailVerification(options?: UseEmailVerificationOptions) {
  const [isSending, setIsSending] = useState(false)
  const [isCaptchaLoading, setIsCaptchaLoading] = useState(false)
  const [captchaChallenge, setCaptchaChallenge] =
    useState<EmailVerificationCaptcha | null>(null)
  const [captchaCode, setCaptchaCode] = useState('')
  const enabled = options?.enabled ?? true
  const {
    secondsLeft,
    isActive,
    start: startCountdown,
    reset: resetCountdown,
  } = useCountdown({ initialSeconds: EMAIL_VERIFICATION_COUNTDOWN })

  const refreshCaptcha = useCallback(async () => {
    if (!enabled) return

    setIsCaptchaLoading(true)
    try {
      const res = await getEmailVerificationCaptcha()
      if (res?.success && res.data?.captcha_id && res.data?.image) {
        setCaptchaChallenge(res.data)
        setCaptchaCode('')
        return
      }
      toast.error(i18next.t('Failed to load captcha'))
      setCaptchaChallenge(null)
    } catch (_error) {
      setCaptchaChallenge(null)
    } finally {
      setIsCaptchaLoading(false)
    }
  }, [enabled])

  useEffect(() => {
    if (!enabled) {
      setCaptchaChallenge(null)
      setCaptchaCode('')
      return
    }

    void refreshCaptcha()
  }, [enabled, refreshCaptcha])

  /**
   * Send verification code to email
   */
  const sendCode = async (email: string) => {
    if (!email) {
      toast.error(i18next.t('Please enter your email first'))
      return false
    }

    // Validate turnstile if validation function is provided
    if (options?.validateTurnstile && !options.validateTurnstile()) {
      return false
    }

    if (!captchaChallenge?.captcha_id) {
      toast.error(i18next.t('Failed to load captcha'))
      await refreshCaptcha()
      return false
    }

    if (!captchaCode.trim()) {
      toast.error(i18next.t('Please enter the captcha code'))
      return false
    }

    setIsSending(true)
    try {
      const res = await sendEmailVerification(email, options?.turnstileToken, {
        captchaId: captchaChallenge.captcha_id,
        captchaCode,
      })
      if (res?.success) {
        startCountdown()
        toast.success(i18next.t('Verification email sent'))
        await refreshCaptcha()
        return true
      }
      await refreshCaptcha()
      return false
    } catch (_error) {
      // Errors are handled by global interceptor
      await refreshCaptcha()
      return false
    } finally {
      setIsSending(false)
    }
  }

  return {
    isSending,
    secondsLeft,
    isActive,
    captchaChallenge,
    captchaCode,
    isCaptchaLoading,
    setCaptchaCode,
    refreshCaptcha,
    reset: resetCountdown,
    sendCode,
  }
}
