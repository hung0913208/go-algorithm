package spawn

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/container"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/db"
	"github.com/hung0913208/telegram-bot-for-kubernetes/lib/telegram"
	"github.com/hung0913208/telegram-bot-for-kubernetes/modules/toolbox"
)

func (self *spawnImpl) Init(timeout time.Duration) error {
	dbModule, err := container.Pick(self.dbConfig)
	if err != nil {
		return err
	}

	dbConn, err := db.Establish(dbModule)
	if err != nil {
		return err
	}

	err = self.initSpawnTargets(dbConn, timeout)
	if err != nil {
		return err
	}

	self.broker = telegram.NewTelegram(os.Getenv("TELEGRAM_TOKEN"))
	return nil
}

func (self *spawnImpl) Deinit() error {
	for _, cmd := range self.processes {
		go func(cmd *exec.Cmd) {
			time.Sleep(60 * time.Second)
			cmd.Process.Signal(os.Kill)
		}(cmd)

		cmd.Process.Signal(os.Interrupt)
	}

	self.wg.Wait()
	return nil
}

func (self *spawnImpl) initSpawnTargets() error {
	var setting toolbox.SettingModel
	var targets []Target

	resp := dbConn.Where("name = ?", "spawn").
		First(&setting)
	if resp.Error != nil {
		return resp.Error
	}

	err := json.Unmarshal([]byte(setting.Value), &targets)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if _, ok := self.processes[target.Name]; ok {
			continue
		}

		command := make([]string, 0)

		for _, binary := range target.Binaries {
			path, err := downloadBinary(binary.Path)
			if err != nil {
				return err
			}

			err = os.Chmod(path, int(binary.Chmod))
			if err != nil {
				return err
			}

			if binary.Chmod == Executable && target.Command.File == binary.Path {
				proc := exec.Command(path, tagret.Command.Args...)

				self.processes[target.Name] = proc
				self.wg.Add(1)

				go func(proc *exec.Cmd) {
					self.wg.Done()

					if err := proc.Run(); err != nil {
						self.logger.Errorf("[%s] %v", target.Name, err)
					}
				}(proc)
				break
			}
		}

	}

	return nil
}

func (self *spawnImpl) downloadBinary(url string) (string, error) {
	name := url[strings.LastIndex(url, "/")+1:]
	outputPath := fmt.Sprintf("%s/%s", self.root, name)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return outputPath, err
}
