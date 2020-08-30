package dsp

import (
	"github.com/saveio/themis/common/log"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
)

//Dsp api
func (this *Endpoint) CreateSector(sectorId uint64, proveLevel uint64, size uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.CreateSector(sectorId, proveLevel, size)
	if err != nil {
		log.Errorf("create sector err:%s", err)
		return "", &DspErr{Code: DSP_NODE_REGISTER_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) DeleteSector(sectorId uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.DeleteSector(sectorId)
	if err != nil {
		log.Errorf("delete sector err:%s", err)
		// TODO: define error code for sector operations
		return "", &DspErr{Code: DSP_NODE_REGISTER_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) GetSectorInfo(sectorId uint64) (*fs.SectorInfo, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	sectorInfo, err := dsp.GetSectorInfo(sectorId)
	if err != nil {
		log.Errorf("get sectorInfo err:%s", err)
		// TODO: define error code for sector operations
		return nil, &DspErr{Code: DSP_NODE_REGISTER_FAILED, Error: err}
	}
	return sectorInfo, nil

}

func (this *Endpoint) GetSectorInfosForNode(addr string) (*fs.SectorInfos, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	sectorInfos, err := dsp.GetSectorInfosForNode(addr)
	if err != nil {
		log.Errorf("Get sectorInfos for node err:%s", err)
		// TODO: define error code for sector operations
		return nil, &DspErr{Code: DSP_NODE_REGISTER_FAILED, Error: err}
	}
	return sectorInfos, nil

}
