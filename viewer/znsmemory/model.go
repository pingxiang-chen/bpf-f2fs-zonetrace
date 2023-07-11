package znsmemory

const SegmentSize = 512

type ValidMap []byte

type Segment struct {
	ValidMap ValidMap
}

type SegmentId struct {
	ZoneNo    int
	SegmentNo int
}

type UpdateSitEntry struct {
	SegmentNo int
	ZoneNo    int
	ValidMap  ValidMap
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
