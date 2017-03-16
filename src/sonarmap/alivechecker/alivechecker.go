package alivechecker

import (
    "fmt"
    "os"
    "sonarmap/config"
    "sonarmap/sdcard"
    "time"
    "github.com/dersebi/golang_exp/exp/inotify"
    "log"
    "path/filepath"
    "path"
    "io"
)


var logger = log.New(os.Stdout, "SonarMap [AliveChecker]: ", log.LstdFlags | log.LUTC)


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

    fp, err := os.OpenFile(isAliveFile, os.O_CREATE | os.O_TRUNC, 0666)
    if err != nil {
        logger.Println("Error", err)
        return
    }
    defer fp.Close()
    fp.WriteString(time.Now().String())
}

func removeIsAlive(liveLogsDir, isAliveFile, liveFile string) {
    defer func() {
        logger.Printf("Removing %s", isAliveFile)
        err := os.Remove(isAliveFile)
        if err != nil {
            logger.Println("Error! Can't remove file", isAliveFile)
        }
    }()

    var err error
    _, err = os.Stat(liveLogsDir)
    if err != nil {
        err = os.Mkdir(liveLogsDir, 0777)
    }

    _, err = os.Stat(liveLogsDir)
    if os.IsNotExist(err) {
        os.Mkdir(liveLogsDir, 0755)
    }

    ext := path.Ext(liveFile)
    logFile := time.Now().Format("20060102_150405") + ext
    logPath := path.Join(liveLogsDir, logFile)

    logger.Printf("Moving %s to %s", liveFile, logPath)

    in, err := os.Open(liveFile)
    if err != nil {
        logger.Println(err)
        return
    }
    defer in.Close()

    out, err := os.Create(logPath)
    if err != nil {
        logger.Println(err)
        return
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    if err != nil {
        logger.Println(err)
        return
    }

    err = in.Close()
    if err != nil {
        logger.Println(err)
        return
    }

    err = os.Remove(liveFile);
    if err != nil {
        logger.Println(err)
        return
    }

    logger.Println("File", liveFile, "has been stopped")
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
    logger.Println("Start working")

    watcher, err = inotify.NewWatcher()
    if err != nil {
        logger.Panicln(err)
    }

    for {
        select {
        case ev := <- watcher.Event:
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

        case <- timer.C:
            removeIsAlive(filepath.Join(mediaDir, config.Current.DirLogs),
                config.Current.FileIsAlive, filepath.Join(mediaDir, lastLive))

        case err := <- watcher.Error:
            logger.Println(err)
            time.Sleep(1 * time.Second)

        case dev := <- devCh:
            if dev == "" {
                logger.Println("Remove watcher")
                timer.Stop()
                watcher.RemoveWatch(watchDir)
                break
            }

            logger.Println("Found SD:", dev)

            mediaDir = fmt.Sprintf(config.Current.DirMedia, dev, config.Current.SdPart)
            watch := filepath.Join(mediaDir, config.Current.FileLive)
            watchDir = filepath.Dir(watch)
            watchFile = filepath.Base(watch)

            removeIsAlive(filepath.Join(mediaDir, config.Current.DirLogs),
                config.Current.FileIsAlive, filepath.Join(mediaDir, lastLive))

            logger.Printf("Start watching for %s", watchDir)

            watcher, err = inotify.NewWatcher()
            if err != nil {
                logger.Panicln(err)
            }

            for i := 1; i <= 10; i++ {
                logger.Printf("Attemt #%d to add watcher...", i)
                err = watcher.AddWatch(watchDir, inotify.IN_CLOSE_WRITE | inotify.IN_CREATE | inotify.IN_MODIFY | inotify.IN_MOVE)
                if err == nil {
                    logger.Println("Success!")
                    break
                }

                logger.Println(err)
                time.Sleep(1 * time.Second)
            }
        }
    }
}
