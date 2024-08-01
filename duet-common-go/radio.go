package duetcommongo

type RadioMetadata struct {
	LastSnr         int32
	LastRssi        int16
	Hops            uint8
	RadioSentTimeMs uint32
}

func (m *RadioMetadata) ToMap() map[string]any {
	if m.LastSnr == 0 {
		return nil
	}
	return map[string]any{
		KEY_RSSI: m.LastRssi,
		KEY_SNR:  m.LastSnr,
		KEY_HOPS: m.Hops,
	}
}
