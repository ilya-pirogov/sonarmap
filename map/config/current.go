// DO NOT EDIT. Generated dynamically.
package config

import "time"

var Current = Sd{
	SCid: "03865a8a4dda39400bd2df13d2c4b378795da5b98aa4404a384567ad9ceda870",
	SdPart: 1,
	SdSys: "/sys/block/%s/device/cid",
	SdDev: "/dev/mmcblk[1-9]",
	DirMedia: "/media/%sp%d",
	DirZeroConfig: ".",
	Build: 59,

	FileWatch: "/Live/Large.at5",
	TimeoutChanges: 5 * time.Second,

	FileLive: "/LIVE.sl?",
	FileIsAlive: "/media/userdata/.StarMaps/is-alive",
	FileWallpaper: "/media/userdata/wallpaper/wallpaper01.jpg",
	DirLogs: "/live_logs",
	TimeoutIsAlive: 5 * time.Second,

    DataPort: "",
    CaptureTags: ParseTags(""),
}
