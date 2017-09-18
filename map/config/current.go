// DO NOT EDIT. Generated dynamically.
package config

import "time"

var Current = Sd{
	SCid: "bfd7f660609c814ba4cfb47497476a37f1503e5931c25b42999331cfffe5c2f0",
	SdPart: 1,
	SdSys: "/sys/block/%s/device/cid",
	SdDev: "/dev/mmcblk[1-9]",
	DirMedia: "/media/%sp%d",
	DirZeroConfig: ".",
	Build: 23,

	FileWatch: "/Live/Large.at5",
	TimeoutChanges: 5 * time.Second,

	FileLive: "/LIVE.sl?",
	FileIsAlive: "/media/userdata/.StarMaps/is-alive",
	FileWallpaper: "/media/userdata/wallpaper/wallpaper01.jpg",
	DirLogs: "/live_logs",
	TimeoutIsAlive: 5 * time.Second,

    DataPort: "123",
    CaptureTags: ParseTags("GGA,GLL,DBT,DPT,SDF"),
}
