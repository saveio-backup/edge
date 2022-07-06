package transport

import (
	"errors"
	"fmt"
	"sync"

	"github.com/saveio/dsp-go-sdk/store"
	"github.com/saveio/themis/common/log"
)

// var DB *leveldb.DB
var LevelDBStore *store.LevelDBStore

//保证一笔交易只能下载一次
var LevelDBPaymedIDStore *store.LevelDBStore

// const (
//     TaskStateDoing  = iota // value --> 0
//     TaskStateDone              // value --> 1
//     TaskTypeUpload = iota            // value --> 0
//     TaskTypeDownload           // value --> 1
// )

func InitDB() {
	// db, err := leveldb.OpenFile("./p2p/http_file.db", nil)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("db init success")
	// DB = db
	dbStore, err := store.NewLevelDBStore("./p2p/http_file.db")
	if err != nil {
		panic(err)
	}
	LevelDBStore = dbStore
	fmt.Println("db init success")
	levelDBPaymedIDStore, err := store.NewLevelDBStore("./p2p/http_paymentId.db")
	if err != nil {
		panic(err)
	}
	LevelDBPaymedIDStore = levelDBPaymedIDStore
	fmt.Println("LevelDBPaymedIDStore init success")
}

type Task struct {
	Lock *sync.RWMutex // lock
}

func NewTask() *Task {
	return &Task{
		Lock: &sync.RWMutex{},
	}
}

//query upload task record
func (task *Task) UploadTaskRecord(peerAddr string, taskState string) []*UploadFile {
	var uploadFileList []*UploadFile
	t := "_up"
	s := "_doing_"
	if taskState == "1" {
		s = "_done_"
	}
	prefix := []byte(peerAddr + t + s)
	keys, err := LevelDBStore.QueryStringKeysByPrefix(prefix)
	if err != nil {
		log.Fatalf("Query Keys By Prefix err: %s", err)
	}
	if len(keys) > 0 {
		uploadFileList = task.GetUploadFileListByStringKeys(keys)
	}
	return uploadFileList
}

//query download task record
func (task *Task) DownloadTaskRecord(peerAddr string, taskState string) []*DownloadFile {
	var downloadFileList []*DownloadFile
	t := "_down"
	s := "_doing_"
	if taskState == "1" {
		s = "_done_"
	}
	prefix := []byte(peerAddr + t + s)
	keys, err := LevelDBStore.QueryStringKeysByPrefix(prefix)
	if err != nil {
		log.Fatalf("Query Keys By Prefix err: %s", err)
	}
	if len(keys) > 0 {
		downloadFileList = task.GetDownloadFileListByStringKeys(keys)
	}
	return downloadFileList
}

//create uoload transport task
func (task *Task) CreateUploadTask(peerAddr string, fileHash string, prefix string) error {
	return CreateUploadFile(peerAddr, fileHash, prefix)
}

//create download transport task
func (task *Task) CreateDownloadTask(peerAddr string, fileHash string) (int32, error) {
	task.Lock.Lock()
	defer task.Lock.Unlock()

	downlaodTaskId, err := CreateDownloadTask(peerAddr)
	if err != nil {
		log.Errorf("peerAddr:%s  get fileHash: %s create download task err: %s", peerAddr, fileHash, err)
		return 0, err
	}
	return downlaodTaskId, nil
}
func (task *Task) GetPaymentId(hashs []string) (int32, int, error) {
	task.Lock.Lock()
	defer task.Lock.Unlock()
	return GetPaymentId(hashs)
}

// transport upload block record to save local database
func (task *Task) UploadPutBlocks(blkInfos []*UploadFileDetail) error {
	task.Lock.Lock()
	defer task.Lock.Unlock()
	file, err := GetUploadFile(blkInfos[0].PeerAddr, blkInfos[0].FileHash)
	if err != nil {
		log.Debugf("GetUploadFile err: %s", err)
		return err
	}
	err = file.UploadPutBlocks(blkInfos)
	if err != nil {
		log.Debugf("upload put blocks err: %s", err)
		return err
	}
	return nil
}

func (task *Task) GetUploadBlockHashs(fileOwnerAddr string, fileHash string) ([]string, []uint64, string, error) {

	task.Lock.Lock()
	defer task.Lock.Unlock()
	key := []byte(fileOwnerAddr + "_up_done_" + fileHash)

	fileBytes, err := LevelDBStore.Get(key)
	if err != nil {
		return nil, nil, "", err
	}
	file := DeserializeUploadFile(fileBytes)
	if file.IsSort {
		return file.Hashs, file.Indexs, file.Prefix, nil
	} else {
		if file.HashsSort() {
			file.IsSort = true
			fileBytes = file.Serialize()
			err = LevelDBStore.Put(key, fileBytes)
			if err != nil {
				return nil, nil, "", err
			}
			return file.Hashs, file.Indexs, file.Prefix, nil
		} else {
			return nil, nil, "", errors.New("file hashs sort error")
		}
	}

}

