package rpc_server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"github.com/ilya-pirogov/sonarmap/updater/structs"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandId() structs.FileID {
	var id structs.FileID
	rand.Read(id[:])
	return id
}

type file struct {
	id        structs.FileID
	fp        *os.File
	error     error
	totalSize int64
	uploaded  int64
	md5s      []byte
	path      string
	done      bool
	started   time.Time
}

type Upload struct {
	files map[structs.FileID]*file
}

func (u *Upload) Start(args *structs.UploadStartArgs, reply *structs.UploadStartReply) error {
	//if u.files[] != "" {
	//    return errors.New("already uploading another file")
	//}

	_, err := os.Stat(args.RemotePath)
	if !os.IsNotExist(err) {
		if args.Replace {
			err = os.Remove(args.RemotePath)
			if err != nil {
				return errors.Wrap(err, "unable to remove file: "+args.RemotePath)
			}
		} else {
			return errors.New("file already exists: " + args.RemotePath)
		}
	}

	id := RandId()
	fileMeta := file{}
	fileMeta.id = id
	fileMeta.path = args.RemotePath

	fp, err := os.Create(args.RemotePath)
	if err != nil {
		return errors.Wrap(err, "unable to crete file: "+args.RemotePath)
	}

	fileMeta.fp = fp
	fileMeta.md5s = args.Md5
	fileMeta.totalSize = args.Size
	fileMeta.uploaded = 0
	fileMeta.started = time.Now()
	reply.Id = fileMeta.id
	u.files[fileMeta.id] = &fileMeta

	go func() {
		timer := time.NewTimer(1 * time.Minute)
		<-timer.C

		delete(u.files, id)
	}()

	return nil
}

func (u *Upload) SendBlock(args *structs.UploadSendBlockArgs, reply *structs.UploadSendBlockReply) error {
	fileMeta := u.files[args.Id]
	if fileMeta == nil {
		return errors.New("not found or time outed")
	}

	if len(args.Data) == 0 {
		return errors.New("no data")
	}

	if int64(len(args.Data))+fileMeta.uploaded > fileMeta.totalSize {
		return errors.New("too big block")
	}

	n, err := fileMeta.fp.Write(args.Data)
	fileMeta.uploaded += int64(n)
	if err != nil {
		return errors.Wrap(err, "unable to write block of "+strconv.Itoa(len(args.Data))+" bytes")
	}

	reply.Uploaded = int64(n)
	reply.Left = int64(fileMeta.totalSize - fileMeta.uploaded)

	return nil
}

func (u *Upload) Finish(args *structs.UploadFinishArgs, reply *structs.UploadFinishReply) error {
	fileMeta := u.files[args.Id]
	if fileMeta == nil {
		return errors.New("not found or time outed")
	}

	left := fileMeta.totalSize - fileMeta.uploaded
	if left != 0 {
		return errors.New("file has not been completely uploaded, bytes left: " + strconv.Itoa(int(left)))
	}

	err := fileMeta.fp.Sync()
	if err != nil {
		return errors.Wrap(err, "unable to sync file pointer")
	}

	_, err = fileMeta.fp.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "unable to seek to the beginning of file")
	}

	h := md5.New()
	_, err = io.Copy(h, fileMeta.fp)
	if err != nil {
		return errors.Wrap(err, "unable to calculate MD5")
	}

	md5sum := h.Sum(nil)
	if bytes.Compare(md5sum, fileMeta.md5s[:]) != 0 {
		msg := fmt.Sprintf("md5 doesn't match! expect: %x; got: %x", fileMeta.md5s[:], md5sum)
		return errors.New(msg)
	}

	fileMeta.fp.Close()
	delete(u.files, fileMeta.id)

	stat, _ := os.Stat(fileMeta.path)

	reply.Stats = &structs.FileInfo{
		Size:     stat.Size(),
		Name:     stat.Name(),
		Mode:     stat.Mode(),
		ModeTime: stat.ModTime(),
	}

	reply.Duration = time.Since(fileMeta.started)

	return nil
}

func StartServer() {
	upload := &Upload{make(map[structs.FileID]*file)}
	rpc.Register(upload)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", structs.PORT)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(listener, nil)
}
