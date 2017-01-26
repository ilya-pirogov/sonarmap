// DO NOT EDIT. Generated dynamically.
package config

import "time"

var Current = Sd{
	SCid: "bfd7f660609c814ba4cfb47497476a37f1503e5931c25b42999331cfffe5c2f0",
	SdPart: 1,
	SdSys: "/sys/block/%s/device/cid",
	SdDev: "/dev/mmcblk[1-9]",
	DirMedia: "/media/%sp%d",

	FileWatch: "/Live/Large.at5",
	TimeoutChanges: 5 * time.Second,

	FileLive: "/LIVE.sl?",
	FileIsAlive: "/media/userdata/.StarMaps/is-alive",
	DirLogs: "/live_logs",
	TimeoutIsAlive: 5 * time.Second,
}
