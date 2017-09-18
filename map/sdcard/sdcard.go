package sdcard

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ilya-pirogov/sonarmap/map/config"
	"github.com/ilya-pirogov/sonarmap/map/lib"
)

var logger = log.New(os.Stdout, "SonarMap [SdCard]: ", log.LstdFlags|log.LUTC)

type SdCard struct {
	dev     string
	scid    [32]byte
	lock    sync.Locker
	queue   []chan string
	timeout time.Duration
}

func (sd *SdCard) set(dev string) {
	if dev == sd.dev {
		return
	}
	logger.Printf("New val: %s", dev)

	sd.dev = dev

	for _, ch := range sd.queue {
		ch <- dev
	}
}

func (sd *SdCard) GetLast() string {
	return sd.dev
}

func (sd *SdCard) IsValid() bool {
	return sd.dev != ""
}

func New(scid string) *SdCard {
	return &SdCard{
		scid:    lib.HexToBytes(scid),
		queue:   make([]chan string, 0),
		timeout: 15 * time.Second,
		lock:    &sync.Mutex{},
	}
}

func (sd *SdCard) checkCid(dev string) (bool, error) {
	if dev == "" {
		dev = sd.dev
	}

	fp, err := os.Open(fmt.Sprintf(config.Current.SdSys, dev))
	if err != nil {
		return false, err
	}

	cid, err := bufio.NewReader(fp).ReadString('\n')
	if err != nil {
		return false, err
	}

	buffer := []byte(strings.TrimSpace(fmt.Sprintf("SM#CID:%s", cid)))
	return sha256.Sum256(buffer) == sd.scid, nil
}

func (sd *SdCard) find() (string, error) {
	matches, err := filepath.Glob(config.Current.SdDev)
	if err != nil {
		sd.set("")
		return "", err
	}

	for _, match := range matches {
		base := filepath.Base(match)
		res, err := sd.checkCid(base)

		if err != nil {
			sd.set("")
			return "", err
		}

		if res {
			sd.set(base)
			return base, nil
		}
	}
	sd.set("")
	return "", nil // lib.SdCardNotFound
}

func (sd *SdCard) hasErr(err error) bool {
	if err != nil {
		sd.set("")

		logger.Println("Error:", err)

		time.Sleep(sd.timeout)
		return true
	}
	return false
}

func (sd *SdCard) Register() chan string {
	ch := make(chan string, 128)

	if sd.dev != "" {
		ch <- sd.dev
	}

	sd.lock.Lock()
	defer sd.lock.Unlock()
	sd.queue = append(sd.queue, ch)

	return ch
}

func (sd *SdCard) Watch() {
	isCorrect := false

	for {
		dev, err := sd.find()
		if sd.hasErr(err) || dev == "" {
			if err != nil {
				logger.Println(err)
			}
			if isCorrect {
				isCorrect = false
				logger.Println("SD Card has been removed")
			}
		} else {
			if !isCorrect {
				isCorrect = true
				logger.Println("SD Card has the correct CID")
			}
		}

		time.Sleep(5 * time.Second)
	}
}
