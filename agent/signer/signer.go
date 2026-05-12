package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// SignRequest computes HMAC-SHA256(method + path + timestamp + body, secret)
// and returns the X-Signature and X-Timestamp header values.
func SignRequest(method, path string, body []byte, secret string) (signature, timestamp string) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(method))
	mac.Write([]byte(path))
	mac.Write([]byte(ts))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil)), ts
}

// MustSignHeader is a convenience wrapper that returns header key-value pairs
// or panics on unrecoverable error (should never happen for valid inputs).
func SignHeaders(method, path string, body []byte, secret string) map[string]string {
	sig, ts := SignRequest(method, path, body, secret)
	return map[string]string{
		"X-Signature": sig,
		"X-Timestamp": ts,
	}
}

// VerifyRequest checks whether the provided signature matches the expected HMAC,
// and whether the timestamp is within the allowed skew window.
func VerifyRequest(method, path, sigHeader, tsHeader string, body []byte, secret []byte, maxSkewSeconds int64) error {
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	now := time.Now().Unix()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}
	if diff > maxSkewSeconds {
		return fmt.Errorf("timestamp skew too large: %d seconds", diff)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(method))
	mac.Write([]byte(path))
	mac.Write([]byte(tsHeader))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigHeader), []byte(expected)) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}
