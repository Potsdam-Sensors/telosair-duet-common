package telosairduetcommon

type DuetSensorState struct {
	Val uint8
}

func (s DuetSensorState) DirectoryName() string {
	return ""
}

func (s DuetSensorState) DirectoryData() map[string]float32 {
	return map[string]float32{
		"sensor_states": float32(s.Val),
	}
}
