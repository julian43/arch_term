package network

import (
	"fmt"
	"net"
	"sort"
	"strings"
)

// TailscaleIPRange is the RFC for Tailscale's IP allocation
const TailscaleIPRange = "100.64.0.0/10"

// NetworkInfo holds information about available network interfaces and Tailscale connectivity.
type NetworkInfo struct {
	LocalIPs     []string
	TailscaleIPs []string
	HasTailscale bool
	BadgerURL    string // Tailscale HTTP API base URL (for future use)
}

// GetNetworkInfo discovers available network interfaces and detects Tailscale connectivity.
func GetNetworkInfo() (*NetworkInfo, error) {
	info := &NetworkInfo{}

	// Get all network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate network interfaces: %w", err)
	}

	// Track the Tailscale subnet allocator CIDR
	tsCIDR := &net.IPNet{}
	_, tsCIDR, err = net.ParseCIDR(TailscaleIPRange)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Tailscale IP range: %w", err)
	}

	for _, iface := range ifaces {
		// Skip down or loopback interfaces (loopback is not Tailscale)
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.String()
			info.LocalIPs = append(info.LocalIPs, ip)

			// Check if this IP is in the Tailscale range (100.64.0.0/10)
			if tsCIDR.Contains(ipNet.IP) {
				info.TailscaleIPs = append(info.TailscaleIPs, ip)
				info.HasTailscale = true
			}
		}
	}

	// Sort for deterministic output
	sort.Strings(info.LocalIPs)
	sort.Strings(info.TailscaleIPs)

	return info, nil
}

// GetTailscaleIP returns the first Tailscale IP if one exists, empty string otherwise.
func GetTailscaleIP() string {
	info, err := GetNetworkInfo()
	if err != nil || len(info.TailscaleIPs) == 0 {
		return ""
	}
	return info.TailscaleIPs[0]
}

// GetTailscaleIPs returns all Tailscale IPs.
func GetTailscaleIPs() []string {
	info, err := GetNetworkInfo()
	if err != nil {
		return nil
	}
	return info.TailscaleIPs
}

// BuildAccessURL constructs a URL that a remote client would use to access
// this ArchTerm instance over Tailscale.
func BuildAccessURL(port int) string {
	tsIP := GetTailscaleIP()
	if tsIP != "" {
		return fmt.Sprintf("http://%s:%d", tsIP, port)
	}
	return ""
}

// PrintStartupBanner prints a human-friendly startup banner with
// Tailscale connectivity information if available.
func PrintStartupBanner(port int) {
	const boxWidth = 54
	header := "ArchTerm Network Info"
	headerLine := fmt.Sprintf("                   %s                   ", header)

	banner := fmt.Sprintf("╔%s╗\n", strings.Repeat("═", boxWidth))
	banner += fmt.Sprintf("║   %s   ║\n", headerLine)
	banner += fmt.Sprintf("╠%s╣\n", strings.Repeat("═", boxWidth))

	// Tailscale section
	tsIP := GetTailscaleIP()
	if tsIP != "" {
		tsLabel := fmt.Sprintf("🦉 Tailscale IP: %s", tsIP)
		tsURL := fmt.Sprintf("🌐 Access URL:    http://%s:%d", tsIP, port)
		banner += fmt.Sprintf("║   %s  ║\n", leftPad(tsLabel, boxWidth))
		banner += fmt.Sprintf("║   %s  ║\n", leftPad(tsURL, boxWidth))
		banner += "║                                                      ║\n"
	} else {
		banner += "║   ⚠️  No Tailscale interface detected.                    ║\n"
		banner += "║    Run on a machine connected to Tailscale, or           ║\n"
		banner += fmt.Sprintf("║    start with: ./archterm -port %d -bind :%d            ║\n", port, port)
		banner += "║                                                      ║\n"
	}

	// Local interfaces (first 3 non-loopback)
	info, err := GetNetworkInfo()
	if err == nil && len(info.LocalIPs) > 0 {
		banner += "║                                                      ║\n"
		banner += "║   Local interfaces:                                   ║\n"

		count := 0
		for _, ip := range info.LocalIPs {
			// Skip loopback
			if strings.HasPrefix(ip, "127.") || strings.HasPrefix(ip, "::1") {
				continue
			}
			banner += fmt.Sprintf("║     - %s  ║\n", leftPad(ip, 22))
			count++
			if count >= 3 {
				break
			}
		}
	}

	banner += "║                                                      ║\n"
	banner += fmt.Sprintf("╚%s╝\n", strings.Repeat("═", boxWidth))

	fmt.Println(banner)
}

// leftPad pads a string to the target width with spaces
func leftPad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
