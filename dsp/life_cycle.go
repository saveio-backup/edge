package dsp

import (
	"errors"
	"sync"

	"github.com/saveio/themis/common/log"
)

type LifeCycleState int

const (
	LifeCycleInit LifeCycleState = iota
	LifeCycleStart
	LifeCycleRun
	LifeCycleStop
	LifeCycleTerminate
)

type LifeCycle struct {
	state LifeCycleState
	lock  *sync.Mutex
}

func NewLifeCycle() *LifeCycle {
	return &LifeCycle{
		state: LifeCycleInit,
		lock:  new(sync.Mutex),
	}
}

func (this *LifeCycle) Start() error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.state != LifeCycleInit && this.state != LifeCycleTerminate {
		log.Errorf("can't stop a init state")
		return errors.New("wrong life cycle state")
	}
	this.state = LifeCycleStart
	return nil
}

func (this *LifeCycle) Run() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.state = LifeCycleRun
}

func (this *LifeCycle) Stop() error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.state != LifeCycleRun {
		log.Errorf("can't stop a un run endpoint")
		return errors.New("wrong life cycle state")
	}
	this.state = LifeCycleStop
	return nil
}

func (this *LifeCycle) Terminate() {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.state = LifeCycleTerminate
}
