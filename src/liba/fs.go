package liba

import (
    "io"
    "io/ioutil"
    "log"
    "os"
    "path"
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

func (fs *Fs) RemoveFile(filename, message string) {
    if !fs.TryRemoveFile(filename, message) {
        panic("RemoveFile failed")
    }
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

func (fs *Fs) CopyFile(src, dst, message string) {
    if !fs.TryCopyFile(src, dst, message) {
        panic("CopyFile failed")
    }
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

func (fs *Fs) CopyDir(src, dst, message string) {
    if !fs.TryCopyDir(src, dst, message) {
        panic("CopyDir failed")
    }
}

func (fs *Fs) TryCopyDir(src, dst, message string) bool {
    // get properties of source dir
    fi, err := os.Stat(src)
    if err != nil {
        fs.logger.Printf(message, err)
        return false
    }

    if !fi.IsDir() {
        fs.logger.Printf(message, "Source is not a directory")
        return false
    }

    // ensure dest dir does not already exist
    _, err = os.Open(dst)
    if !os.IsNotExist(err) {
        fs.logger.Printf(message, "Destination already exists")
        return false
    }

    // create dest dir
    err = os.MkdirAll(dst, fi.Mode())
    if err != nil {
        fs.logger.Printf(message, err)
        return false
    }

    entries, err := ioutil.ReadDir(src)
    for _, entry := range entries {
        sfp := path.Join(src, entry.Name())
        dfp := path.Join(dst, entry.Name())
        if entry.IsDir() {
            res := fs.TryCopyDir(sfp, dfp, message)
            if !res {
                return res
            }
        } else {
            // perform copy
            res := fs.TryCopyFile(sfp, dfp, message)
            if !res {
                return res
            }
        }

    }
    return true
}

func (fs *Fs) MoveFile(src, dst, message string) {
    if !fs.TryMoveFile(src, dst, message) {
        panic("MoveFile failed")
    }
}

func (fs *Fs) TryMoveFile(src, dst, message string) bool {
    if fs.debug {
        fs.logger.Printf("<fs> Move from '%s' to '%s'", src, dst)
    }

    fs.SuspendDebug()
    defer fs.ResumeDebug()

    return fs.TryCopyFile(src, dst, message) && fs.TryRemoveFile(src, message)
}
