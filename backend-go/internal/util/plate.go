package util

import (
	"fmt"
	"regexp"
	"strings"

	"plat-detection-system/backend-go/internal/model"
)

var plateRe = regexp.MustCompile(`^([A-Z]{1,2})(\d{1,4})([A-Z]{1,3})$`)

func NormalizePlateRaw(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	return s
}

func SplitAndFormatPlate(raw string) (model.PlateComponents, string) {
	raw = NormalizePlateRaw(raw)
	m := plateRe.FindStringSubmatch(raw)
	if len(m) != 4 {
		return model.PlateComponents{}, raw
	}
	c := model.PlateComponents{Prefix: m[1], Number: m[2], Suffix: m[3]}
	return c, fmt.Sprintf("%s %s %s", c.Prefix, c.Number, c.Suffix)
}

