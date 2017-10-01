package main

import (
	"flag"
	"sonarmap/updater/rpc-client"
	"log"
	"os"
)

var localFile string
var remotePath string
var ip string
var replace bool

func init() {
	flag.StringVar(&localFile, "local-file", "", "local file which will be copied")
	flag.StringVar(&remotePath, "remote-path", "", "full remote path where file will be created")
	flag.StringVar(&ip, "ip", "", "IP to connect")
	flag.BoolVar(&replace, "replace", false, "replace destination file?")

	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	flag.Parse()

	if localFile == "" || remotePath == "" || ip == "" {
		flag.Usage()
		return
	}

	client, err := rpc_client.NewClient(ip)
	if err != nil {
		log.Fatal(err)
	}

	stat, err := os.Stat(localFile)
	if err != nil {
		log.Fatal(err)
	}

	fp, err := os.Open(localFile)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Upload(fp, remotePath, stat.Size(), replace)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("DONE!")
}
