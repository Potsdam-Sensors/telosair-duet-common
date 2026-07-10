package telosairduetcommon

import (
	"fmt"
)

const (
	KEY_PID_PPB    = "pid_ppb"
	KEY_PID_RAW_MV = "pid_raw_mv"
)

type PidMeasurement struct {
	Ppb   float32
	RawMv float32
}

// String provides a clean, readable representation of the PID data
func (m PidMeasurement) String() string {
	return fmt.Sprintf("%.2f ppb (%.2f mV)", m.Ppb, m.RawMv)
}

// ToMap serializes the data for database ingestion or payload generation
func (m PidMeasurement) ToMap() map[string]any {
	return map[string]any{
		KEY_PID_PPB:    m.Ppb,
		KEY_PID_RAW_MV: m.RawMv,
	}
}

// FloatMap provides a typed float32 map representation
func (m PidMeasurement) FloatMap() map[string]float32 {
	return map[string]float32{
		KEY_PID_PPB:    m.Ppb,
		KEY_PID_RAW_MV: m.RawMv,
	}
}

// DirectoryName matches the structure used by your storage drivers
func (m PidMeasurement) DirectoryName() string {
	return "pid"
}

// DirectoryData exposes the float32 dataset for folder/file logging
func (m PidMeasurement) DirectoryData() map[string]float32 {
	return m.FloatMap()
}