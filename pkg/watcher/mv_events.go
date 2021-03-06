package watcher

import (
	"sync"
	"time"
)

type mvFromEvent struct {
	cookie   int
	parentWd int
	name     string
	isDir    bool
	done     chan struct{}
}

type mvToEvent struct {
	cookie   int
	parentWd int
	name     string
}

type mvEvent struct {
	oldParentWd int
	newParentWd int
	oldName     string
	newName     string
	isDir       bool
}

type mvEvents struct {
	mx     sync.Mutex
	mvFrom map[int]*mvFromEvent
	queue  chan *mvEvent
	done   chan struct{}
}

func newMvEvents() *mvEvents {
	return &mvEvents{
		queue:  make(chan *mvEvent, 1),
		mvFrom: map[int]*mvFromEvent{},
		done:   make(chan struct{}),
	}
}

func (me *mvEvents) addMvFrom(cookie int, name string, parentWd int, isDir bool) {
	done := make(chan struct{})

	me.mx.Lock()
	me.mvFrom[cookie] = &mvFromEvent{
		cookie:   cookie,
		parentWd: parentWd,
		name:     name,
		isDir:    isDir,
		done:     done,
	}
	me.mx.Unlock()

	go func() {
		select {
		case <-done:
		case <-me.done:
		case <-time.After(time.Millisecond * 100):
			me.queue <- &mvEvent{
				oldParentWd: parentWd,
				oldName:     name,
				newParentWd: -1,
				isDir:       isDir,
			}
		}

		me.mx.Lock()
		delete(me.mvFrom, cookie)
		me.mx.Unlock()
	}()
}

func (me *mvEvents) addMvTo(cookie int, name string, parentWd int, isDir bool) {
	me.mx.Lock()
	mvFrom := me.mvFrom[cookie]
	me.mx.Unlock()

	if mvFrom != nil {
		me.mx.Lock()
		delete(me.mvFrom, cookie)
		me.mx.Unlock()

		close(mvFrom.done)

		me.queue <- &mvEvent{
			oldParentWd: mvFrom.parentWd,
			oldName:     mvFrom.name,
			newParentWd: parentWd,
			newName:     name,
			isDir:       isDir,
		}

		return
	}

	me.queue <- &mvEvent{
		oldParentWd: -1,
		newParentWd: parentWd,
		newName:     name,
	}
}

func (me *mvEvents) close() {
	me.mx.Lock()
	defer me.mx.Unlock()

	close(me.done)
}