// transport Download block record to save local database
func (task *Task) DownloadPutBlocks(blkInfos []*DownloadFileDetail, paymentId int32) {
	task.Lock.Lock()
	defer task.Lock.Unlock()
	file, err := GetDownloadFile(blkInfos[0].PeerAddr, blkInfos[0].DownloadTaskId)
	if err != nil {
		log.Debugf("GetDownloadFile err: %s", err)
		panic(err)
	}

	err = file.DownloadPutBlocks(blkInfos, paymentId)
	if err != nil {
		log.Debugf("download put blocks err: %s", err)
		panic(err)
	}
}

//transport Completed
func (task *Task) TransportCompleted(taskType string, peerAddr string, taskId string) error {
	//	修改任务状态 由传输中 修改为传输完毕
	t := "_down"

	s := "_doing_"
	//上传完毕
	if taskType == "0" {
		t = "_up"
	}
	key := []byte(peerAddr + t + s + taskId)
	fileByte, err := LevelDBStore.Get(key)
	if err != nil {
		log.Errorf("get peerAddr:%s taskId:%s transport data error :%s ", peerAddr, taskId, err)
		return err
	}
	if taskType == "1" {
		//下载
		downloadFile := DeserializeDownLoadFile(fileByte)
		downloadFile.Status = 1
		newKey := []byte(peerAddr + t + "_done_" + taskId)
		newFileByte := downloadFile.Serialize()
		err := LevelDBStore.Put(newKey, newFileByte)
		if err != nil {
			log.Errorf("put peerAddr:%s taskId:%s  state is over new data err:%s", peerAddr, taskId, err)
			return err
		}
		err = LevelDBStore.Delete(key)
		if err != nil {
			log.Errorf("delete peerAddr:%s taskId:%s old data  err:%s", peerAddr, taskId, err)
			return err
		}
		return nil
	} else {
		//上传
		fmt.Println("上传")
		uploadFile := DeserializeUploadFile(fileByte)
		uploadFile.Status = 1
		newKey := []byte(peerAddr + t + "_done_" + taskId)
		fmt.Println("1")
		newFileByte := uploadFile.Serialize()
		err := LevelDBStore.Put(newKey, newFileByte)
		if err != nil {
			log.Errorf("put peerAddr:%s taskId:%s  state is over new data err:%s", peerAddr, taskId, err)
			return err
		}
		fmt.Println("2")
		err = LevelDBStore.Delete(key)
		fmt.Println("3")
		if err != nil {
			log.Errorf("delete peerAddr:%s taskId:%s old data  err:%s", peerAddr, taskId, err)
			return err
		}
		return nil
	}

}

func (task *Task) DeleteTransportTask(taskTypes []string, transportState string, peerAddr string, taskIds []string) []bool {
	var isDeletes []bool
	//	修改任务状态 由传输中 修改为传输完毕
	t := "_up"
	s := "_doing_"
	if transportState == "1" {
		s = "_done_"
	}
	for i, taskType := range taskTypes {
		if taskType == "1" {
			t = "_down"
		}
		key := []byte(peerAddr + t + s + taskIds[i])
		err := LevelDBStore.Delete(key)
		if err != nil {
			log.Errorf("delete peerAddr:%s taskId:%s old data  err:%s", peerAddr, taskIds[i], err)
			isDeletes = append(isDeletes, false)
			//return isDelete, err
		} else {
			isDeletes = append(isDeletes, true)
		}
	}

	return isDeletes
}

//task upload detail
func (task *Task) GetUploadTaskDetail(peerAddr string, taskId string, taskState string) (*UploadFile, error) {
	s := "_doing_"
	if taskState == "1" {
		s = "_done_"
	}
	key := []byte(peerAddr + "_up" + s + taskId)
	fileByte, err := LevelDBStore.Get(key)
	if err != nil {
		log.Errorf("peerAddr:%s taskId:%s ,get transport upload task detail error :%s ", peerAddr, taskId, err)
		return nil, err
	}
	return DeserializeUploadFile(fileByte), nil

}

//task download detail
func (task *Task) GetDownloadTaskDetail(peerAddr string, taskId string, taskState string) (*DownloadFile, error) {
	s := "_doing_"
	if taskState == "1" {
		s = "_done_"
	}
	key := []byte(peerAddr + "_down" + s + taskId)
	fileByte, err := LevelDBStore.Get(key)
	if err != nil {
		log.Errorf("peerAddr:%s taskId:%s ,get transport download task detail error :%s  ", peerAddr, taskId, err)
		return nil, err
	}
	return DeserializeDownLoadFile(fileByte), nil
}

//get downloadFile list
func (task *Task) GetDownloadFileListByStringKeys(keys []string) []*DownloadFile {
	var downloadFileList []*DownloadFile
	for _, key := range keys {
		fileByte, err := LevelDBStore.Get([]byte(key))
		if err != nil {
			log.Fatalf("get file block err :%s", err)
		}
		downloadFileList = append(downloadFileList, DeserializeDownLoadFile(fileByte))
	}
	return downloadFileList
}

//get uploadFile list
func (task *Task) GetUploadFileListByStringKeys(keys []string) []*UploadFile {
	var uploadFileList []*UploadFile
	for _, key := range keys {
		fileByte, err := LevelDBStore.Get([]byte(key))
		if err != nil {
			log.Fatalf("get file block err :%s", err)
		}
		uploadFileList = append(uploadFileList, DeserializeUploadFile(fileByte))
	}
	return uploadFileList
}
