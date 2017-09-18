package nmea_handler

import (
    "log"
    "net"
    "os"
    "time"

    "github.com/ilya-pirogov/sonarmap/map/config"
)

var logger = log.New(os.Stdout, "SonarMap [SdCard]: ", log.LstdFlags|log.LUTC)

func StartNmeaHandling() {
    var conn *net.TCPConn

    for {
        conn.Close()

        log.Println("Waiting 30 second before start NMEA handling")
        time.Sleep(30 * time.Second)

        serverAddr := "127.0.0.1:" + config.Current.DataPort
        tcpAddr, err := net.ResolveTCPAddr("tcp", serverAddr)
        if err != nil {
            logger.Println("ResolveTCPAddr failed:", err.Error())
            continue
        }

        conn, err = net.DialTCP("tcp", nil, tcpAddr)
        if err != nil {
            println("Dial failed:", err.Error())
            os.Exit(1)
        }
        reply := make([]byte, 1024)

        _, err = conn.Read(reply)
        if err != nil {
            logger.Println("Write to server failed:", err.Error())
            continue
        }

        logger.Println("reply from server=", string(reply))
    }
}