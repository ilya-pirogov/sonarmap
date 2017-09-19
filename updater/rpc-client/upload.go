package rpc_client

import (
	"crypto/md5"
	"io"
	"log"
	"net/rpc"

	"github.com/ilya-pirogov/sonarmap/updater/structs"
	"github.com/pkg/errors"
)

type Client struct {
	rpcClient *rpc.Client
}

func NewClient(ip string) (*Client, error) {
	client := &Client{}
	rpcClient, err := rpc.DialHTTP("tcp", ip+structs.PORT)
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect to "+ip+structs.PORT)
	}

	client.rpcClient = rpcClient
	return client, nil
}

func (c *Client) Close() {
	c.rpcClient.Close()
}

func (c *Client) Upload(reader io.ReadSeeker, remotePath string, size int64, replace bool) error {
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "unable to seek the beginning of file")
	}

	h := md5.New()
	_, err = io.Copy(h, reader)
	if err != nil {
		return errors.Wrap(err, "unable to calculate MD5 sum")
	}
	var md5sum []byte
	md5sum = h.Sum(nil)

	var replyStart structs.UploadStartReply

	err = c.rpcClient.Call("Upload.Start", &structs.UploadStartArgs{
		Size:       size,
		RemotePath: remotePath,
		Md5:        md5sum,
		Replace:    replace,
	}, &replyStart)

	if err != nil {
		return errors.Wrap(err, "rpc error from Upload.Start")
	}

	fid := replyStart.Id

	reader.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, "unable to seek the beginning of file")
	}

	var left int64 = size
	var buffer [structs.BLOCK_SIZE]byte
	for left > 0 {
		var replyBlock structs.UploadSendBlockReply

		n, err := reader.Read(buffer[:])
		if err != nil {
			return errors.Wrap(err, "unable to read next block")
		}

		err = c.rpcClient.Call("Upload.SendBlock", &structs.UploadSendBlockArgs{
			Id:   fid,
			Data: buffer[:n],
		}, &replyBlock)
		if err != nil {
			return errors.Wrap(err, "rpc error from Upload.SendBlock")
		}

		left = replyBlock.Left
	}

	var replyFinish structs.UploadFinishReply
	err = c.rpcClient.Call("Upload.Finish", &structs.UploadFinishArgs{
		Id: fid,
	}, &replyFinish)
	if err != nil {
		return errors.Wrap(err, "rpc error from Upload.Finish")
	}

	log.Println(replyFinish.Stats)
	log.Printf("Uploaded for %s", replyFinish.Duration)

	return nil
}
