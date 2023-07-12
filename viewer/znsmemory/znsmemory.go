package znsmemory

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

type ZNSMemory interface {
	UpdateSegment(updateSitEntry *SitEntryUpdate)
	GetZone(zoneNum int) (*Zone, error)
	GetSegment(zoneNum, segmentNum int) (*Segment, error)
	GetZoneInfo() *ZoneInfo
	Subscribe() *Subscriber
	UnSubscribe(sub *Subscriber)
	StartReceiveTrace(ctx context.Context, r *bufio.Reader)
}

type Subscriber struct {
	Event chan SegmentId
}

type memory struct {
	zns             ZonedStorage
	updateSitCh     chan *SitEntryUpdate
	subscribers     []*Subscriber
	subscriberMutex sync.RWMutex
	isReceiving     bool
}

func (m *memory) GetZoneInfo() *ZoneInfo {
	return &m.zns.ZoneInfo
}

func (m *memory) UpdateSegment(updateSitEntry *SitEntryUpdate) {
	m.updateSitCh <- updateSitEntry
}

func (m *memory) GetZone(zoneNum int) (*Zone, error) {
	if zoneNum < 0 || zoneNum >= m.zns.TotalZone {
		return nil, fmt.Errorf("zone %d not found", zoneNum)
	}
	return &m.zns.Zones[zoneNum], nil
}

func (m *memory) GetSegment(zoneNum, segmentNum int) (*Segment, error) {
	zone, err := m.GetZone(zoneNum)
	if err != nil {
		return nil, err
	}
	if segmentNum < 0 || segmentNum >= len(zone.Segments) {
		return nil, fmt.Errorf("segment %d not found", segmentNum)
	}
	return &zone.Segments[segmentNum], nil
}

func (m *memory) Subscribe() *Subscriber {
	m.subscriberMutex.Lock()
	defer m.subscriberMutex.Unlock()
	sub := &Subscriber{Event: make(chan SegmentId, 1024)}
	m.subscribers = append(m.subscribers, sub)
	return sub
}

func (m *memory) UnSubscribe(sub *Subscriber) {
	m.subscriberMutex.Lock()
	defer m.subscriberMutex.Unlock()
	for i := range m.subscribers {
		if m.subscribers[i] == sub {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			return
		}
	}
}

func (m *memory) StartReceiveTrace(ctx context.Context, r *bufio.Reader) {
	if m.isReceiving {
		panic("already receiving trace")
	}
	m.isReceiving = true
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			u, err := ReadSitEntryUpdate(r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("input closed")
					return
				}
				fmt.Printf("readSegment: %v\n", err)
				continue
			}
			m.UpdateSegment(u)
		}
	}()
}

func (m *memory) startEventLoop(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updateSitEntry := <-m.updateSitCh:
				if updateSitEntry.ZoneNo < 0 || updateSitEntry.ZoneNo >= m.zns.TotalZone {
					fmt.Printf("invalid zone no %d\n", updateSitEntry.ZoneNo)
					continue
				}
				if updateSitEntry.SegmentNo < 0 || updateSitEntry.SegmentNo >= m.zns.TotalSegmentPerZone {
					fmt.Printf("invalid segment no %d\n", updateSitEntry.SegmentNo)
					continue
				}
				m.zns.Zones[updateSitEntry.ZoneNo].Segments[updateSitEntry.SegmentNo].ValidMap = updateSitEntry.ValidMap
				func() {
					m.subscriberMutex.RLock()
					defer m.subscriberMutex.RUnlock()
					for i := range m.subscribers {
						sub := m.subscribers[i]
						go func() {
							sub.Event <- SegmentId{
								ZoneNo:    updateSitEntry.ZoneNo,
								SegmentNo: updateSitEntry.SegmentNo,
							}
						}()
					}
				}()
			}
		}
	}()
}

func New(ctx context.Context, info ZoneInfo) ZNSMemory {
	zones := make([]Zone, 0, info.TotalZone)
	for i := 0; i < info.TotalZone; i++ {
		zones = append(zones, Zone{
			ZoneNo:   i,
			Segments: make([]Segment, info.TotalSegmentPerZone),
		})
	}
	zns := ZonedStorage{
		ZoneInfo: info,
		Zones:    zones,
	}
	m := &memory{
		zns:             zns,
		updateSitCh:     make(chan *SitEntryUpdate, 1024),
		subscribers:     nil,
		subscriberMutex: sync.RWMutex{},
	}
	m.startEventLoop(ctx)
	return m
}
