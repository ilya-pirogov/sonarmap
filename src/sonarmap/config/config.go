package config

import "time"

type Sd struct {
    // sd card settings
    SCid string
    SdPart int
    SdSys string
    SdDev string
    DirMedia string
    FileZeroConfig string

    // flush cache settings
    FileWatch string
    TimeoutChanges time.Duration

    // alive checker settings
    FileLive string
    FileIsAlive string
    DirLogs string
    TimeoutIsAlive time.Duration
}
