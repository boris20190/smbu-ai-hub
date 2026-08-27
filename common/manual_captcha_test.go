package common

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGenerateManualCaptchaChallenge(t *testing.T) {
	challenge, err := GenerateManualCaptchaChallenge()
	if err != nil {
		t.Fatalf("GenerateManualCaptchaChallenge() error = %v", err)
	}
	if challenge.CaptchaId == "" {
		t.Fatal("CaptchaId is empty")
	}
	if !strings.HasPrefix(challenge.Image, "data:image/png;base64,") {
		t.Fatalf("Image prefix = %q", challenge.Image[:22])
	}
	if challenge.ExpiresIn != ManualCaptchaTTLSeconds {
		t.Fatalf("ExpiresIn = %d, want %d", challenge.ExpiresIn, ManualCaptchaTTLSeconds)
	}

	manualCaptchaMutex.Lock()
	delete(manualCaptchaMap, challenge.CaptchaId)
	manualCaptchaMutex.Unlock()
}

func TestVerifyManualCaptchaConsumesChallenge(t *testing.T) {
	const captchaId = "test-captcha"
	manualCaptchaMutex.Lock()
	manualCaptchaMap[captchaId] = manualCaptchaValue{
		answer:    "AB234",
		expiresAt: time.Now().Add(time.Minute),
	}
	manualCaptchaMutex.Unlock()

	if !VerifyManualCaptcha(" "+captchaId+" ", " ab234 ") {
		t.Fatal("VerifyManualCaptcha() = false, want true")
	}
	if VerifyManualCaptcha(captchaId, "AB234") {
		t.Fatal("VerifyManualCaptcha() reused challenge = true, want false")
	}
}

func TestVerifyManualCaptchaRejectsExpiredChallenge(t *testing.T) {
	const captchaId = "expired-captcha"
	manualCaptchaMutex.Lock()
	manualCaptchaMap[captchaId] = manualCaptchaValue{
		answer:    "AB234",
		expiresAt: time.Now().Add(-time.Second),
	}
	manualCaptchaMutex.Unlock()

	if VerifyManualCaptcha(captchaId, "AB234") {
		t.Fatal("VerifyManualCaptcha() expired challenge = true, want false")
	}
}

func TestGenerateManualCaptchaChallengeTrimsExpiredChallenges(t *testing.T) {
	clearManualCaptchasForTest()

	manualCaptchaMutex.Lock()
	manualCaptchaMap["expired-captcha"] = manualCaptchaValue{
		answer:    "AB234",
		expiresAt: time.Now().Add(-time.Second),
	}
	manualCaptchaMap["fresh-captcha"] = manualCaptchaValue{
		answer:    "CD567",
		expiresAt: time.Now().Add(time.Minute),
	}
	manualCaptchaMutex.Unlock()

	challenge, err := GenerateManualCaptchaChallenge()
	if err != nil {
		t.Fatalf("GenerateManualCaptchaChallenge() error = %v", err)
	}

	manualCaptchaMutex.Lock()
	_, expiredStillPresent := manualCaptchaMap["expired-captcha"]
	_, freshStillPresent := manualCaptchaMap["fresh-captcha"]
	delete(manualCaptchaMap, challenge.CaptchaId)
	delete(manualCaptchaMap, "fresh-captcha")
	manualCaptchaMutex.Unlock()

	if expiredStillPresent {
		t.Fatal("expired captcha was not trimmed")
	}
	if !freshStillPresent {
		t.Fatal("fresh captcha was trimmed unexpectedly")
	}
}

func TestGenerateManualCaptchaChallengeCapsStoredChallenges(t *testing.T) {
	clearManualCaptchasForTest()

	now := time.Now()
	manualCaptchaMutex.Lock()
	for i := 0; i < manualCaptchaMaxSize; i++ {
		manualCaptchaMap[fmt.Sprintf("captcha-%03d", i)] = manualCaptchaValue{
			answer:    "AB234",
			expiresAt: now.Add(time.Minute + time.Duration(i)*time.Millisecond),
		}
	}
	manualCaptchaMutex.Unlock()

	challenge, err := GenerateManualCaptchaChallenge()
	if err != nil {
		t.Fatalf("GenerateManualCaptchaChallenge() error = %v", err)
	}

	manualCaptchaMutex.Lock()
	size := len(manualCaptchaMap)
	_, oldestStillPresent := manualCaptchaMap["captcha-000"]
	delete(manualCaptchaMap, challenge.CaptchaId)
	for i := 0; i < manualCaptchaMaxSize; i++ {
		delete(manualCaptchaMap, fmt.Sprintf("captcha-%03d", i))
	}
	manualCaptchaMutex.Unlock()

	if size > manualCaptchaMaxSize {
		t.Fatalf("manualCaptchaMap size = %d, want <= %d", size, manualCaptchaMaxSize)
	}
	if oldestStillPresent {
		t.Fatal("oldest captcha was not trimmed when max size was exceeded")
	}
}

func clearManualCaptchasForTest() {
	manualCaptchaMutex.Lock()
	defer manualCaptchaMutex.Unlock()
	for key := range manualCaptchaMap {
		delete(manualCaptchaMap, key)
	}
}
