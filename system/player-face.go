package system

import (
	"path/filepath"
	"server/utils"
	"strings"
)

func GetFaceFile(name string) string {
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return filepath.Join(utils.Config.Player.FaceCacheFolder, name+".png")
}
