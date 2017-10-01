package alivechecker

import (
    "os"
    "path"
    "path/filepath"
    "time"

    "sonarmap/map/config"
)

func cleanInternal() {
    logger.Println("Start cleaning internal memory")
    fs.TryRemoveFile(config.Current.FileIsAlive, "Unable to remove is-alive file: %s")
}

func cleanMedia(dev string) {
    logger.Printf("Start cleaning SDCard: %s", dev)

    pattern := config.Current.WatchPathPattern(dev)
    lives, err := filepath.Glob(pattern)
    if err != nil {
        logger.Printf("Incorrect pattern %s. Error: %s", pattern, err)
        return
    }

    for _, live := range lives {
        if live == "." || live == ".." || path.Ext(live) == "" {
            continue
        }

        moveLiveLog(config.Current.MediaDirLogs(dev), live)
        time.Sleep(2 * time.Millisecond)
    }

    zipFile := filepath.Join(config.Current.WatchDir(dev), "LIVE.zip")
    _, err = os.Stat(zipFile)
    if err == nil {
        fs.TryRemoveFile(zipFile, "Unable to remove LIVE.zip: %s")
    }
}
