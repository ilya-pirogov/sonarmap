package structs

import (
	"os"
	"time"
)

type FileID [16]byte

type FileInfo struct {
	Name     string
	Size     int64
	Mode     os.FileMode
	ModeTime time.Time
}

type UploadStartArgs struct {
	RemotePath string
	Replace    bool
	Size       int64
	Md5        []byte
}

type UploadStartReply struct {
	Id FileID
}

type UploadSendBlockArgs struct {
	Id   FileID
	Data []byte
}

type UploadSendBlockReply struct {
	Uploaded int64
	Left     int64
}

type UploadFinishArgs struct {
	Id FileID
}

type UploadFinishReply struct {
	Stats    *FileInfo
	Duration time.Duration
}
