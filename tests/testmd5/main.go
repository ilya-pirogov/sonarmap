package main

import (
    "crypto/md5"
    "encoding/hex"
    "flag"
    "io"
    "os"
)

func main() {
    var (
        file string
    )

    flag.StringVar(&file, "file", "", "full path to file")
    flag.Parse()

    fp, err := os.Open(file)
    if err != nil {
        panic(err)
    }

    d := md5.New()
    io.Copy(d, fp)
    hash := d.Sum(nil)
    println(hex.EncodeToString(hash))
}