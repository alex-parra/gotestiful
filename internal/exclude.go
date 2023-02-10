package internal

import (
	"fmt"
	"regexp"
)

func excludePackages(packages, excludes []string) ([]string, []string, error) {
	if len(excludes) == 0 {
		return packages, nil, nil
	}

	excludeRegexs := make([]*regexp.Regexp, 0, len(excludes))
	for _, exclude := range excludes {
		if exclude == "" {
			// this would catch all, probably not the intention
			continue
		}
		regex, err := regexp.Compile("^" + exclude)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot compile regex %q: %+v", exclude, regex)
		}
		excludeRegexs = append(excludeRegexs, regex)
	}

	var included, excluded []string
	for _, pkg := range packages {
		isIncluded := true
		for _, regex := range excludeRegexs {
			if regex.MatchString(pkg) {
				isIncluded = false
			}
		}
		if isIncluded {
			included = append(included, pkg)
		} else {
			excluded = append(excluded, pkg)
		}
	}
	return included, excluded, nil
}
