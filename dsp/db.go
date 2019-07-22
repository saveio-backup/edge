package dsp

import (
	"encoding/json"

	"github.com/saveio/dsp-go-sdk/task"
)

const (
	PROGRESS_LIST_LEY = "edges:progressinfo-listkey:"
	FILE_HASH_URL_LEY = "edges:file-hash-urlkey:"
)

func (this *Endpoint) AddProgress(v *task.ProgressInfo) error {
	existed, err := this.GetProgress(v.TaskId)
	if existed != nil && err == nil && existed.Result != nil {
		return nil
	}
	if existed == nil || err != nil {
		allTasks := make([]string, 0)
		listKeys, err := this.db.Get([]byte(PROGRESS_LIST_LEY))
		if err == nil || len(listKeys) > 0 {
			err = json.Unmarshal(listKeys, &allTasks)
			if err != nil {
				allTasks = make([]string, 0)
			}
		}
		allTasks = append([]string{v.TaskId}, allTasks...)
		allTaskBuf, err := json.Marshal(allTasks)
		if err != nil {
			return err
		}
		err = this.db.Put([]byte(PROGRESS_LIST_LEY), allTaskBuf)
		if err != nil {
			return err
		}
	}
	progressBuf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	err = this.db.Put([]byte(v.TaskId), progressBuf)
	if err != nil {
		return err
	}
	return nil
}

func (this *Endpoint) GetProgress(taskId string) (*task.ProgressInfo, error) {
	buf, err := this.db.Get([]byte(taskId))
	if err != nil {
		return nil, err
	}
	p := &task.ProgressInfo{}
	err = json.Unmarshal(buf, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (this *Endpoint) GetAllProgressKeys() ([]string, error) {
	buf, err := this.db.Get([]byte(PROGRESS_LIST_LEY))
	if err != nil {
		return nil, err
	}
	var p []string
	err = json.Unmarshal(buf, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (this *Endpoint) SetUrlForHash(hash, url string) error {
	key := FILE_HASH_URL_LEY + hash
	return this.db.Put([]byte(key), []byte(url))
}

func (this *Endpoint) GetUrlFromHash(hash string) (string, error) {
	key := FILE_HASH_URL_LEY + hash
	url, err := this.db.Get([]byte(key))
	if err != nil {
		return "", err
	}
	return string(url), nil
}
