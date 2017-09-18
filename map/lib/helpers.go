package lib

import (
    "encoding/hex"
    "log"
    "os"
)

var logger = log.New(os.Stdout, "SonarMap [HexToBytes]: ", log.LstdFlags | log.LUTC)


func HexToBytes(val string) (result [32]byte) {
    buff, err := hex.DecodeString(val)
    if err != nil {
        logger.Panic("Can't decode string: ", val)
    }

    copy(result[:], buff)
    //logger.Println(result)
    return
}

