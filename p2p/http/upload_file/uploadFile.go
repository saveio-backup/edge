package uploadfile

import (
	"bytes"
	"encoding/gob"
	"time"
)

type UploadFile struct {
	WalletAddr string
	FileId     string
	CreateTime time.Time
	SliceArr   []int64 //文件已经上传后的下标合集
	// Id         string //本次上传id
	// FileName   string
	// FileSize   string
	// FileType   string
	// NodeList   []string
	// SliceSize  int64

	// Status     int64 // 0 上传中 1 上传完成 2 已经删除
	// Payment    int64

}

func (uploadFile *UploadFile) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(uploadFile)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

//反序列化
func DeserializeUploadFile(blockBytes []byte) *UploadFile {
	var uploadFile UploadFile
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&uploadFile)
	if err != nil {
		panic(err)
	}
	return &uploadFile
}
