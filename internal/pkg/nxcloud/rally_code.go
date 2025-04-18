package nxcloud

import "strings"

func IsOfficialRallyCode(rallyCode string) bool {
	return strings.HasSuffix(rallyCode, "00000")
}

func ParseRallyCode(rallyCode string) (channel, language string, generation int) {
	panic("implement me")
}
