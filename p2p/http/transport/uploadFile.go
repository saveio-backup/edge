package transport

import (
	"bytes"
	"encoding/gob"
	"sort"
	"time"

	"github.com/saveio/themis/common/log"
)

type UploadFile struct {
	PeerAddr    string
	FileHash    string
	Hashs       []string //file block hash
	Prefix      string
	NodeListMap map[string][]string
	Indexs      []uint64
	CreatTime   time.Time
	Status      int64 // 0 uploading 1 uploaded 2 delete 3 corrupt
	// Payment    int64
	IsSort bool
}
type UploadFileDetail struct {
	PeerAddr string
	FileHash string
	Hash     string //file block hash
	NodeList []string
	Index    uint64
}

func CreateUploadFile(peerAddr string, fileHash string, prefix string) error {
	uploadFile := UploadFile{
		PeerAddr:    peerAddr,
		FileHash:    fileHash,
		Status:      0,
		Prefix:      prefix,
		CreatTime:   time.Now(),
		NodeListMap: make(map[string][]string),
		IsSort:      false,
	}
	fileByte := uploadFile.Serialize()
	return LevelDBStore.Put([]byte(peerAddr+"_up_doing_"+fileHash), fileByte)

}

func GetUploadFile(peerAddr string, fileHash string) (*UploadFile, error) {
	fileByte, err := LevelDBStore.Get([]byte(peerAddr + "_up_doing_" + fileHash))
	if err != nil {

		return nil, err

	}
	file := DeserializeUploadFile(fileByte)
	return file, nil
}

func FilterUploadHashs(peerAddr string, fileHash string, hashs []string) ([]int, []int) {
	uploadFile, err := GetUploadFile(peerAddr, fileHash)
	if err != nil {
		log.Errorf("peerAddr:%s fileHash:%s get upload file err:%s", peerAddr, fileHash, err)
	}
	var containIndexs []int
	var notContainIndexs []int
	for i := 0; i < len(hashs); i++ {
		b := false
		for j := 0; j < len(uploadFile.Hashs); j++ {
			if hashs[i] == uploadFile.Hashs[j] {
				b = true
				containIndexs = append(containIndexs, i)
			}
		}
		if !b {
			notContainIndexs = append(notContainIndexs, i)
		}
	}
	return notContainIndexs, containIndexs

}

func (uploadFile *UploadFile) UploadPutBlocks(blocks []*UploadFileDetail) error {
	for _, block := range blocks {
		uploadFile.Indexs = append(uploadFile.Indexs, block.Index)
		uploadFile.Hashs = append(uploadFile.Hashs, block.Hash)
		uploadFile.NodeListMap[block.Hash] = block.NodeList
	}
	fileByte := uploadFile.Serialize()
	// err := task.DB.Put([]byte(uploadFile.PeerAddr+"_up_doing_"+uploadFile.TxHash), fileByte, nil)
	err := LevelDBStore.Put([]byte(uploadFile.PeerAddr+"_up_doing_"+uploadFile.FileHash), fileByte)
	if err != nil {
		return err
	}
	log.Debug("fileUplaod success")
	return nil
}
func (uploadFile *UploadFile) PutBlock(block *UploadFileDetail) error {
	return uploadFile.UploadPutBlocks([]*UploadFileDetail{block})
}
func (uploadFile *UploadFile) HashsSort() bool {
	hashs := uploadFile.Hashs
	indexs := uploadFile.Indexs
	var newHashs []string
	var newIndexs []uint64
	if len(hashs) != len(indexs) {
		log.Error("hashs length not euqal index length")
		return false
	}
	m := map[int]string{}
	var keys []int
	for i, hash := range hashs {
		m[int(indexs[i])] = hash
		keys = append(keys, int(indexs[i]))
	}

	sort.Ints(keys)

	for _, key := range keys {
		newHashs = append(newHashs, m[key])
		newIndexs = append(newIndexs, uint64(key))
	}
	uploadFile.Hashs = newHashs
	uploadFile.Indexs = newIndexs
	return true

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

func DeserializeUploadFile(blockBytes []byte) *UploadFile {
	var uploadFile UploadFile
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&uploadFile)
	if err != nil {
		panic(err)
	}
	return &uploadFile
}
