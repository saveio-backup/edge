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

type UploadResultData struct {
	Index  int64  `json:"index"`
	FileId string `json:"fileId"`
}
type JsonResult struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var DB *leveldb.DB

func init() {
	db, err := leveldb.OpenFile("./p2p_http_file.db", nil)
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
	index, _ := strconv.ParseInt(r.PostFormValue("index"), 10, 64)
	// payment, _ := strconv.ParseInt(r.PostFormValue("index"), 10, 64) //只有支付成功才能到上传这里

	//data := r.PostFormValue("data") //将数据上传到fs模块 walletaddr fileId index

	//查询levelDb中是否含有该文件
	fileByte, err := DB.Get([]byte("up_"+walletAddr+fileId), nil)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			//查询数据不存在 进行创建
			uploadfile := uploadfile.UploadFile{
				WalletAddr: walletAddr,
				FileId:     fileId,
				CreateTime: time.Now(),
			}
			uploadfile.SliceArr = append(uploadfile.SliceArr, index)
			w.WriteHeader(200)
			uploadResultData:=UploadResultData{
				Index:index,
				FileId:fileId,
			}
			res, _ := json.Marshal(JsonResult{
				Code: 200,
				Msg:  "success",
				Data:uploadResultData,
			})
			w.Write(res)
		} else {
			panic(err)
		}
	}
	//加入已经存在
	file := uploadfile.DeserializeUploadFile(fileByte)

	file.SliceArr = append(file.SliceArr, index)
	fileByte = file.Serialize()
	err = DB.Put([]byte("up_"+walletAddr+fileId), fileByte, nil)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	uploadResultData:=UploadResultData{
		Index:index,
		FileId:fileId,
	}
	res, _ := json.Marshal(JsonResult{
		Code: 200,
		Msg:  "success",
		Data:uploadResultData,
	})
	w.Write(res)

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
		//1.扣费

		//2.fs模块获取数据 walletAddr fileId index

		//查询levelDb中是否含有该文件
		fileByte, err := DB.Get([]byte("down_"+walletAddr+id), nil)
		if err != nil {
			if err.Error() == "leveldb: not found" {
				//查询数据不存在 进行创建
				downloadFile := downloadfile.DownLoadFile{
					WalletAddr : walletAddr,
					FileId : fileId,
					Payment : 0.01,//测试费用
					Id : "45646", //测试id
					CreateTime: time.Now(),
				}
				downloadFile.SliceArr = append(downloadFile.SliceArr, index)

				downloadResultData := DownloadResultData{
					Data:   "返回的文件数据",
					Index:  index,
					FileId: fileId,
				}

				w.Header().Set("content-type", "application/json;charset=utf-8")
				res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: downloadResultData})
				w.Write(res)

			} else {
				panic(err)
			}
		}
		//假如已经存在
		file := downloadfile.DeserializeDownLoadFile(fileByte)
		file.Payment += 0.01//加上扣除的费用
		file.SliceArr = append(file.SliceArr, index)
		fileByte = file.Serialize()
		err = DB.Put([]byte("down_"+walletAddr+id), fileByte, nil)
		if err != nil {
			panic(err)
		}
		downloadResultData := DownloadResultData{
			Data:   "返回的文件数据",
			Index:  index,
			FileId: fileId,
		}
		w.Header().Set("content-type", "application/json;charset=utf-8")
		res, _ := json.Marshal(JsonResult{Code: 200, Msg: "success", Data: downloadResultData})
		w.Write(res)	
	}
}

func main() {
	http.HandleFunc("/upload", FileUpload)
	http.HandleFunc("/download", FileDownload)
	http.ListenAndServe(":8000", nil)
}
