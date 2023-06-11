package spawn

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/hung0913208/go-algorithm/lib/container"
	"github.com/hung0913208/go-algorithm/lib/telegram"
)

type ChmodType int

const (
	Readonly   ChmodType = 0644
	ReadWrite  ChmodType = 0755
	Executable ChmodType = 0777
)

type Spawn interface {
	container.RestModule
}

type Command struct {
	File string   `json:"file"`
	Args []string `json:"args"`
}

type Binary struct {
	Path  string `json:"path"`
	Chmod int    `json:"chmod"`
}

type Target struct {
	Name     string   `json:"name"`
	Command  Command  `json:"command"`
	Binaries []Binary `json:"binaries"`
}

type spawnImpl struct {
	wg        sync.WaitGroup
	root      string
	broker    telegram.Telegram
	writer    http.ResponseWriter
	reader    *http.Request
	timeout   time.Duration
	processes map[string]*exec.Cmd
}

func NewSpawnModule(dbConfig, root string) Spawn {
	var broker telegram.Telegram

	if len(os.Getenv("TELEGRAM_TOKEN")) > 0 {
		broker = telegram.NewTelegram(os.Getenv("TELEGRAM_TOKEN"))
	}

	return &spawnImpl{
		broker: broker,
		root:   root,
	}, nil
}

func (self *spawnImpl) GetTimeout() time.Duration {
	return self.timeout
}

func (self *spawnImpl) SetResponseWriter(writer http.ResponseWriter) {
	self.writer = writer
}

func (self *spawnImpl) SetRequestReader(reader *http.Request) {
	self.reader = reader
}
