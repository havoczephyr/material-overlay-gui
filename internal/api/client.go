package api

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/havoczephyr/material-overlay-gui/internal/ratelimit"
)

type Client struct {
	rateLimiter *ratelimit.Limiter
	baseURL     string
	userAgent   string
}

func NewClient(limiter *ratelimit.Limiter) *Client {
	return &Client{
		rateLimiter: limiter,
		baseURL:     "https://yugipedia.com/api.php",
		userAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

// Param is a key-value pair for building query strings.
type Param struct {
	Key, Value string
}

// doRequest executes a GET request via curl to avoid Cloudflare TLS fingerprint
// blocking of Go's net/http client.
func (c *Client) doRequest(params []Param) ([]byte, error) {
	c.rateLimiter.Wait()

	var parts []string
	for _, p := range params {
		key := strings.ReplaceAll(p.Key, " ", "+")
		val := strings.ReplaceAll(p.Value, " ", "+")
		parts = append(parts, key+"="+val)
	}
	rawURL := c.baseURL + "?" + strings.Join(parts, "&")

	cmd := exec.Command("curl",
		"-s",           // silent
		"--globoff",    // don't interpret brackets
		"-f",           // fail on HTTP errors
		"--compressed", // handle gzip/br automatically
		"-H", "User-Agent: "+c.userAgent,
		"-H", "Accept: application/json",
		rawURL,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 22 {
			return nil, fmt.Errorf("HTTP error from Yugipedia API (possibly rate-limited or banned)")
		}
		return nil, fmt.Errorf("request failed: %w: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
