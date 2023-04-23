package transport

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"time"

	"github.com/saveio/themis/common/log"
)

type DownloadFile struct {
	DownloadTaskId int32
	PeerAddr       string
	TxHashs        []string
	Hashs          []string //file block hash
	NodeListMap    map[string][]string
	Indexs         []uint64 //downloaded index
	Status         int64    // 0 downloading 1 downloaded 2 delete 3 corrupt
	CreateTime     time.Time
}
type DownloadFileDetail struct {
	DownloadTaskId int32
	PeerAddr       string
	TxHash         string
	Hash           string //file block hash
	NodeList       []string
	Index          uint64
}

type Payment struct {
	//downloaded hash
	Hashs []string
}

func CreateDownloadTask(peerAddr string) (int32, error) {
	//create downloadTaskId
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	downloadTaskId := r.Int31()

	downloadFile := DownloadFile{
		PeerAddr:       peerAddr,
		DownloadTaskId: downloadTaskId,
		Status:         0,
		NodeListMap:    make(map[string][]string),
		CreateTime:     time.Now(),
	}
	fileByte := downloadFile.Serialize()
	//store download file
	err := LevelDBStore.Put([]byte(downloadFile.PeerAddr+"_down_doing_"+fmt.Sprint(downloadFile.DownloadTaskId)), fileByte)
	if err != nil {
		return 0, err
	}

	return downloadTaskId, nil
}

func GetPaymentId(hashs []string) (int32, int, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	paymentId := r.Int31()
	payment := Payment{
		Hashs: hashs,
	}
	paymentBytes := payment.Serialize()

	err := LevelDBPaymentIDStore.Put([]byte(fmt.Sprint(paymentId)), paymentBytes)
	if err != nil {
		return 0, 0, err
	}
	amount := len(hashs) * 256
	return paymentId, amount, nil
}

func FilterDownloadHash(paymentId int32, hashs []string) ([]int, []string) {
	paymentBytes, err := LevelDBPaymentIDStore.Get([]byte(fmt.Sprint(paymentId)))
	if err != nil {
		log.Error("get payment error:%s", err)
	}
	payment := DeserializePayment(paymentBytes)
	var unHaveHashs []string
	var haveHashIndexs []int
	for index, hash := range hashs {
		t := false
		for _, paymentedHash := range payment.Hashs {
			if paymentedHash == hash {
				t = true
			}
		}
		if t {
			haveHashIndexs = append(haveHashIndexs, index)
		} else {
			unHaveHashs = append(unHaveHashs, hash)
		}
	}
	return haveHashIndexs, unHaveHashs
}

func GetDownloadFile(peerAddr string, downloadTaskId int32) (*DownloadFile, error) {
	fileByte, err := LevelDBStore.Get([]byte(peerAddr + "_down_doing_" + fmt.Sprint(downloadTaskId)))
	if err != nil {
		return nil, err
	}
	file := DeserializeDownLoadFile(fileByte)
	return file, nil
}
func (downloadFile *DownloadFile) DownloadPutBlocks(blocks []*DownloadFileDetail, paymentId int32) error {
	for _, block := range blocks {
		downloadFile.Indexs = append(downloadFile.Indexs, block.Index)
		downloadFile.Hashs = append(downloadFile.Hashs, block.Hash)
		downloadFile.NodeListMap[block.Hash] = block.NodeList
	}
	fileByte := downloadFile.Serialize()
	err := LevelDBStore.Put([]byte(downloadFile.PeerAddr+"_down_doing_"+fmt.Sprint(downloadFile.DownloadTaskId)), fileByte)
	if err != nil {
		return err
	}

	paymentBytes, err := LevelDBPaymentIDStore.Get([]byte(fmt.Sprint(paymentId)))
	if err != nil {
		return err
	}
	payment := DeserializePayment(paymentBytes)
	payment.Hashs = nil
	paymentBytes = payment.Serialize()
	err = LevelDBPaymentIDStore.Put([]byte(fmt.Sprint(paymentId)), paymentBytes)
	if err != nil {
		return err
	}
	log.Debug("file download success")
	return nil
}
func (downloadFile *DownloadFile) DownloadBlock(block *DownloadFileDetail, paymentId int32) error {
	return downloadFile.DownloadPutBlocks([]*DownloadFileDetail{block}, paymentId)
}
func (downloadFile *DownloadFile) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(downloadFile)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func DeserializeDownLoadFile(blockBytes []byte) *DownloadFile {
	var downloadFile DownloadFile
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&downloadFile)
	if err != nil {
		panic(err)
	}
	return &downloadFile
}

func (payment *Payment) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(payment)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func DeserializePayment(blockBytes []byte) *Payment {
	var payment Payment
	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))
	err := decoder.Decode(&payment)
	if err != nil {
		panic(err)
	}
	return &payment
}
