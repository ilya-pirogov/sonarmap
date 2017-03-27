package shells

import (
    "bytes"
    "crypto/md5"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net"
    "strings"
    "time"

    "github.com/ziutek/telnet"
    "sonarfirmware/config"
)

const (
    cmdPrefix = "echo -ne '"
    cmdSuffix = "' >> "
    buffLen = 32
    port = "4879"
    totalAttemts = 10
)

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

type TelnetShell struct {
    Addr string
    User string
    Conn *telnet.Conn
    Timeout time.Duration
    password string
    isConnected bool
    terminator byte
}

func (shell *TelnetShell) expect(d ...string) (err error) {
    if err = shell.Conn.SetReadDeadline(time.Now().Add(shell.Timeout)); err != nil {
        return
    }

    err = shell.Conn.SkipUntil(d...)
    return
}

func (shell *TelnetShell) sendLn(s string) (err error) {
    ts := time.Now().Add(shell.Timeout)
    if err = shell.Conn.SetReadDeadline(ts); err != nil {
        return
    }
    buf := make([]byte, len(s)+1)
    copy(buf, s)
    buf[len(s)] = '\n'
    _, err = shell.Conn.Write(buf)
    return
}

func (shell *TelnetShell) sendBytes(buf []byte) (err error) {
    if err = shell.Conn.SetReadDeadline(time.Now().Add(shell.Timeout)); err != nil {
        return
    }
    _, err = shell.Conn.Write(buf)
    return
}

func (shell *TelnetShell) connect() (err error) {
    log.Println("Connecting to " + shell.Addr)
    if shell.Conn, err = telnet.Dial("tcp", shell.Addr + ":23"); err != nil {
        return
    }
    shell.Conn.SetUnixWriteMode(true)

    if err = shell.expect("login: "); err != nil { return }
    if err = shell.sendLn(shell.User); err != nil { return }
    if err = shell.expect("ssword: "); err != nil { return }
    if err = shell.sendLn(shell.password); err != nil { return }
    if err = shell.expect(string(shell.terminator)); err != nil { return }

    shell.isConnected = true
    return
}

func (shell *TelnetShell) Run(cmd string) (out string, err error) {
    log.Printf("%s %s %s\n", shell.Addr, string(shell.terminator), cmd)
    var (
        data []byte
    )

    if !shell.isConnected {
        if err = shell.connect(); err != nil { return }
    }

    shell.sendLn(cmd)
    if data, err = shell.Conn.ReadBytes(shell.terminator); err != nil { return }
    out = string(data)
    return
}

func (shell *TelnetShell) CopyBytes(data []byte, remotePath string, permissions string) (err error) {
    var (
        total int64
        conn net.Conn
    )

    md5bytes := md5.Sum(data)
    hash := hex.EncodeToString(md5bytes[:])
    fileReader := bytes.NewBuffer(data)

    if !shell.isConnected {
        if err = shell.connect(); err != nil { return }
    }

    log.Printf("Start uploading %s", remotePath)
    attempt := totalAttemts

    for attempt > 0 {
        attempt--
        log.Printf("Attempt #%d", totalAttemts - attempt)

        if _, err = shell.Run(fmt.Sprintf(config.NetCatCmd, port, remotePath)); err != nil {
            log.Println(err)
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        time.Sleep(1 * time.Second)

        if conn, err = net.Dial("tcp", shell.Addr + ":" + port); err != nil {
            log.Println(err)
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        log.Printf("Starting transfer file %s. Hash: %s", remotePath, hash)
        if total, err = io.Copy(conn, fileReader); err != nil {
            log.Println(err)
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        shell.Run("sync")
        time.Sleep(3 * time.Second)

        conn.Close()
        shell.Run("sync")
        shell.Run("true")

        res, err := shell.Run("md5sum " + remotePath)
        lines := strings.Split(res, "\n")

        if err != nil {
            log.Println(err)
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        if len(lines) < 2 || len(lines[1]) < 32 {
            err = errors.New(fmt.Sprintf("Unable to calculate hash. Got: %s", hash))
            log.Println(err)
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        newHash := lines[1][:32]

        if hash != newHash {
            err = errors.New(fmt.Sprintf("Inccorect hash. Got: %s", newHash))
            log.Println(err)
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        _, err = shell.Run(fmt.Sprintf("chmod %s %s", permissions, remotePath))
        if err != nil {
            err = errors.New(fmt.Sprintf("Unable to change permission to %s. Error: %s", permissions, err))
            if attempt == 0 { return err }
            time.Sleep(10 * time.Second)
            continue
        }

        if err == nil {
            break
        }
    }

    if err != nil {
        return err
    }

    log.Printf("Successful copied %d bytes into %s\n", total, remotePath)
    return
}

func NewTelnetShell(ip, login, password string) TelnetShell {
    terminator := byte('$')
    if login == "root" {
        terminator = byte('#')
    }
    return TelnetShell {
        Addr: ip,
        User: login,
        password: password,
        isConnected: false,
        Timeout: 60 * time.Second,
        terminator: terminator,
    }
}
