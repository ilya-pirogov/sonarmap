package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"path"
	"time"

	"sonarmap/firmware/api"
	"sonarmap/firmware/bindata"
	"sonarmap/firmware/config"
	"sonarmap/firmware/shells"
	"sonarmap/firmware/structs"
	"sonarmap/firmware/utils"
	sonarmap "sonarmap/map/config"
	srpc "sonarmap/map/rpc"
)

const (
	srvAddr             = "239.2.1.1:2052"
	maxDatagramSize     = 8192
	timeoutReadSettings = 15 * time.Second
	timeoutSuccessFlash = 30 * time.Second
	timeoutFailFlash    = 5 * time.Second
)

func watchSettings(ip string, settingsC chan *structs.Settings) {
	addr, err := net.ResolveUDPAddr("udp", ip)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := utils.ListenUdpMulticast(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	log.Println(addr)

	listener.SetReadBuffer(maxDatagramSize)

	buffer := make([]byte, maxDatagramSize)
	for {
		_, _, err := listener.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}

		s := string(buffer[:])
		log.Println(s)
		//
		dec := json.NewDecoder(bytes.NewReader(buffer))
		var settings *structs.Settings = &structs.Settings{}
		dec.Decode(settings)
		settingsC <- settings
	}
}

func tryFlash(settings *structs.Settings) bool {
	var (
		err               error
		sonarMap          []byte
		wallpaperAsset    []byte
		translations      []string
		translationAssets []api.AssetFile
	)

	log.Println("Start flashing...")

	ssh := shells.NewTelnetShell(settings.IP, config.Username, config.Password)
	a := api.New(ssh)

	if a.GetVersion() >= sonarmap.Current.Build {
		log.Println("Alread flashed")
		return true
	}

	sonarMap, err = bindata.Asset("sonarmap")
	if err != nil {
		log.Println("Error: ", err)
		return false
	}

	wallpaperAsset, err = bindata.Asset("wallpaper.jpg")
	if err != nil {
		log.Println("Error: ", err)
		return false
	}

	translations, err = bindata.AssetDir("translations")
	if err != nil {
		log.Println("Error: ", err)
		return false
	}
	translationAssets = make([]api.AssetFile, 0, len(translations))
	for _, v := range translations {
		tmpData, err := bindata.Asset(path.Join("translations", v))
		if err != nil {
			log.Println("Error: ", err)
			return false
		}
		tmpAsset := &api.AssetFile{tmpData, v}
		translationAssets = append(translationAssets, *tmpAsset)
	}

	if err = a.StopService(); err != nil {
		log.Println("StopService error:", err)
		return false
	}

	if err = a.UploadSonarMap(sonarMap); err != nil {
		log.Println("UploadSonarMap error:", err)
		return false
	}

	if err = a.UploadWallpaper(wallpaperAsset); err != nil {
		log.Println("UploadWallpaper error:", err)
		return false
	}

	if err = a.UploadTranslations(translationAssets); err != nil {
		log.Println("UploadTranslations error:", err)
		return false
	}

	if err = a.PatchRc(); err != nil {
		log.Println("PatchRc error:", err)
		return false
	}

	if err = a.SetVersion(sonarmap.Current.Build); err != nil {
		log.Println("SetVersion error:", err)
		return false
	}

	//if err = a.ChangePassword("c7195563b28d9ffff104342dcb5d4cb7"); err != nil {
	//    log.Println("ChangePassword error:", err)
	//    return false
	//}

	if err = a.PowerOff(); err != nil {
		log.Println("PowerOff error:", err)
		return false
	}

	return true
}

func IsFlashed(ip string) bool {
	client, err := rpc.DialHTTP("tcp", ip+":7654")
	if err != nil {
		//log.Println(err)
		return false
	}

	args := &srpc.GetVersionArgs{}
	reply := srpc.GetVersionReply{}

	err = client.Call("SonarRpc.GetVersion", args, &reply)
	if err != nil {
		log.Println("RPC Error: ", err)
		return false
	}

	return reply.Version > 0 && reply.IsValidSd
}

func CreateZeroconfig(ip string) {
	var (
		fileName string
		err      error
		fp       *os.File
	)

	_, err = os.Stat(sonarmap.Current.DirZeroConfig)
	if os.IsNotExist(err) {
		os.MkdirAll(sonarmap.Current.DirZeroConfig, 0755)
	}

	fileName = path.Join(sonarmap.Current.DirZeroConfig, ip)

	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		fp, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		fp, err = os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0644)
	}

	if err != nil {
		log.Println("WriteZeroconfig error", err)
		return
	}
	defer fp.Close()

	fp.WriteString(time.Now().UTC().String())
}

func DeleteZeroconfig(ip string) {
	var (
		fileName string
		err      error
	)

	_, err = os.Stat(sonarmap.Current.DirZeroConfig)
	if os.IsNotExist(err) {
		return
	}

	fileName = path.Join(sonarmap.Current.DirZeroConfig, ip)

	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}

	err = os.Remove(fileName)

	if err != nil {
		log.Println("WriteZeroconfig error", err)
		return
	}
}

func main() {
	var (
		settings    *structs.Settings
		zeroConfigs = make(map[string]time.Time)
	)

	settings = &structs.Settings{}

	println(settings.IP)

	log.Printf("Start working. Current version: %d", sonarmap.Current.Build)

	sigC := make(chan os.Signal, 1)
	settingsC := make(chan *structs.Settings)
	flashTimer := time.NewTimer(timeoutReadSettings)

	signal.Notify(sigC, os.Interrupt)

	go watchSettings(srvAddr, settingsC)

	reset := func(d time.Duration) {
		log.Println("Reset timer:", d)
		if !flashTimer.Stop() {
			select {
			case <-flashTimer.C:
			default:
			}
		}
		flashTimer.Reset(d)
	}

	for {
		select {
		case <-sigC:
			println("\rInterrupted. Exiting...")
			return
		case settings = <-settingsC:
			log.Println("Detected IP: " + settings.IP)
			if IsFlashed(settings.IP) {
				zeroConfigs[settings.IP_Zeroconfig] = time.Now()
				CreateZeroconfig(settings.IP_Zeroconfig)
			}

			for ip, ts := range zeroConfigs {
				if time.Now().Sub(ts) > 15*time.Second {
					delete(zeroConfigs, ip)
					DeleteZeroconfig(ip)
				}
			}

			// reset(timeoutReadSettings)
		case <-flashTimer.C:
			for ip, ts := range zeroConfigs {
				if time.Now().Sub(ts) > 15*time.Second {
					delete(zeroConfigs, ip)
					DeleteZeroconfig(ip)
				}
			}

			if settings.IP == "" {
				reset(timeoutFailFlash)
				continue
			}
			if tryFlash(settings) {
				log.Println("The device has been successfull flashed")
				reset(timeoutSuccessFlash)
			} else {
				log.Println("Fail to flash the device")
				reset(timeoutFailFlash)
			}
		}
	}
}
