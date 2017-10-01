package alivechecker

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dersebi/golang_exp/exp/inotify"
	"sonarmap/lib"
	"sonarmap/map/config"
	"sonarmap/map/sdcard"
)

var logger = log.New(os.Stdout, "SonarMap [AliveChecker]: ", log.LstdFlags|log.LUTC)
var fs = lib.NewFs(logger)

func addIsAlive(isAliveFile string) {
	var err error
	dir := path.Dir(isAliveFile)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			logger.Println("Error", err)
			return
		}
	}

	logger.Println("Touching", isAliveFile)
	_, err = os.Stat(isAliveFile)
	if os.IsNotExist(err) {
		logger.Println("File", isAliveFile, "is alive")
	}

	fp, err := os.OpenFile(isAliveFile, os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		logger.Println("Error", err)
		return
	}
	defer fp.Close()
	fp.WriteString(time.Now().String())
}

func moveLiveLog(liveLogsDir, liveFile string) {
	if !fs.TryMkdir(liveLogsDir, "Unable to create log dir: %s") {
		return
	}

	ext := path.Ext(liveFile)
	logFile := time.Now().Format("20060102_150405") + ext
	logPath := path.Join(liveLogsDir, logFile)

	if !fs.TryMoveFile(liveFile, logPath, "Unable to move live log: %s") {
		return
	}
}

func stopLiveLogging(liveLogsDir, isAliveFile, liveFile string) {
	defer fs.TryRemoveFile(isAliveFile, "Unable to remove is-alive file: %s")

	moveLiveLog(liveLogsDir, liveFile)

	logger.Println("It seems to live logging to", liveFile, "has been stopped")
}

func StartWatch(sd *sdcard.SdCard) {
	var watcher *inotify.Watcher
	var mediaDir string
	var watchDir string
	var watchFile string
	var lastLive string
	var err error

	timer := time.NewTimer(0)
	timer.Stop()
	devCh := sd.Register()

	cleanInternal()

	logger.Println("Start working")

	watcher, err = inotify.NewWatcher()
	if err != nil {
		logger.Panicln(err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			//logger.Println("Event:", ev)
			match, err := filepath.Match(filepath.Join(watchDir, "/", watchFile), ev.Name)
			if err != nil {
				logger.Println("Error:", err)
			}

			if match {
				timer.Reset(config.Current.TimeoutIsAlive)
				lastLive, err = filepath.Rel(mediaDir, ev.Name)
				addIsAlive(config.Current.FileIsAlive)
				time.Sleep(1 * time.Second)
			}

		case <-timer.C:
			stopLiveLogging(filepath.Join(mediaDir, config.Current.DirLogs),
				config.Current.FileIsAlive, filepath.Join(mediaDir, lastLive))

		case err := <-watcher.Error:
			logger.Println(err)
			time.Sleep(1 * time.Second)

		case dev := <-devCh:
			if dev == "" {
				logger.Println("Remove watcher")
				timer.Stop()
				watcher.RemoveWatch(watchDir)
				break
			}

			logger.Println("Found SD:", dev)

			mediaDir = config.Current.MediaDir(dev)
			watchDir = config.Current.WatchDir(dev)
			watchFile = config.Current.WatchFilePattern(dev)

			logger.Printf("Start watching for %s", watchDir)

			watcher, err = inotify.NewWatcher()
			if err != nil {
				logger.Panicln(err)
			}

			for i := 1; i <= 10; i++ {
				logger.Printf("Attemt #%d to add watcher...", i)
				err = watcher.AddWatch(watchDir, inotify.IN_CLOSE_WRITE|inotify.IN_CREATE|inotify.IN_MODIFY|inotify.IN_MOVE)
				if err == nil {
					logger.Println("Success!")
					break
				}

				logger.Println(err)
				time.Sleep(1 * time.Second)
			}

			cleanMedia(dev)
		}
	}
}
