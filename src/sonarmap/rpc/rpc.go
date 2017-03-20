package rpc

import (
    "net/rpc"
    "net"
    "log"
    "net/http"
    "os/exec"
    "bytes"
    "sonarfirmware/config"
    "io/ioutil"
    "strconv"
    "os"
    "strings"
)

var logger = log.New(os.Stdout, "SonarMap [RPC]: ", log.LstdFlags | log.LUTC)

type SonarRpc struct {
}

type ExecArgs struct {
    Cmd string
    Args []string
}

type ExecReply struct {
    Stdout string
    Stderr string
}

func (t *SonarRpc) Exec(args *ExecArgs, reply *ExecReply) error {
    var (
        stdOut bytes.Buffer
        stdErr bytes.Buffer
    )
    cmd := exec.Command(args.Cmd, args.Args...)
    cmd.Stdout = &stdOut
    cmd.Stderr = &stdErr
    err := cmd.Run()
    if err != nil {
        return err
    }

    reply.Stdout = stdOut.String()
    reply.Stderr = stdErr.String()
    return nil
}

type GetVersionArgs struct {
}

type GetVersionReply struct {
    Version int
}

func (t *SonarRpc) GetVersion(args *GetVersionArgs, reply *GetVersionReply) error {
    buf, err := ioutil.ReadFile(config.VerFile)
    if err != nil {
        return err
    }

    ver, err := strconv.Atoi(strings.TrimSpace(string(buf)))
    if err != nil {
        return err
    }

    reply.Version = ver
    return nil
}

func Start() {
    sonar := SonarRpc{}
    rpc.Register(&sonar)
    rpc.HandleHTTP()
    listener, err := net.Listen("tcp", ":7654")
    if err != nil {
        log.Fatal("listen error:", err)
    }
    go http.Serve(listener, nil)
    logger.Println("RPC started at port: 7654")
}
