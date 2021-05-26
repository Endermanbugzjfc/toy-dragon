package playersession

import (
	"path/filepath"
	"server/utils"
	"strings"
)

var config utils.CustomConfig

func GetFaceFile(name string) string {
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return filepath.Join(config.Notification.FaceCacheFolder, name+".png")
}
