package path

import "strings"

func isValidPath(s string) bool {
	// If input string contains control characters, consider it invalid.
	if strings.ContainsFunc(s, func(r rune) bool {
		return 0x01 <= r && r <= 0x1F
	}) {
		return false
	}

	// Extract and remove drive letter if present.
	d, ok := extractDriveLetter(s)
	if ok {
		s = s[len(d):]
	}

	// Extract and remove device path prefix if present.
	d, ok = extractDevicePathPrefix(s)
	if ok {
		s = s[len(d):]
	}

	// If input string contains Windows specific invalid characters, consider it invalid.
	return !strings.ContainsAny(s, `<>:"|?*`)
}

// extractDriveLetter extracts the drive letter from input string.
// If successful, it returns the drive letter and true.
// Otherwise, it returns an empty string and false.
func extractDriveLetter(s string) (string, bool) {
	if len(s) < 2 {
		return "", false
	}

	// Check if the first two characters represent a drive letter (e.g., "C:").
	if ('A' <= s[0] && s[0] <= 'Z') || ('a' <= s[0] && s[0] <= 'z') {
		if s[1] == ':' {
			// Return the drive letter (e.g., "C:") and true.
			return s[0:2], true
		}
	}
	return "", false
}

// extractDevicePathPrefix extracts the device path prefix from input string.
// If successful, it returns the device path prefix and true.
// Otherwise, it returns an empty string and false.
func extractDevicePathPrefix(s string) (string, bool) {
	// Check for device path prefixes (e.g., "\\.\", "\\?\").
	if strings.HasPrefix(s, `\\.\`) || strings.HasPrefix(s, `\\?\`) {
		// Return the device path prefix (e.g., "\\.\", "\\?\") and true.
		return s[0:4], true
	}
	return "", false
}
