package downloadfile

import (
	"bytes"
	"encoding/gob"
	"time"
)

type DownLoadFile struct {
	Id         string
	WalletAddr string
	FileId     string
	FileName   string
	FileSize   string
	FileType   string
	NodeList   []string
	SliceSize  int64
	SliceArr   []int64 //文件已经下载后的下标合集
	Payment    int64
	Status     int64 //0 下载中 1下载完成 2 已删除
	CreateTime time.Time
}

func (downloadFile *DownLoadFile) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(downloadFile)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializeDownLoadFile(blockBytes []byte) *DownLoadFile {
	var downloadFile DownLoadFile
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&downloadFile)
	if err != nil {
		panic(err)
	}
	return &downloadFile
}
