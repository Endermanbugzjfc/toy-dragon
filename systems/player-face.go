package systems

import (
	"github.com/df-mc/dragonfly/server/player"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"server/utils"
	"strings"
)

func SavePlayerFace(p *player.Player) error {
	path := GetFaceFilePath(p)

	_ = os.MkdirAll(filepath.Dir(path), os.ModePerm)

	var size int
	if p.Skin().Bounds().Max.X < 128 {
		size = 8
	} else {
		size = 16
	}
	bounds := image.NewRGBA(image.Rect(0, 0, size, size))
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			bounds.Set(x, y, p.Skin().At(size+x, size+y))
		}
	}

	stream, err := os.Create(path)
	//goland:noinspection GoUnhandledErrorResult
	defer stream.Close()
	if err != nil {
		return err
	}

	err2 := png.Encode(stream, bounds)
	if err2 != nil {
		return err
	}
	return nil
}

func GetFaceFilePath(p *player.Player) string {
	name := p.Name()
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return filepath.Join(utils.Conf.Player.FaceCacheFolder, name+".png")
}
