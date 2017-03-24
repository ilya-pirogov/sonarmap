package main

import (
    "bytes"
    "crypto/sha256"
    "encoding/hex"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "strconv"
    "strings"
    "text/template"
    "time"

    "github.com/jteeuwen/go-bindata"
    "liba"
    "sonarmap/config"
)

const configFile = "src/sonarmap/config/current.go"

const structTpl = `// DO NOT EDIT. Generated dynamically.
package config

import "time"

var Current = Sd{
	SCid: "{{.SCid}}",
	SdPart: {{.SdPart}},
	SdSys: "{{.SdSys}}",
	SdDev: "{{.SdDev}}",
	DirMedia: "{{.DirMedia}}",
	DirZeroConfig: "{{.DirZeroConfig}}",
	Build: {{.Build}},

	FileWatch: "{{.FileWatch}}",
	TimeoutChanges: {{.TimeoutChanges.Seconds}} * time.Second,

	FileLive: "{{.FileLive}}",
	FileIsAlive: "{{.FileIsAlive}}",
	FileWallpaper: "{{.FileWallpaper}}",
	DirLogs: "{{.DirLogs}}",
	TimeoutIsAlive: {{.TimeoutIsAlive.Seconds}} * time.Second,
}
`

var prodDefaults = config.Sd{
    SdPart:        1,
    SdSys:         "/sys/block/%s/device/cid",
    SdDev:         "/dev/mmcblk[1-9]",
    DirMedia:      "/media/%sp%d",
    DirZeroConfig: ".",

    // flush cache settings
    FileWatch:      "/Live/Large.at5",
    TimeoutChanges: 5 * time.Second,

    // alive checker settings
    FileLive:       "/LIVE.sl?",
    FileIsAlive:    "/media/userdata/.StarMaps/is-alive",
    FileWallpaper:  "/media/userdata/wallpaper/wallpaper01.jpg",
    DirLogs:        "/live_logs",
    TimeoutIsAlive: 5 * time.Second,
}

var devDefaults = config.Sd{
    SdPart:        1,
    SCid:          "bfd7f660609c814ba4cfb47497476a37f1503e5931c25b42999331cfffe5c2f0",
    SdSys:         "/home/ilya/fake-sys/%s/cid",
    SdDev:         "/home/ilya/fake-dev/mm?",
    DirMedia:      "/home/ilya/fake-media/%sp%d",
    DirZeroConfig: "zeroconfig.txt",

    // flush cache settings
    FileWatch:      "/Live/Large.at5",
    TimeoutChanges: 5 * time.Second,

    // alive checker settings
    FileLive:       "/LIVE.sl?",
    FileIsAlive:    "/media/userdata/is-alive",
    FileWallpaper:  "/media/userdata/wallpaper/wallpaper01.jpg",
    DirLogs:        "/live_logs",
    TimeoutIsAlive: 5 * time.Second,
}

var logger = log.New(os.Stdout, "SonarFirmware: ", log.Ltime)
var fs = liba.NewFs(logger)

func getGoEnvs(archEnv, osEnv string) []string {
    pwd, err := os.Getwd()
    if err != nil {
        logger.Fatalln(err)
    }

    return []string{
        "GOOS=" + osEnv,
        "GOARCH=" + archEnv,
        "GOPATH=" + pwd,
        "HOME=" + pwd,
        "GOROOT=" + os.Getenv("GOROOT"),
        "TEMP=" + os.Getenv("TEMP"),
        "TMPDIR=" + os.Getenv("TMPDIR"),
        "TMP=" + os.Getenv("TMP"),
    }
}

