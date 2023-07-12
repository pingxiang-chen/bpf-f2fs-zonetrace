package znsmemory

const SegmentSize = 512

type SegmentType int

const (
	UnknownSegment SegmentType = iota - 1

	// HotDataSegment is type for frequently accessed data segments
	HotDataSegment

	// WarmDataSegment is type for commonly accessed data segments
	WarmDataSegment

	// ColdDataSegment is type for infrequently accessed data segments
	ColdDataSegment

	// HotNodeSegment is type for frequently accessed node segments
	HotNodeSegment

	// WarmNodeSegment is type for commonly accessed node segments
	WarmNodeSegment

	// ColdNodeSegment is type for infrequently accessed node segments
	ColdNodeSegment
)

func (t SegmentType) IsValid() bool {
	return HotDataSegment <= t && t <= ColdNodeSegment
}

type ValidMap []byte

type Segment struct {
	ValidMap ValidMap
}

type SegmentId struct {
	ZoneNo    int
	SegmentNo int
}

type SitEntryUpdate struct {
	SegmentNo   int
	ZoneNo      int
	SegmentType SegmentType
	ValidMap    ValidMap
}

type ZoneInfo struct {
	TotalZone               int
	BlockPerSegment         int
	TotalBlockPerZone       int
	AvailableBlockPerZone   int
	TotalSegmentPerZone     int
	AvailableSegmentPerZone int
}

type Zone struct {
	ZoneNo   int
	Segments []Segment
}

type ZonedStorage struct {
	ZoneInfo
	Zones []Zone
}
