package flushcache

import (
    "crypto/md5"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/dersebi/golang_exp/exp/inotify"
    "sonarmap/map/config"
    "sonarmap/map/sdcard"
)


var md5Sum [16]byte
var logger = log.New(os.Stdout, "SonarMap [FlushCache]: ", log.LstdFlags | log.LUTC)


func restartMediaDaemon() {
    cmd := exec.Command("killall", "-USR1", "MediaDaemon")
    err := cmd.Run()
    if err != nil {
        logger.Printf("Unable to restart MediaDaemon: %s", err.Error())
    }
}

func isFileChanged(filePath string) bool {
    prevMd5Sum := md5Sum
    buff := make([]byte, 4096)

    _, errStat := os.Stat(filePath)
    if os.IsNotExist(errStat) {
        md5Sum = [16]byte{}
        return prevMd5Sum != md5Sum
    }

    file, err := os.Open(filePath)
    if err != nil {
        logger.Panicln("Can't calculate MD5:", err)
        return false
    }

    hash := md5.New()

    for {
        n, err := file.Read(buff)
        hash.Write(buff[:n])

        if err == io.EOF {
            copy(md5Sum[:], hash.Sum(nil))
            return prevMd5Sum != md5Sum
        }

        if err != nil {
            logger.Panicln("Can't calculate MD5:", err)
            return false
        }
    }

    return true
}


func StartWatch(sd *sdcard.SdCard) {
    var watcher *inotify.Watcher
    var mediaDir string
    var watchDir string
    var watchFile string
    var lastChanged string
    var hasChanges bool
    var err error

    ticker := time.NewTicker(config.Current.TimeoutChanges)
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

            if match && isFileChanged(ev.Name) {
                lastChanged, err = filepath.Rel(mediaDir, ev.Name)
                hasChanges = true
            }

        case <- ticker.C:
            if hasChanges {
                logger.Printf("File %s has been changed", lastChanged)
                hasChanges = false
                restartMediaDaemon()
            }

        case err := <- watcher.Error:
            logger.Println(err)
            time.Sleep(1 * time.Second)

        case dev := <- devCh:
            if dev == "" {
                logger.Println("Remove watcher")
                watcher.RemoveWatch(watchDir)
                break
            }

            logger.Println("Found SD:", dev)

            mediaDir = fmt.Sprintf(config.Current.DirMedia, dev, config.Current.SdPart)
            watch := filepath.Join(mediaDir, config.Current.FileWatch)
            watchDir = filepath.Dir(watch)
            watchFile = filepath.Base(watch)

            logger.Printf("Start watching for %s", watchDir)

            watcher, err = inotify.NewWatcher()
            if err != nil {
                logger.Panicln(err)
            }

            for i := 1; i <= 10; i++ {
                logger.Printf("Attemt #%d to add watcher...", i)
                err = watcher.AddWatch(watchDir, inotify.IN_CLOSE_WRITE | inotify.IN_MODIFY | inotify.IN_CREATE | inotify.IN_DELETE | inotify.IN_MOVE)
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