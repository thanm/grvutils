package testutils

import (
	"bufio"
	"fmt"
	"strings"
)

func split(s string) []string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanWords)
	var res []string
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}
	return res
}

func Check(actual string, expected string) string {
	vactual := split(actual)
	vexpected := split(expected)
	reason := ""
	if len(vactual) != len(vexpected) {
		reason = fmt.Sprintf("lengths differ (have %d want %d)",
			len(vactual), len(vexpected))
	} else {
		for i := 0; i < len(vactual); i += 1 {
			if vactual[i] != vexpected[i] {
				reason = fmt.Sprintf("diff at slot %d (have %s want %s)",
					i, vactual[i], vexpected[i])
			}
		}
	}
	if reason == "" {
		return ""
	}
	return fmt.Sprintf("%s\nactual='%s'  wanted '%s':", reason,
		actual, expected)
}
