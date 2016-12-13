package main

import (
	"strings"
)

// parse comam separated tags into a string slice
func parseTags(raw string) (tags []string) {
	for _, tag := range strings.Split(raw, ",") {
		tag = strings.ToLower(strings.TrimSpace(tag))

		if tag == "" {
			continue
		}

		tags = append(tags, tag)
	}

	return
}
