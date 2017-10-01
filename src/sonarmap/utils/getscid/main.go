package main

import (
    "bufio"
    "crypto/sha256"
    "fmt"
    "os"
    "strings"
)

//func calculateScid(dev string) (scid [32]byte, err error) {
//    //fp, err := os.Open(fmt.Sprintf("/home/ilya/fake-sys/%s/cid", dev))
//    fp, err := os.Open(fmt.Sprintf("/sys/block/%s/device/cid", dev))
//    if (err != nil) {
//        return
//    }
//
//    cid, err := bufio.NewReader(fp).ReadString('\n')
//    if (err != nil) {
//        return
//    }
//
//    buffer := []byte(strings.TrimSpace(fmt.Sprintf("SM#CID:%s", cid)))
//    scid = sha256.Sum256(buffer)
//    return
//}

func main() {
    fmt.Print("Enter CID: ")
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()

    buffer := []byte(strings.TrimSpace(fmt.Sprintf("SM#CID:%s", scanner.Text())))
    scid := sha256.Sum256(buffer)
    fmt.Printf("Secure CID: %x\n", scid)
    fmt.Println("Press enter...")
    scanner.Scan()

    //
    ////matches, err := filepath.Glob("/home/ilya/fake-dev/mm?")
    //matches, err := filepath.Glob("/dev/mmcblk[1-9]")
    //if err != nil {
    //    fmt.Println(err)
    //}
    //
    //for _, match := range matches {
    //    dev := filepath.Base(match)
    //    scid, err := calculateScid(dev)
    //    if err != nil {
    //        fmt.Println(err)
    //    }
    //    fmt.Printf("%s: %x\n", dev, scid)
    //}
}