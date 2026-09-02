package telosairduetcommon

import (
	"fmt"
	"strconv"
)

type AriaMeasurement struct {
	ProtocolVersion uint8
	McuSn string
	BinBoundsUpper []uint16
	BinCounts [][][]uint16
	SampleUs[] uint32
	OpcError[] uint8
}

func (m *AriaMeasurement) String() string {
	return fmt.Sprintf("AriaMeasurement: ProtocolVersion %d, McuSn %s, BinBoundsUpper %v, BinCounts %v, SampleUs %v, OpcError %v", m.ProtocolVersion, m.McuSn, m.BinBoundsUpper, m.BinCounts, m.SampleUs, m.OpcError)
}

/*
Convert the sample to a map, adding the suffix to the end of each key.
*/
func (m *AriaMeasurement) ToMap() map[string]any {
	// TODO
	return map[string]any{
		"aria_sn": m.McuSn,
	}
}

const aria_split_str_len_min = 6
/*
	Returns the number of fields consumed, or an error if the split string is invalid.
*/
func (m *AriaMeasurement) PopulateFromSplitString(split []string) (int, error) {
	if n := len(split); n < aria_split_str_len_min {
		return 0, fmt.Errorf("expected at least %d fields for AriaMeasurement, instead have: %d: %v", aria_split_str_len_min, n, split)
	}

	cur := 0

	// Protocol Version
	if protVer, err := strconv.ParseUint(split[cur], 10, 8); err != nil {
		return 0, fmt.Errorf("failed to convert protocol version to uint8 for AriaMeasurement, \"%s\": %v", split[cur], err)
	} else {
		m.ProtocolVersion = uint8(protVer)
	}
	cur++

	// MCU SN
	if mcuSnLen, err := strconv.ParseUint(split[cur], 10, 8); err != nil {
		return 0, fmt.Errorf("failed to convert mcuSnLen to uint8 for AriaMeasurement, \"%s\": %v", split[cur], err)
	} else {
		cur++
		if mcuSnLen > 0 {
			m.McuSn = split[cur]
		}
	}
	cur++

	// Block Count
	var blockCount uint8
	if blockCount_, err := strconv.ParseUint(split[cur], 10, 8); err != nil {
		return 0, fmt.Errorf("failed to convert blockCount to uint8 for AriaMeasurement, \"%s\": %v", split[cur], err)
	} else {
		blockCount = uint8(blockCount_)
	}
	cur++

	// Channel Count
	var channelCount uint8
	if channelCount_, err := strconv.ParseUint(split[cur], 10, 8); err != nil {
		return 0, fmt.Errorf("failed to convert channelCount to uint8 for AriaMeasurement, \"%s\": %v", split[cur], err)
	} else {
		channelCount = uint8(channelCount_)
	}
	cur++

	// Bin Count
	var binCount uint8
	if binCount_, err := strconv.ParseUint(split[cur], 10, 8); err != nil {
		return 0, fmt.Errorf("failed to convert binCount to uint8 for AriaMeasurement, \"%s\": %v", split[cur], err)
	} else {
		binCount = uint8(binCount_)
	}
	cur++

	// Bin Bounds Upper
	m.BinBoundsUpper = make([]uint16, binCount)
	for i := uint8(0); i < binCount; i++ {
		var val, err = strconv.ParseUint(split[cur], 10, 16)
		if err != nil {
			return 0, fmt.Errorf("failed to convert binBoundsUpper[%d] to uint16 for AriaMeasurement, \"%s\": %v", i, split[cur], err)
		}
		m.BinBoundsUpper[i] = uint16(val)
		cur++
	}

	// Bin Counts
	m.BinCounts = make([][][]uint16, blockCount)
	for b := uint8(0); b < blockCount; b++ {
		m.BinCounts[b] = make([][]uint16, channelCount)
		for c := uint8(0); c < channelCount; c++ {
			m.BinCounts[b][c] = make([]uint16, binCount)
			for i := uint8(0); i < binCount; i++ {
				var val, err = strconv.ParseUint(split[cur], 10, 16)
				if err != nil {
					return 0, fmt.Errorf("failed to convert binCounts[%d][%d][%d] to uint16 for AriaMeasurement, \"%s\": %v", b, c, i, split[cur], err)
				}
				m.BinCounts[b][c][i] = uint16(val)
				cur++
			}
		}
	}

	// Sample Us
	m.SampleUs = make([]uint32, blockCount)
	for b := uint8(0); b < blockCount; b++ {
		var val, err = strconv.ParseUint(split[cur], 10, 32)
		if err != nil {
			return 0, fmt.Errorf("failed to convert sampleUs[%d] to uint32 for AriaMeasurement, \"%s\": %v", b, split[cur], err)
		}
		m.SampleUs[b] = uint32(val)
		cur++
	}

	// Opc Error
	m.OpcError = make([]uint8, blockCount)
	for b := uint8(0); b < blockCount; b++ {
		var val, err = strconv.ParseUint(split[cur], 10, 8)
		if err != nil {
			return 0, fmt.Errorf("failed to convert opcError[%d] to uint8 for AriaMeasurement, \"%s\": %v", b, split[cur], err)
		}
		m.OpcError[b] = uint8(val)
		cur++
	}

	return cur, nil
}