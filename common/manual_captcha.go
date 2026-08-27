package common

import (
	"bytes"
	"crypto/subtle"
	"encoding/base64"
	"image"
	"image/color"
	imagedraw "image/draw"
	"image/png"
	"math/big"
	"strings"
	"sync"
	"time"

	"crypto/rand"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	ManualCaptchaTTLSeconds = 180
	manualCaptchaLength     = 5
	manualCaptchaMaxSize    = 200
	manualCaptchaCharset    = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
)

type ManualCaptchaChallenge struct {
	CaptchaId string `json:"captcha_id"`
	Image     string `json:"image"`
	ExpiresIn int    `json:"expires_in"`
}

type manualCaptchaValue struct {
	answer    string
	expiresAt time.Time
}

var (
	manualCaptchaMutex sync.Mutex
	manualCaptchaMap   = make(map[string]manualCaptchaValue)
)

func GenerateManualCaptchaChallenge() (ManualCaptchaChallenge, error) {
	answer, err := generateManualCaptchaAnswer()
	if err != nil {
		return ManualCaptchaChallenge{}, err
	}
	imageDataURL, err := renderManualCaptchaImage(answer)
	if err != nil {
		return ManualCaptchaChallenge{}, err
	}

	captchaId := GetUUID()
	now := time.Now()
	expiresAt := now.Add(time.Duration(ManualCaptchaTTLSeconds) * time.Second)
	manualCaptchaMutex.Lock()
	trimManualCaptchasLocked(now)
	manualCaptchaMap[captchaId] = manualCaptchaValue{
		answer:    answer,
		expiresAt: expiresAt,
	}
	if len(manualCaptchaMap) > manualCaptchaMaxSize {
		trimManualCaptchasLocked(now)
	}
	manualCaptchaMutex.Unlock()

	return ManualCaptchaChallenge{
		CaptchaId: captchaId,
		Image:     imageDataURL,
		ExpiresIn: ManualCaptchaTTLSeconds,
	}, nil
}

func VerifyManualCaptcha(captchaId string, answer string) bool {
	captchaId = strings.TrimSpace(captchaId)
	normalizedAnswer := strings.ToUpper(strings.TrimSpace(answer))
	if captchaId == "" || normalizedAnswer == "" {
		return false
	}

	manualCaptchaMutex.Lock()
	defer manualCaptchaMutex.Unlock()

	value, ok := manualCaptchaMap[captchaId]
	if !ok {
		return false
	}
	delete(manualCaptchaMap, captchaId)
	if time.Now().After(value.expiresAt) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(normalizedAnswer), []byte(value.answer)) == 1
}

func generateManualCaptchaAnswer() (string, error) {
	answer := make([]byte, manualCaptchaLength)
	max := big.NewInt(int64(len(manualCaptchaCharset)))
	for i := range answer {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		answer[i] = manualCaptchaCharset[n.Int64()]
	}
	return string(answer), nil
}

func renderManualCaptchaImage(answer string) (string, error) {
	smallBounds := image.Rect(0, 0, 12+len(answer)*8, 24)
	small := image.NewRGBA(smallBounds)
	imagedraw.Draw(small, small.Bounds(), image.NewUniform(color.RGBA{R: 246, G: 248, B: 250, A: 255}), image.Point{}, imagedraw.Src)

	d := &font.Drawer{
		Dst:  small,
		Src:  image.NewUniform(color.RGBA{R: 32, G: 41, B: 58, A: 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(6, 17),
	}
	d.DrawString(answer)

	scale := 3
	largeBounds := image.Rect(0, 0, smallBounds.Dx()*scale, smallBounds.Dy()*scale)
	large := image.NewRGBA(largeBounds)
	xdraw.NearestNeighbor.Scale(large, large.Bounds(), small, small.Bounds(), xdraw.Over, nil)
	addManualCaptchaNoise(large)

	var buf bytes.Buffer
	if err := png.Encode(&buf, large); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func addManualCaptchaNoise(img *image.RGBA) {
	bounds := img.Bounds()
	for i := 0; i < 80; i++ {
		x, err := randomManualCaptchaInt(bounds.Dx())
		if err != nil {
			return
		}
		y, err := randomManualCaptchaInt(bounds.Dy())
		if err != nil {
			return
		}
		img.Set(bounds.Min.X+x, bounds.Min.Y+y, color.RGBA{R: 120, G: 135, B: 160, A: 120})
	}
	for i := 0; i < 3; i++ {
		x1, err := randomManualCaptchaInt(bounds.Dx())
		if err != nil {
			return
		}
		y1, err := randomManualCaptchaInt(bounds.Dy())
		if err != nil {
			return
		}
		x2, err := randomManualCaptchaInt(bounds.Dx())
		if err != nil {
			return
		}
		y2, err := randomManualCaptchaInt(bounds.Dy())
		if err != nil {
			return
		}
		drawManualCaptchaLine(img, x1, y1, x2, y2)
	}
}

func randomManualCaptchaInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func drawManualCaptchaLine(img *image.RGBA, x1 int, y1 int, x2 int, y2 int) {
	dx := absManualCaptchaInt(x2 - x1)
	dy := -absManualCaptchaInt(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx + dy
	lineColor := color.RGBA{R: 88, G: 106, B: 135, A: 100}
	for {
		img.Set(x1, y1, lineColor)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x1 += sx
		}
		if e2 <= dx {
			err += dx
			y1 += sy
		}
	}
}

func absManualCaptchaInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func trimManualCaptchasLocked(now time.Time) {
	for key, value := range manualCaptchaMap {
		if now.After(value.expiresAt) {
			delete(manualCaptchaMap, key)
		}
	}
	for len(manualCaptchaMap) > manualCaptchaMaxSize {
		var oldestKey string
		var oldestExpiry time.Time
		for key, value := range manualCaptchaMap {
			if oldestKey == "" || value.expiresAt.Before(oldestExpiry) {
				oldestKey = key
				oldestExpiry = value.expiresAt
			}
		}
		if oldestKey == "" {
			return
		}
		delete(manualCaptchaMap, oldestKey)
	}
}
