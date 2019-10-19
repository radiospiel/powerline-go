package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

/*
 * iterate through an array of strings, returning the first one which
 * matches the yield() function.
 */
func detect(strings []string, yield func(string) bool) string {
	for _, str := range strings {
		if yield(str) {
			return str
		}
	}

	var null string
	return null
}

type batteryStatus struct {
	percentage    string
	ac_attached   bool
	charging      bool
	estimatedTime string
}

func getBatteryStatusLineFromPmset() (string, error) {
	b, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`InternalBattery`)
	line := detect(strings.Split(string(b), "\n"), func(line string) bool {
		return re.MatchString(line)
	})

	return line, nil
}

func compileRegexp(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

func firstMatch(re string, str string) (string, bool) {
	var r string
	matches := compileRegexp(re).FindAllStringSubmatch(str, -1)
	if matches == nil {
		return r, false
	} else if len(matches[0]) == 0 {
		return r, false
	} else {
		return matches[0][0], true
	}
}

func getBatteryStatus() (batteryStatus, error) {
	var bs batteryStatus

	batteryStatusLine, err := getBatteryStatusLineFromPmset()
	if err != nil {
		return bs, err
	}

	// batteryStatusLine can be something like
	//
	// `-InternalBattery-0 (id=4325475)	10%; charging; 2:17 remaining present: true`
	// `-InternalBattery-0 (id=4325475)	9%; discharging; (no estimate) present: true`
	// `-InternalBattery-0 (id=4325475)	14%; AC attached; not charging present: true`
	//
	// Note the tab before the percentage. We split this on "\t", and then the right part on ";"
	//

	parts := strings.Split(batteryStatusLine, "\t")

	if len(parts) < 2 {
		err := errors.New("Cannot parse battery_line")
		return bs, err
	}

	batteryStatusParts := regexp.MustCompile(`; `).Split(parts[1], -1)

	for _, str := range batteryStatusParts {
		if match, ok := firstMatch(`\d+%`, str); ok {
			bs.percentage = match
		} else if match, ok = firstMatch(`AC attached`, str); ok {
			bs.ac_attached = true
		} else if match, ok = firstMatch(`(not charging)|(discharging)`, str); ok {
			bs.charging = false
		} else if match, ok = firstMatch(`charging`, str); ok {
			bs.charging = true
		} else if match, ok = firstMatch(`\d+:\d+`, str); ok {
			// sometimes pmset reports "0:00", which is obviously nonsense
			if match != "0:00" {
				bs.estimatedTime = match
			}
		}
	}

	return bs, nil
}

func parseEstimatedTime(str string) int {
	m := regexp.MustCompile(`(\d+):(\d+)`).FindAllStringSubmatch(str, -1)
	if m != nil {
		hours, _ := strconv.ParseInt(m[0][0], 10, 32)
		minutes, _ := strconv.ParseInt(m[0][1], 10, 32)
		return int(hours)*60 + int(minutes)
	}

	return 0
}

func segmentBatteryLabel(bs batteryStatus) string {
	if bs.estimatedTime != "" {
		return fmt.Sprintf("%s (%s)", bs.percentage, bs.estimatedTime)
	} else {
		return fmt.Sprintf("%s", bs.percentage)
	}
}

func segmentBattery(p *powerline) {
	bs, err := getBatteryStatus()
	if err != nil {
		// log.Fatal(err)
		return
	}

	type batteryStatus struct {
		percentage    string
		ac_attached   bool
		charging      bool
		estimatedTime string
	}

	// If you want to deal with theme settings based on the estimatedTime
	// estimatedTimeMinutes := parseEstimatedTime(bs.estimatedTime)

	if bs.charging || bs.ac_attached {
		p.appendSegment("battery", segment{
			content: segmentBatteryLabel(bs),
			// background: p.theme.GitStashedBg,
			// foreground: p.theme.GitStashedFg,
			background: p.theme.GitAheadBg,
			foreground: p.theme.GitAheadFg,
		})
	} else {
		p.appendSegment("battery", segment{
			content:    segmentBatteryLabel(bs),
			background: p.theme.GitUntrackedBg,
			foreground: p.theme.GitUntrackedFg,
		})
	}
}
