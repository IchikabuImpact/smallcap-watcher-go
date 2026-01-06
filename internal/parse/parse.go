package parse

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	previousCloseRe = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)`)
	commaRe         = regexp.MustCompile(`,`)
)

func ParseNumeric(value string) (float64, bool) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return 0, false
	}

	multiplier := 1.0
	switch {
	case strings.Contains(clean, "兆"):
		multiplier = 1e12
		clean = strings.ReplaceAll(clean, "兆", "")
	case strings.Contains(clean, "億"):
		multiplier = 1e8
		clean = strings.ReplaceAll(clean, "億", "")
	case strings.Contains(clean, "万"):
		multiplier = 1e4
		clean = strings.ReplaceAll(clean, "万", "")
	}

	clean = strings.NewReplacer("円", "", "%", "", "倍", "", "株", "").Replace(clean)
	clean = commaRe.ReplaceAllString(clean, "")
	clean = strings.TrimSpace(clean)

	if clean == "" {
		return 0, false
	}

	val, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, false
	}

	return val * multiplier, true
}

func ParsePreviousClose(value string) (float64, bool) {
	matches := previousCloseRe.FindStringSubmatch(value)
	if len(matches) < 2 {
		return 0, false
	}
	return ParseNumeric(matches[1])
}
