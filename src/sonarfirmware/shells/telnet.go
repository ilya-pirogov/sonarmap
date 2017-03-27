package shells

import (
    "bytes"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net"
    "strconv"
    "strings"
    "time"

    "github.com/ziutek/telnet"
    "sonarfirmware/config"
)

const (
    cmdPrefix = "echo -ne '"
    cmdSuffix = "' >> "
    buffLen = 32
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

func (shell *TelnetShell) CopyString(data string, remotePath string, permissions string) error {
    buff := strings.NewReader(data)
    return shell.CopyFile(buff, remotePath, permissions)
}

func (shell *TelnetShell) CopyBytes(data []byte, remotePath string, permissions string) error {
    buff := bytes.NewBuffer(data)
    return shell.CopyFile(buff, remotePath, permissions)
}

// Copies the contents of an io.Reader to a remote location
func (shell *TelnetShell) CopyFile(fileReader io.Reader, remotePath string, permissions string) (err error) {
    var (
        total int64
        conn net.Conn
    )

    d := md5.New()
    io.Copy(d, fileReader)
    hash := hex.EncodeToString(d.Sum(nil))

    if !shell.isConnected {
        if err = shell.connect(); err != nil { return }
    }

    log.Printf("Start uploading %s", remotePath)
    attempt := totalAttemts

    for attempt > 0 {
        attempt--
        log.Printf("Attempt #%d", totalAttemts - attempt)

        port := strconv.Itoa(rand.Intn(1024) + 4096)
        if _, err = shell.Run(fmt.Sprintf(config.NetCatCmd, port, remotePath)); err != nil {
            log.Println(err)
            continue
        }
        if conn, err = net.Dial("tcp", shell.Addr + ":" + port); err != nil {
            log.Println(err)
            continue
        }

        log.Printf("Starting transfer file %s. Hash: %s", remotePath, hash)
        if total, err = io.Copy(conn, fileReader); err != nil {
            log.Println(err)
            continue
        }

        time.Sleep(500 * time.Millisecond)
        conn.Close()
        shell.Run("sync")
        shell.Run("true")

        res, err := shell.Run("md5sum " + remotePath)
        if err != nil {
            log.Println(err)
            continue
        }

        if len(res) < 32 {
            log.Printf("Unable to calculate hash. Got: %s", hash)
            continue
        }

        if hash != res[:32] {
            log.Printf("Inccorect hash. Got: %s", res[:32])
            continue
        }

        _, err = shell.Run(fmt.Sprintf("chmod %s %s", permissions, remotePath))
        if err != nil {
            log.Printf("Unable to change permission to %s. Error: %s", permissions, err)
        }
    }

    if err != nil {
        return
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
