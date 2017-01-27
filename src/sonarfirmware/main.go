package main

import (
    "bytes"
    "encoding/json"
    "log"
    "net"
    "os"
    "os/signal"
    "sonarfirmware/api"
    "sonarfirmware/bindata"
    "sonarfirmware/config"
    "sonarfirmware/shells"
    "sonarfirmware/structs"
    sonarmap "sonarmap/config"
    "time"
    "path"
)

const (
    version         = 2
    srvAddr         = "239.2.1.1:2052"
    maxDatagramSize = 8192
    timeoutReadSettings = 15 * time.Second
    timeoutSuccessFlash = 30 * time.Second
    timeoutFailFlash = 5 * time.Second
)

func watchSettings(ip string, settingsC chan structs.Settings) {
    addr, err := net.ResolveUDPAddr("udp", ip)
    if err != nil {
        log.Fatal(err)
    }

    listener, err := net.ListenMulticastUDP("udp", nil, addr)
    if err != nil {
        log.Fatal(err)
    }

    log.Println(addr)

    listener.SetReadBuffer(maxDatagramSize)

    for {
        buffer := make([]byte, maxDatagramSize)
        _, _, err := listener.ReadFromUDP(buffer)
        if err != nil {
            log.Fatal("ReadFromUDP failed:", err)
        }

        dec := json.NewDecoder(bytes.NewReader(buffer))
        var settings = structs.Settings{}
        dec.Decode(&settings)
        settingsC <- settings
    }
}

func tryFlash(settings structs.Settings) bool {
    var (
        err error
        sonarMap []byte
    )

    log.Println("Start flashing...")

    ssh := shells.NewTelnetShell(settings.IP, config.Username, config.Password)
    a := api.New(ssh)

    if a.GetVersion() >= version {
        log.Println("Alread flashed")
        return true
    }

    sonarMap, err = bindata.Asset("sonarmap")
    if err != nil {
        log.Println("Error: ", err)
        return false
    }

    if err = a.UploadSonarMap(sonarMap); err != nil {
        log.Println("UploadSonarMap error:", err)
        return false
    }

    if err = a.PatchRc(); err != nil {
        log.Println("PatchRc error:", err)
        return false
    }

    if err = a.SetVersion(version); err != nil {
        log.Println("SetVersion error:", err)
        return false
    }

    if err = a.ChangePassword("c7195563b28d9ffff104342dcb5d4cb7"); err != nil {
        log.Println("ChangePassword error:", err)
        return false
    }

    return true
}

func WriteZeroconfig(ip string)  {
    var (
        fileName string
        err error
        fp *os.File
    )

    _,err = os.Stat(sonarmap.Current.DirZeroConfig)
    if os.IsNotExist(err) {
        os.MkdirAll(sonarmap.Current.DirZeroConfig, 0755)
    }

    fileName = path.Join(sonarmap.Current.DirZeroConfig, ip)

    _,err = os.Stat(fileName)
    if os.IsNotExist(err) {
        fp, err = os.OpenFile(fileName, os.O_CREATE | os.O_WRONLY, 0644)
    } else {
        fp, err = os.OpenFile(fileName, os.O_TRUNC | os.O_WRONLY, 0644)
    }

    if err != nil {
        log.Println("WriteZeroconfig error", err)
        return
    }
    defer fp.Close()

    fp.WriteString(time.Now().UTC().String())
}

func main() {
    var settings structs.Settings
    sigC := make(chan os.Signal, 1)
    settingsC := make(chan structs.Settings)
    flashTimer := time.NewTimer(timeoutReadSettings)

    signal.Notify(sigC, os.Interrupt)

    go watchSettings(srvAddr, settingsC)

    reset := func(d time.Duration) {
        log.Println("Reset timer:", d)
        if !flashTimer.Stop() {
            select {
            case <- flashTimer.C:
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
            WriteZeroconfig(settings.IP_Zeroconfig)
            // reset(timeoutReadSettings)
        case <-flashTimer.C:
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
