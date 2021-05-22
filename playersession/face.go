package playersession

import (
	"path/filepath"
	"server/system"
	"strings"
)

func GetFaceFile(name string) string {
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return filepath.Join(system.Config.Notification.FaceCacheFolder, name+".png")
}
