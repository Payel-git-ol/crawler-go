package utils

import "strings"

func IsOpenLicense(licenseKey string) bool {
	openLicenses := []string{"mit", "apache", "gpl", "bsd", "mpl", "epl"}
	lowerKey := strings.ToLower(licenseKey)
	for _, l := range openLicenses {
		if strings.Contains(lowerKey, l) {
			return true
		}
	}
	return false
}
