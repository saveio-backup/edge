package dsp

import (
	"encoding/json"
	"fmt"

	"github.com/saveio/dsp-go-sdk/task"
)

const (
	PROGRESS_LIST_LEY = "edges:progress-info-listkey:"
	FILE_HASH_URL_LEY = "edges:file-hash-urlkey:"
)

func (this *Endpoint) AddProgress(v *task.ProgressInfo) error {
	key := genProgressKey(this.Account.Address.ToBase58(), v.TaskId)
	existed, err := this.GetProgress(key)
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
		allTasks = append([]string{key}, allTasks...)
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
	err = this.db.Put([]byte(key), progressBuf)
	if err != nil {
		return err
	}
	return nil
}

func genProgressKey(walletAddr, taskId string) string {
	return fmt.Sprintf("file-progress:%s-%s", walletAddr, taskId)
}

func (this *Endpoint) GetProgress(key string) (*task.ProgressInfo, error) {
	buf, err := this.db.Get([]byte(key))
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

func (this *Endpoint) DeleteProgress(taskIds []string) error {
	allProgressId := make([]string, 0)
	listKeys, err := this.db.Get([]byte(PROGRESS_LIST_LEY))
	if err == nil || len(listKeys) > 0 {
		err = json.Unmarshal(listKeys, &allProgressId)
		if err != nil {
			return err
		}
	}
	keyM := make(map[string]struct{}, 0)
	batch := this.db.NewBatch()
	for _, taskId := range taskIds {
		key := genProgressKey(this.Account.Address.ToBase58(), taskId)
		keyM[key] = struct{}{}
		this.db.BatchDelete(batch, []byte(key))
	}
	newProgressIds := make([]string, 0, len(allProgressId)-len(taskIds))
	for _, id := range allProgressId {
		_, ok := keyM[id]
		if ok {
			continue
		}
		newProgressIds = append(newProgressIds, id)
	}
	newTaskBuf, err := json.Marshal(newProgressIds)
	if err != nil {
		return err
	}
	this.db.BatchPut(batch, []byte(PROGRESS_LIST_LEY), newTaskBuf)
	err = this.db.BatchCommit(batch)
	return err
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
