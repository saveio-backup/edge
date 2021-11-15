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
		sum        uint64
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
			sum:        startNonce + nonce,
		})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].sum < pairs[j].sum
	})
	startNonce := uint64(0)
	if len(pairs) > 0 {
		lastPlotFile := pairs[len(pairs)-1]
		startNonce = lastPlotFile.startNonce + lastPlotFile.nonce
	}
	return startNonce, nil
}
