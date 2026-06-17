package ops

import (
	"fmt"
	"os"
)

func makeBackup(path string, backupType string) string {
	if _, err := os.Lstat(path); err != nil {
		return ""
	}
	if backupType == "" || backupType == "none" || backupType == "off" {
		return ""
	}

	suffix := backupSuffix("")
	backupPath := path + suffix

	switch backupType {
	case "numbered", "t":
		n := 1
		for {
			candidate := fmt.Sprintf("%s.~%d~", path, n)
			if _, err := os.Stat(candidate); err != nil {
				backupPath = candidate
				break
			}
			n++
		}
	case "existing", "nil":
		n := 1
		hasNumbered := false
		for {
			candidate := fmt.Sprintf("%s.~%d~", path, n)
			if _, err := os.Stat(candidate); err != nil {
				if !hasNumbered {
					backupPath = path + suffix
				} else {
					backupPath = candidate
				}
				break
			}
			hasNumbered = true
			n++
		}
	case "never", "simple":
		backupPath = path + suffix
	default:
		backupPath = path + suffix
	}

	os.Rename(path, backupPath)
	return backupPath
}

func backupSuffix(custom string) string {
	if custom != "" {
		return custom
	}
	if env := os.Getenv("SIMPLE_BACKUP_SUFFIX"); env != "" {
		return env
	}
	return "~"
}
