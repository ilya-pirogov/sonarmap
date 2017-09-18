package rpc

import (
    "bytes"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "net/rpc"
    "os"
    "os/exec"
    "strconv"
    "strings"

    "github.com/ilya-pirogov/sonarmap/firmware/config"
    "github.com/ilya-pirogov/sonarmap/map/sdcard"
)

var logger = log.New(os.Stdout, "SonarMap [RPC]: ", log.LstdFlags | log.LUTC)

type SonarRpc struct {
    currentSd *sdcard.SdCard
}

// Method: Exec
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

// Method: GetVersion
type GetVersionArgs struct {
}

type GetVersionReply struct {
    Version int
    IsValidSd bool
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
    reply.IsValidSd = t.currentSd.IsValid()
    return nil
}

// Method: StartUpload
type StartUploadArgs struct {
    FullPath string
    Force bool
}

type StartUploadReply struct {
    RequestId string
}

func (t *SonarRpc) StartUpload(args *StartUploadArgs, reply *StartUploadReply) error {
    return nil
}

func Start(currentSd *sdcard.SdCard) {
    sonar := SonarRpc{
        currentSd: currentSd,
    }

    rpc.Register(&sonar)
    rpc.HandleHTTP()
    listener, err := net.Listen("tcp", ":7654")
    if err != nil {
        log.Fatal("listen error:", err)
    }
    go http.Serve(listener, nil)
    logger.Println("RPC started at port: 7654")
}
