package config

import (
    "fmt"
    "path/filepath"
)

func (cfg *Sd) MediaDir(dev string) string {
    return fmt.Sprintf(cfg.DirMedia, dev, cfg.SdPart)
}

func (cfg *Sd) WatchPathPattern(dev string) string {
    return filepath.Join(cfg.MediaDir(dev), cfg.FileLive)
}

func (cfg *Sd) WatchDir(dev string) string {
    return filepath.Dir(cfg.WatchPathPattern(dev))
}

func (cfg *Sd) WatchFilePattern(dev string) string {
    return filepath.Base(cfg.WatchPathPattern(dev))
}

func (cfg *Sd) MediaDirLogs(dev string) string {
    return filepath.Join(cfg.MediaDir(dev), cfg.DirLogs)
}
