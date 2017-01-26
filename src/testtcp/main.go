package main

import (
    "net"
    "fmt"
    "time"
)

func main() {
    conn, err := net.Dial("tcp", "127.0.0.1:2345");
    if err != nil {
        panic(err)
    }

    err = conn.SetWriteDeadline(time.Now().Add(time.Minute))

    fmt.Fprintf(conn, "asd123fgh\r\n\r\n")
    //conn.Close()
}