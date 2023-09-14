package znsmemory

import (
	"context"
	"fmt"
	"sync"
)

var _ ZNSMemory = (*memory)(nil)

// ZNSMemory is an interface that stores all states of a ZNS SSD and provides event subscription for state updates.
type ZNSMemory interface {
	// UpdateSegment updates a ZNS memory segment based on the provided SitEntryUpdate.
	UpdateSegment(updateSitEntry *SitEntryUpdate)

	// GetZone retrieves a zone by its zone number.
	GetZone(zoneNum int) (*Zone, error)

	// GetSegment retrieves a segment by its zone and segment numbers.
	GetSegment(zoneNum, segmentNum int) (*Segment, error)

	// GetZoneInfo returns information about the ZNS memory.
	GetZoneInfo() *ZoneInfo

	// Subscribe creates a new subscriber for ZNS state updates.
	Subscribe() *Subscriber

	// UnSubscribe removes a subscriber from the list of subscribers.
	UnSubscribe(sub *Subscriber)
}

// Subscriber is a struct representing a subscriber for ZNS memory events.
type Subscriber struct {
	Event chan SegmentId
}

// memory is the implementation of the ZNSMemory interface.
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

// startEventLoop starts a goroutine to handle ZNS memory events.
func (m *memory) startEventLoop(ctx context.Context) {
	maxSegmentFullNo := m.zns.TotalSegmentPerZone * m.zns.TotalZone
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

				if updateSitEntry.SegmentFullNo < 0 || maxSegmentFullNo < updateSitEntry.SegmentFullNo {
					fmt.Printf("invalid segment full no %d\n", updateSitEntry.SegmentFullNo)
					continue
				}
				segmentNo := updateSitEntry.SegmentFullNo % m.zns.TotalSegmentPerZone

				if m.zns.Zones[updateSitEntry.ZoneNo].LastSegmentType != updateSitEntry.SegmentType {
					m.zns.Zones[updateSitEntry.ZoneNo].LastSegmentType = updateSitEntry.SegmentType
				}

				m.zns.Zones[updateSitEntry.ZoneNo].Segments[segmentNo] = Segment{
					ValidMap:    updateSitEntry.ValidMap,
					SegmentType: updateSitEntry.SegmentType,
				}
				func() {
					m.subscriberMutex.RLock()
					defer m.subscriberMutex.RUnlock()
					for i := range m.subscribers {
						sub := m.subscribers[i]
						go func() {
							sub.Event <- SegmentId{
								ZoneNo:      updateSitEntry.ZoneNo,
								SegmentNo:   segmentNo,
								SegmentType: updateSitEntry.SegmentType,
							}
						}()
					}
				}()
			}
		}
	}()
}

// New creates a new instance of ZNSMemory with the provided context and ZoneInfo.
func New(ctx context.Context, info ZoneInfo) ZNSMemory {
	// Initialize ZNS memory zones based on ZoneInfo.
	zones := make([]Zone, 0, info.TotalZone)
	for i := 0; i < info.TotalZone; i++ {
		zones = append(zones, Zone{
			ZoneNo:          i,
			Segments:        make([]Segment, info.TotalSegmentPerZone),
			LastSegmentType: UnknownSegment,
		})
	}
	zns := ZonedStorage{
		ZoneInfo: info,
		Zones:    zones,
	}
	// Create a memory instance and start the event loop.
	m := &memory{
		zns:             zns,
		updateSitCh:     make(chan *SitEntryUpdate, 1024),
		subscribers:     nil,
		subscriberMutex: sync.RWMutex{},
	}
	m.startEventLoop(ctx)
	return m
}
