package shells

import (
    "bytes"
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

    if !shell.isConnected {
        if err = shell.connect(); err != nil { return }
    }

    port := strconv.Itoa(rand.Intn(1024) + 4096)
    if _, err = shell.Run(fmt.Sprintf(config.NetCatCmd, port, remotePath)); err != nil { return }
    if conn, err = net.Dial("tcp", shell.Addr + ":" + port); err != nil { return };

    log.Println("Starting transfer file...")
    if total, err = io.Copy(conn, fileReader); err != nil {
        panic(err)
    }

    time.Sleep(1 * time.Second)

    conn.Close()
    shell.Run("true")

    log.Printf("Copied %d bytes into %s\n", total, remotePath)
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