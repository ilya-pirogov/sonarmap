package liba

import (
    "os"
    "log"
    "io"
)

type Fs struct {
    logger *log.Logger
    debug  bool
    prevDebug bool
}

func NewFs(logger *log.Logger) *Fs {
    return &Fs{logger: logger, debug: true}
}

func (fs *Fs) SuspendDebug() {
    fs.prevDebug = fs.debug
    fs.debug = false
}

func (fs *Fs) ResumeDebug() {
    fs.debug = fs.prevDebug
}

func (fs *Fs) TryMkdir(dir, message string) bool {
    if fs.debug {
        fs.logger.Printf("<fs> Create dir: %s", dir)
    }

    _, err := os.Stat(dir)
    if os.IsNotExist(err) {
        err = os.Mkdir(dir, 0755)
        if err != nil {
            fs.logger.Printf(message, err)
            return false
        }
    }
    return true
}

func (fs *Fs) TryRemoveFile(filename, message string) bool {
    if fs.debug {
        fs.logger.Printf("<fs> Remove file: %s", filename)
    }

    err := os.Remove(filename)
    if err != nil {
        fs.logger.Printf(message, err)
    }
    return err != nil
}

func (fs *Fs) TryCopyFile(src, dst, message string) bool {
    if fs.debug {
        fs.logger.Printf("<fs> Copy from '%s' to '%s'", src, dst)
    }

    in, err := os.Open(src)
    if err != nil {
        fs.logger.Printf(message, err)
        return false
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        fs.logger.Printf(message, err)
        return false
    }
    defer out.Close()

    _, err = io.Copy(out, in)
    if err != nil {
        fs.logger.Printf(message, err)
        return false
    }

    err = in.Close()
    if err != nil {
        fs.logger.Printf(message, err)
        return false
    }

    return true
}

func (fs *Fs) TryMoveFile(src, dst, message string) bool {
    if fs.debug {
        fs.logger.Printf("<fs> Move from '%s' to '%s'", src, dst)
    }

    fs.SuspendDebug()
    defer fs.ResumeDebug()

    return fs.TryCopyFile(src, dst, message) && fs.TryRemoveFile(src, message)
}
