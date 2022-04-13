package main

import (
	"encoding/json"
	downloadfile "github.com/saveio/edge/p2p/http/download_file"
	uploadfile "github.com/saveio/edge/p2p/http/upload_file"
	"net/http"
	"strconv"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

type DownloadResultData struct {
	Index  int64  `json:"index"`
	FileId string `json:"fileId"`
	Data   string `json:"data"`
}
type JsonResult struct {
	Code int                `json:"code"`
	Msg  string             `json:"msg"`
	Data DownloadResultData `json:"data"`
}

var DB *leveldb.DB

func init() {
	db, err := leveldb.Opile("./p2p_http_file.db", nil)
	if err != nil {
		panic(err)
	}
	DB = db
}

func FileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}
	//walletAddr fileId sliceSize index data
	walletAddr := r.PostFormValue("walletAddr")
	fileId := r.PostFormValue("fileId")
	sliceSize, _ := strconv.ParseInt(r.PostFormValue("scliceSize"), 10, 64)
	index, _ := strconv.ParseInt(r.PostFormValue("index"), 10, 64)
	//data := r.PostFormValue("data") //将数据上传到fs模块 walletaddr fileId index

	//查询levelDb中是否含有该文件
	fileByte, err := DB.Get([]byte("up_"+walletAddr+fileId), nil)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			//查询数据不存在 进行创建
			uploadfile := uploadfile.UploadFile{
				WalletAddr: walletAddr,
				FileId:     fileId,
				SliceSize:  sliceSize,
				SliceArr:   make([]int64, 100),
				CreateTime: time.Now(),
			}
			uploadfile.SliceArr = append(uploadfile.SliceArr, index)
			w.WriteHeader(200)
			res, _ := json.Marshal(JsonResult{
				Code: 200,
				Msg:  "success",
			})
			w.Write(res)
		} else {
			panic(err)
		}
	}
	//加入已经存在
	file := uploadfile.DeserializeUploadFile(fileByte)
	isHave := false
	for _, v := range file.SliceArr {
		if v == index {
			isHave = true
		}
	}
	if !isHave {
		file.SliceArr = append(file.SliceArr, index)
		fileByte = file.Serialize()
		err = DB.Put([]byte("up_"+walletAddr+fileId), fileByte, nil)
		w.WriteHeader(200)
		if err != nil {
			panic(err)
		}
	} else {
		w.WriteHeader(200)
		res, _ := json.Marshal(JsonResult{
			Code: 200,
			Msg:  "uploaded",
		})
		w.Write(res)
	}
}

func FileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		return
	}
	id := r.URL.Query().Get("id")
	walletAddr := r.URL.Query().Get("walletAddr")
	fileId := r.URL.Query().Get("fileId")
	index, err := strconv.ParseInt(r.URL.Query().Get("index"), 10, 64)

	if walletAddr == "" || fileId == "" || err != nil {
		w.WriteHeader(400)
	} else {
		//从fs模块获取数据 walletAddr fileId index
		//查询levelDb中是否含有该文件
		fileByte, err := DB.Get([]byte("down_"+walletAddr+id), nil)
		if err != nil {
			if err.Error() == "leveldb: not found" {
				//查询数据不存在 进行创建
				downloadFile := downloadfile.DownLoadFile{
					WalletAddr: walletAddr,
					FileId:     fileId,
					Id:         "45646", //测试id
					CreateTime: time.Now(),
				}
				downloadFile.SliceArr = append(downloadFile.SliceArr, index)
				w.WriteHeader(200)
				res, _ := json.Marshal(JsonResult{
					Code: 200,
					Msg:  "success",
				})
				w.Write(res)

			} else {
				panic(err)
			}
		}
		//加入已经存在
		file := downloadfile.DeserializeDownLoadFile(fileByte)
		isHave := false
		for _, v := range file.SliceArr {
			if v == index {
				isHave = true
			}
		}
		if !isHave {
			file.SliceArr = append(file.SliceArr, index)
			fileByte = file.Serialize()
			err = DB.Put([]byte("down_"+walletAddr+id), fileByte, nil)
			res, _ := json.Marshal(JsonResult{
				Code: 200,
				Msg:  "downloaded",
			})
			w.Write(res)
			if err != nil {
				panic(err)
			}
		} else {
			downloadResultData := DownloadResultData{
				Data:   "返回的文件数据",
				Index:  index,
				FileId: fileId,
			}
			w.Header().Set("content-type", "application/json;charset=utf-8")
			result, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: downloadResultData})
			w.Write(result)
		}
	}
}

func main() {
	http.HandleFunc("/upload", FileUpload)
	http.HandleFunc("/download", FileDownload)
	http.ListenAndServe(":8000", nil)
}
