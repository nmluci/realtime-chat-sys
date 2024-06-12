package initutil

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/nmluci/realtime-chat-sys/internal/config"
)

func InitDirectory() {
	conf := config.Get()

	sysDir := []string{"logs", "db"}

	for _, dir := range sysDir {
		joinDir := filepath.Join(conf.FilePath, dir)

		_, err := os.Stat(joinDir)
		if os.IsNotExist(err) {
			os.Mkdir(joinDir, fs.ModeDir)
		}
	}
}
