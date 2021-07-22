package plot

import (
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

func GetMinStartNonce(numericId, path string) (uint64, error) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}

	type noncePair struct {
		startNonce uint64
		nonce      uint64
	}

	pairs := make([]noncePair, 0)

	for _, info := range infos {
		if !strings.HasPrefix(info.Name(), numericId) {
			continue
		}
		parts := strings.Split(info.Name(), "_")

		if len(parts) != 3 {
			continue
		}

		startNonce, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		nonce, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			continue
		}
		pairs = append(pairs, noncePair{
			startNonce: startNonce,
			nonce:      nonce,
		})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].startNonce < pairs[j].startNonce
	})

	startNonce := uint64(0)
	for i, p := range pairs {
		if i == 0 {
			startNonce = p.startNonce + p.nonce
			continue
		}
		if startNonce == p.startNonce {
			startNonce = p.startNonce + p.nonce
			continue
		}
		break

	}

	return startNonce, nil
}
