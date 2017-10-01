package shells

import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "path"
    "strings"

    "golang.org/x/crypto/ssh"
)

type SshShell struct {
    Ip string
    Config *ssh.ClientConfig
}

func (shell *SshShell) connect() (client *ssh.Client, session *ssh.Session, err error) {
    if client, err = ssh.Dial("tcp", shell.Ip, shell.Config); err != nil {
        return
    }

    if session, err = client.NewSession(); err != nil {
        return
    }
    return
}

func (shell *SshShell) Run(cmd string) (out string, err error) {
    var (
        outVal []byte
        client *ssh.Client
        session *ssh.Session
        stdout io.Reader
        stderr io.Reader
    )

    if client, session, err = shell.connect(); err != nil {
        return
    }
    defer client.Close()
    defer session.Close()

    if stdout, err = session.StdoutPipe(); err != nil {
        return
    }

    if stderr, err = session.StderrPipe(); err != nil {
        return
    }

    if err = session.Run(cmd); err != nil {
        errVal, _ := ioutil.ReadAll(stderr)
        if errVal == nil {
            return out, errors.New("Unknown error")
        }
        return out, errors.New(strings.TrimSpace(string(errVal)))
    }

    if outVal, err = ioutil.ReadAll(stdout); err != nil {
        return
    }

    out = string(outVal)
    err = nil
    return
}

func (shell *SshShell) CopyString(data string, remotePath string, permissions string) error {
    buff := strings.NewReader(data)
    return shell.CopyFile(buff, remotePath, permissions)
}

func (shell *SshShell) CopyBytes(data []byte, remotePath string, permissions string) error {
    buff := bytes.NewBuffer(data)
    return shell.CopyFile(buff, remotePath, permissions)
}

// Copies the contents of an io.Reader to a remote location
func (shell *SshShell) CopyFile(fileReader io.Reader, remotePath string, permissions string) (err error) {
    var (
        client *ssh.Client
        session *ssh.Session
        contentsBytes []byte
    )

    if client, session, err = shell.connect(); err != nil {
        return
    }
    defer client.Close()
    defer session.Close()

    if contentsBytes, err = ioutil.ReadAll(fileReader); err != nil {
        return
    }

    contents := string(contentsBytes)
    filename := path.Base(remotePath)
    directory := path.Dir(remotePath)

    go func() {
        w, _ := session.StdinPipe()
        defer w.Close()
        fmt.Fprintln(w, "C"+permissions, len(contents), filename)
        fmt.Fprintln(w, contents)
        fmt.Fprintln(w, "\x00")
    }()

    if err = session.Run("/usr/bin/scp -t " + directory); err != nil {
        return
    }

    return nil
}

func NewSshShell(ip, login, password string) SshShell {
    config := &ssh.ClientConfig{
        User: login,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
    }

    return SshShell{
        Ip: ip,
        Config: config,
    }
}