func main() {
    var (
        pwd     string
        fp      *os.File
        err     error
        cid     string
        stdErr  bytes.Buffer
        current = config.Sd{}
    )

    var Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s -cid CID [options]\n", os.Args[0])
        flag.PrintDefaults()
    }

    configTemplate, err := template.New("sd").Parse(structTpl)
    if err != nil {
        panic(err)
    }

    flag.StringVar(&cid, "cid", "", "CID value of SD Card")
    flag.IntVar(&current.SdPart, "sd-part", prodDefaults.SdPart, "partiotion number of SD Card")
    flag.StringVar(&current.SdSys, "sd-sys", prodDefaults.SdSys, "path to cid file of /sys")
    flag.StringVar(&current.SdDev, "sd-dev", prodDefaults.SdDev, "path to device block of /dev")
    flag.StringVar(&current.DirMedia, "dir-media", prodDefaults.DirMedia, "path to mount point")
    flag.StringVar(&current.DirZeroConfig, "dir-zero-config", prodDefaults.DirZeroConfig, "direcory to putting zeroconf files")
    flag.Int64Var(&current.Build, "build", -1, "number of build. autoincrements if equal -1")

    flag.StringVar(&current.FileWatch, "file-watch", prodDefaults.FileWatch, "relative path to Large.at5")
    flag.DurationVar(&current.TimeoutChanges, "timeout-changes", prodDefaults.TimeoutChanges, "timeout between flush caches")

    flag.StringVar(&current.FileLive, "file-live", prodDefaults.FileLive, "relative path to LIVE file")
    flag.StringVar(&current.FileIsAlive, "file-is-alive", prodDefaults.FileIsAlive, "absolute path to is-alive file")
    flag.StringVar(&current.FileWallpaper, "file-wallpaper", prodDefaults.FileWallpaper, "absolute path to wallpaper")
    flag.StringVar(&current.DirLogs, "dir-logs", prodDefaults.DirLogs, "relative path to logs")
    flag.DurationVar(&current.TimeoutIsAlive, "timeout-is-alive", prodDefaults.TimeoutIsAlive, "timout for ia-alive check")

    flag.Parse()

    if cid == "" {
        Usage()
        return
    }

    buffer := []byte(strings.TrimSpace(fmt.Sprintf("SM#CID:%s", cid)))
    scid := sha256.Sum256(buffer[:])
    current.SCid = hex.EncodeToString(scid[:])
    if current.Build == -1 {
        var (
            buff = make([]byte, 16)
            fp2  *os.File
            num  int
        )

        fp2, err = os.OpenFile("build.txt", os.O_RDWR|os.O_CREATE, 0644)
        if err != nil {
            logger.Panicln(err)
        }
        defer fp2.Close()
        num, err = fp2.Read(buff)
        println(num)

        if num == 0 {
            current.Build = 1
        } else {
            current.Build, err = strconv.ParseInt(string(buff[0:num]), 10, 64)
            if err != nil {
                logger.Panicln(err)
            }
            current.Build++
        }

        _, err = fp2.Seek(0, 0)
        if err != nil {
            logger.Panicln(err)
        }
        fp2.WriteString(strconv.FormatInt(current.Build, 10))
    }

    _, err = os.Stat(configFile)
    if os.IsNotExist(err) {
        fp, err = os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY, 0644)
    } else {
        fp, err = os.OpenFile(configFile, os.O_TRUNC|os.O_WRONLY, 0644)
    }

    if cid == "dev" {
        current = devDefaults
    }

    buf := bytes.NewBufferString("")
    configTemplate.Execute(buf, current)
    configTemplate.Execute(fp, current)

    logger.Printf("Generating config...\n%s", buf.String())

    tmpDir, err := ioutil.TempDir("", "sonarmap-build")
    if err != nil {
        logger.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    tmpFile := filepath.Join(tmpDir, "sonarmap")

    tmpWallpaper := filepath.Join(tmpDir, "wallpaper.jpg")
    fileWallpaper := filepath.Join("data", "wallpaper.jpg")
    logger.Println("Copy assets")
    fs.CopyFile(fileWallpaper, tmpWallpaper, "Unable to copy wallpaper: %s")
    fs.CopyDir(filepath.Join("data", "translations"), filepath.Join(tmpDir, "translations"), "Unable to copy translations: %s")

    fileSonarmapGo := filepath.Join("src", "sonarmap", "main.go")
    goLang := exec.Command("go", "build", "-o", tmpFile, fileSonarmapGo)
    goLang.Env = getGoEnvs("arm", "linux")
    goLang.Stderr = &stdErr
    goLang.Dir = pwd

    logger.Printf("Building sonarmap to %s", tmpFile)
    err = goLang.Run()
    if err != nil {
        logger.Fatalf("Unable to build sonarmap:\n%s", stdErr.String())
    }

    binDataPath := filepath.Join("src", "sonarfirmware", "bindata", "bindata.go")
    _, err = os.Stat(binDataPath)
    if !os.IsNotExist(err) {
        err = os.Remove(binDataPath)
        if err != nil {
            logger.Fatalln(err)
        }
    }

    logger.Printf("Generating bindata: %s", binDataPath)
    cfg := bindata.NewConfig()
    cfg.Package = "bindata"
    cfg.Output = binDataPath
    cfg.Prefix = tmpDir
    cfg.Input = []bindata.InputConfig{
        {
            Path:      filepath.Clean(tmpDir),
            Recursive: true,
        },
    }
    err = bindata.Translate(cfg)
    if err != nil {
        logger.Fatalln(err)
    }

    fileSonarfirmwareGo := filepath.Join("src", "sonarfirmware", "main.go")
    distFile := filepath.Join("dist", "sonarfirmware.exe")
    goLang = exec.Command("go", "build", "-o", distFile, fileSonarfirmwareGo)
    goLang.Env = getGoEnvs(runtime.GOARCH, runtime.GOOS)
    goLang.Stderr = &stdErr
    goLang.Dir = pwd

    _, err = os.Stat(distFile)
    if !os.IsNotExist(err) {
        err = os.Remove(distFile)
        if err != nil {
            logger.Fatalln(err)
        }
    }

    logger.Printf("Building sonarfirmware to %s", distFile)
    err = goLang.Run()
    if err != nil {
        logger.Fatalf("Unable to build sonarfirmware:\n%s", stdErr.String())
    }
}
