package main

import (
	"fmt"
	"flag"
	"os"
	"strings"
	"crypto/sha256"
	"encoding/hex"
	"time"
	"sonarmap/config"
	"text/template"
	"bytes"
	"log"
	"os/exec"
	"io/ioutil"

	"github.com/jteeuwen/go-bindata"
	"path/filepath"
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

	FileWatch: "{{.FileWatch}}",
	TimeoutChanges: {{.TimeoutChanges.Seconds}} * time.Second,

	FileLive: "{{.FileLive}}",
	FileIsAlive: "{{.FileIsAlive}}",
	DirLogs: "{{.DirLogs}}",
	TimeoutIsAlive: {{.TimeoutIsAlive.Seconds}} * time.Second,
}
`

var prodDefaults = config.Sd{
	SdPart: 1,
	SdSys: "/sys/block/%s/device/cid",
	SdDev: "/dev/mmcblk[1-9]",
	DirMedia: "/media/%sp%d",

	// flush cache settings
	FileWatch: "/Live/Large.at5",
	TimeoutChanges: 5 * time.Second,

	// alive checker settings
	FileLive: "/LIVE.sl?",
	FileIsAlive: "/media/userdata/.StarMaps/is-alive",
	DirLogs: "/live_logs",
	TimeoutIsAlive: 5 * time.Second,
}

var devDefaults = config.Sd{
	SdPart: 1,
	SCid: "bfd7f660609c814ba4cfb47497476a37f1503e5931c25b42999331cfffe5c2f0",
	SdSys: "/home/ilya/fake-sys/%s/cid",
	SdDev: "/home/ilya/fake-dev/mm?",
	DirMedia: "/home/ilya/fake-media/%sp%d",

	// flush cache settings
	FileWatch: "/Live/Large.at5",
	TimeoutChanges: 5 * time.Second,

	// alive checker settings
	FileLive: "/LIVE.sl?",
	FileIsAlive: "/media/userdata/is-alive",
	DirLogs: "/live_logs",
	TimeoutIsAlive: 5 * time.Second,
}

func main() {
	var (
		pwd     string
		fs      *os.File
		err     error
		cid     string
		stdErr  bytes.Buffer
		current= config.Sd{}
	)

	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
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

	flag.StringVar(&current.FileWatch, "file-watch", prodDefaults.FileWatch, "relative path to Large.at5")
	flag.DurationVar(&current.TimeoutChanges, "timeout-changes", prodDefaults.TimeoutChanges, "timeout between flush caches")

	flag.StringVar(&current.FileLive, "file-live", prodDefaults.FileLive, "relative path to LIVE file")
	flag.StringVar(&current.FileIsAlive, "file-is-alive", prodDefaults.FileIsAlive, "absolute path to is-alive file")
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

	_, err = os.Stat(configFile)
	if os.IsNotExist(err) {
		fs, err = os.OpenFile(configFile, os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		fs, err = os.OpenFile(configFile, os.O_TRUNC|os.O_WRONLY, 0644)
	}

	if cid == "dev" {
		current = devDefaults
	}

	buf := bytes.NewBufferString("")
	configTemplate.Execute(buf, current)
	configTemplate.Execute(fs, current)

	log.Printf("Generating config...\n%s", buf.String())

	tmpDir, err := ioutil.TempDir("", "sonarmap-build")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "sonarmap")

	pwd, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Working dir: %s; GOROOT: %s", pwd, os.Getenv("GOROOT"))

	goLang := exec.Command("go", "build", "-o", tmpFile, "src/sonarmap/main.go")
	goLang.Env = []string{"GOOS=linux", "GOARCH=arm", "GOPATH=" + pwd, "HOME=" + pwd, "GOROOT=" + os.Getenv("GOROOT") }
	goLang.Stderr = &stdErr
	goLang.Dir = pwd

	log.Printf("Building sonarmap to %s", tmpFile)
	err = goLang.Run()
	if err != nil {
		log.Fatalf("Unable to build sonarmap:\n%s", stdErr.String())
	}

	binDataPath := "src/sonarfirmware/bindata/bindata.go"
	_, err = os.Stat(binDataPath)
	if !os.IsNotExist(err) {
		err = os.Remove(binDataPath)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Printf("Generating bindata: %s", binDataPath)
	cfg := bindata.NewConfig()
	cfg.Package = "bindata"
	cfg.Output = binDataPath
	cfg.Prefix = tmpDir
	cfg.Input = []bindata.InputConfig{
		{
			Path:      filepath.Clean(tmpDir),
			Recursive: false,
		},
	}
	err = bindata.Translate(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	distFile := "dist/sonarfirmware.exe"
	goLang = exec.Command("go", "build", "-o", distFile, "src/sonarfirmware/main.go")
	goLang.Env = []string{"GOOS=windows", "GOARCH=386", "GOPATH=" + pwd, "HOME=" + pwd, "GOROOT=" + os.Getenv("GOROOT")}
	goLang.Stderr = &stdErr
	goLang.Dir = pwd

	_, err = os.Stat(distFile)
	if !os.IsNotExist(err) {
		err = os.Remove(distFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	log.Printf("Building sonarfirmware to %s", distFile)
	err = goLang.Run()
	if err != nil {
		log.Fatalf("Unable to build sonarfirmware:\n%s", stdErr.String())
	}
}
