package collector

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const metaCacheTTL = 10 * time.Minute

type AgentMeta struct {
	SystemName string
	PublicIPv4 string
	PublicIPv6 string
	HasIPv4    bool
	HasIPv6    bool
}

type metaCache struct {
	meta      AgentMeta
	expiresAt time.Time
}

var (
	cachedMeta metaCache
	metaMu     sync.Mutex
)

func CollectAgentMeta() AgentMeta {
	metaMu.Lock()
	defer metaMu.Unlock()

	if time.Now().Before(cachedMeta.expiresAt) {
		return cachedMeta.meta
	}

	meta := AgentMeta{
		SystemName: detectSystemName(),
		PublicIPv4: detectPublicIP("https://api.ipify.org"),
		PublicIPv6: detectPublicIP("https://api6.ipify.org"),
	}
	meta.HasIPv4 = meta.PublicIPv4 != ""
	meta.HasIPv6 = meta.PublicIPv6 != ""

	cachedMeta = metaCache{
		meta:      meta,
		expiresAt: time.Now().Add(metaCacheTTL),
	}

	return meta
}

func detectSystemName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		if pretty := readLinuxPrettyName(); pretty != "" {
			return pretty
		}
		return "Linux"
	default:
		if runtime.GOOS == "" {
			return "Unknown"
		}
		return strings.ToUpper(runtime.GOOS[:1]) + runtime.GOOS[1:]
	}
}

func readLinuxPrettyName() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}

	return ""
}

func detectPublicIP(url string) string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

func (m AgentMeta) PrimaryIP() string {
	if m.PublicIPv4 != "" {
		return m.PublicIPv4
	}
	return m.PublicIPv6
}

func (m AgentMeta) String() string {
	return fmt.Sprintf("system=%s ipv4=%s ipv6=%s", m.SystemName, m.PublicIPv4, m.PublicIPv6)
}
