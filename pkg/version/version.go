package version

import (
	"fmt"
	"runtime/debug"
)

func Get() string {
	info, _ := debug.ReadBuildInfo()
	commit, date, modified := "0", "-", ""
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			commit = s.Value[:7]
		case "vcs.time":
			date = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				modified = " (modified)"
			}
		}
	}

	return fmt.Sprintf("%s-%s%s", commit, date, modified)
}
