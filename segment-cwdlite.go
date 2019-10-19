package main

import (
	"os"
	"strings"
)

func cwdToLitePathSegments(p *powerline, cwd string) []pathSegment {
	pathSegments := make([]pathSegment, 0)

	home, _ := os.LookupEnv("HOME")
	if strings.HasPrefix(cwd, home) {
		cwd = "~" + cwd[len(home):]
	}

	cwd = strings.Trim(cwd, "/")

	pathSegments = append(pathSegments, pathSegment{
		path: cwd,
	})

	return maybeAliasPathSegments(p, pathSegments)
}

func segmentCwdLite(p *powerline) {
	cwd := p.cwd
	if cwd == "" {
		cwd, _ = os.LookupEnv("PWD")
	}

	if *p.args.CwdMode == "plain" {
		home, _ := os.LookupEnv("HOME")
		if strings.HasPrefix(cwd, home) {
			cwd = "~" + cwd[len(home):]
		}

		p.appendSegment("cwd", segment{
			content:    cwd,
			foreground: p.theme.CwdFg,
			background: p.theme.PathBg,
		})
	} else {
		pathSegments := cwdToLitePathSegments(p, cwd)

		for idx, pathSegment := range pathSegments {
			isLastDir := idx == len(pathSegments)-1
			foreground, background, special := getColor(p, pathSegment, isLastDir)

			segment := segment{
				content:    escapeVariables(p, maybeShortenName(p, pathSegment.path)),
				foreground: foreground,
				background: background,
			}

			if !special {
				if p.align == alignRight && p.supportsRightModules() && idx != 0 {
					segment.separator = p.symbolTemplates.SeparatorReverseThin
					segment.separatorForeground = p.theme.SeparatorFg
				} else if (p.align == alignLeft || !p.supportsRightModules()) && !isLastDir {
					segment.separator = p.symbolTemplates.SeparatorThin
					segment.separatorForeground = p.theme.SeparatorFg
				}
			}

			origin := "cwd-path"
			if isLastDir {
				origin = "cwd"
			}

			p.appendSegment(origin, segment)
		}
	}
}
