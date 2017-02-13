package rpc

import (
    "net/rpc"
    "net"
    "log"
    "net/http"
    "os/exec"
    "bytes"
)

type SonarRpc struct {
}

type Sign struct {
    Sign
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

func Start() {
    sonar := SonarRpc{}
    rpc.Register(&sonar)
    rpc.HandleHTTP()
    l, e := net.Listen("tcp", ":7654")
    if e != nil {
        log.Fatal("listen error:", e)
    }
    go http.Serve(l, nil)
    log.Println("RPC started at port: 7654")
}
