package main

import (
    "encoding/json"
    "log"
    "net"
    "os"
    "os/signal"
    "time"

    "sonarfirmware/structs"
)

var jsonMock = structs.Settings{
    AppVersion: "53.1.110",
    Barcode: "106329065\n",
    Brand: "Lowrance",
    ContentID: "AC4057044",
    IP: "127.0.0.1",
    IP_Zeroconfig: "169.254.233.2",
    Language: "ru",
    LanguagePack: "Russian Ukrainian",
    Model: "HDS3-9",
    Name: "HDS3-9",
    PlatformType: "5_mx6_solo_800_1.1",
    PlatformVersion: "17.0.32906-r1",
    SerialNumber: "3257312656",
    Services: []structs.Service {
        {
            Port: 80,
            Service: "http",
            Version: 1,
        },
        {
            Port: 21,
            Service: "ftp",
            Version: 1,
        },
        {
            Port: 554,
            Service: "rtsp",
            Version: 1,
        },
        {
            Port: 10110,
            Service: "nmea-0183",
            Version: 1,
        },
        {
            Port: 6633,
            Service: "navico-mfd-rp",
            Version: 1,
        },
        {
            Port: 2053,
            Service: "navico-nav-ws",
            Version: 1,
        },
    },
}

const (
    srvAddr = "239.2.1.1:2052"
)

func main() {
    sigC := make(chan os.Signal, 1)
    signal.Notify(sigC, os.Interrupt)

    go ping(srvAddr)

    for {
        select {
        case <- sigC:
            println("\rInterrupted. Exiting...")
            return
        }
    }
}

func ping(a string) {
    addr, err := net.ResolveUDPAddr("udp", a)
    if err != nil {
        log.Fatal(err)
    }
    c, err := net.DialUDP("udp", nil, addr)

    buff, err := json.Marshal(jsonMock)
    if err != nil {
        log.Fatal(err)
    }

    for {
        println("ping", buff)
        c.Write(buff)
        time.Sleep(1 * time.Second)
    }
}
