package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

var GasSensorNames = []string{"co", "o3", "nh3", "no", "no2", "so2", "ch2o", "voc", "ch4"}

const (
	NUM_GAS_SENSORS = 9

	KEY_GAS_CO   = "co"
	KEY_GAS_O3   = "o3"
	KEY_GAS_NH3  = "nh3"
	KEY_GAS_NO   = "no"
	KEY_GAS_NO2  = "no2"
	KEY_GAS_SO2  = "so2"
	KEY_GAS_CH2O = "ch2o"
	KEY_GAS_VOC  = "voc"
	KEY_GAS_CH4  = "ch4"
	KEY_GAS_H2S  = "h2s"
)

type GasSensorsMeasurement struct {
	SensorBitField                            uint16
	Co, O3, Nh3, No, No2, So2, Ch2o, Voc, Ch4 float32
}

func checkBitSet(bitfield uint16, bitmask uint16) bool {
	return (bitfield & bitmask) != 0x00
}
func (m *GasSensorsMeasurement) PopulateFromPrimitive(floatSlice []float32) error {
	if n := len(floatSlice); n != NUM_GAS_SENSORS {
		return fmt.Errorf("number of floats provided, %d, does not match expected, %d", n, NUM_GAS_SENSORS)
	}

	for idx, floatAddr := range []*float32{&m.Co, &m.O3, &m.Nh3, &m.No, &m.No2, &m.So2, &m.Ch2o, &m.Voc, &m.Ch4} {
		bitmask := uint16(1) << idx
		if checkBitSet(m.SensorBitField, bitmask) {
			*floatAddr = floatSlice[idx]
		}
	}
	return nil
}
func (m *GasSensorsMeasurement) PopulateFromBytes(buff []byte) error {
	if n := len(buff); n != NUM_GAS_SENSORS*4 {
		return fmt.Errorf("expected %d bytes for gas sensors, got %d", NUM_GAS_SENSORS*4, n)
	}
	floats := make([]float32, NUM_GAS_SENSORS)
	reader := bytes.NewReader(buff)
	for i := 0; i < 9; i++ {
		var gasVal float32
		if err := binary.Read(reader, binary.LittleEndian, &gasVal); err != nil {
			return fmt.Errorf("error converting bytes to float (for gas val #%d): %w", i, err)
		}
		floats[i] = gasVal
	}
	return m.PopulateFromPrimitive(floats)
}

func (m *GasSensorsMeasurement) PopulateFromString(s string) error {
	subStrs := strings.Split(strings.Trim(strings.TrimSpace(s), "[]"), ",")
	if n := len(subStrs); n != NUM_GAS_SENSORS {
		return fmt.Errorf("expected %d comma-separated values, only got %d", NUM_GAS_SENSORS, n)
	}

	floats := make([]float32, NUM_GAS_SENSORS)
	for idx, valStr := range subStrs {
		if val, err := strconv.ParseFloat(valStr, 32); err != nil {
			return fmt.Errorf("error converting str, %s, to float: %w", valStr, err)
		} else {
			floats[idx] = float32(val)
		}
	}

	return m.PopulateFromPrimitive(floats)
}
func (m GasSensorsMeasurement) ToMap() map[string]any {
	var retMap = map[string]any{}
	if checkBitSet(m.SensorBitField, 1) {
		retMap[KEY_GAS_CO] = m.Co
	}
	if checkBitSet(m.SensorBitField, 2) {
		retMap[KEY_GAS_O3] = m.O3
	}
	if checkBitSet(m.SensorBitField, 4) {
		retMap[KEY_GAS_NH3] = m.Nh3
	}
	if checkBitSet(m.SensorBitField, 8) {
		retMap[KEY_GAS_NO] = m.No
	}
	if checkBitSet(m.SensorBitField, 16) {
		retMap[KEY_GAS_NO2] = m.No2
	}
	if checkBitSet(m.SensorBitField, 32) {
		retMap[KEY_GAS_SO2] = m.So2
	}
	if checkBitSet(m.SensorBitField, 64) {
		retMap[KEY_GAS_CH2O] = m.Ch2o
	}
	if checkBitSet(m.SensorBitField, 128) {
		retMap[KEY_GAS_VOC] = m.Voc
	}
	if checkBitSet(m.SensorBitField, 256) {
		retMap[KEY_GAS_CH4] = m.Ch4
	}
	return retMap
}

func (m GasSensorsMeasurement) FloatMap() map[string]float32 {
	var retMap = map[string]float32{}
	if checkBitSet(m.SensorBitField, 1) {
		retMap[KEY_GAS_CO] = m.Co
	}
	if checkBitSet(m.SensorBitField, 2) {
		retMap[KEY_GAS_O3] = m.O3
	}
	if checkBitSet(m.SensorBitField, 4) {
		retMap[KEY_GAS_NH3] = m.Nh3
	}
	if checkBitSet(m.SensorBitField, 8) {
		retMap[KEY_GAS_NO] = m.No
	}
	if checkBitSet(m.SensorBitField, 16) {
		retMap[KEY_GAS_NO2] = m.No2
	}
	if checkBitSet(m.SensorBitField, 32) {
		retMap[KEY_GAS_SO2] = m.So2
	}
	if checkBitSet(m.SensorBitField, 64) {
		retMap[KEY_GAS_CH2O] = m.Ch2o
	}
	if checkBitSet(m.SensorBitField, 128) {
		retMap[KEY_GAS_VOC] = m.Voc
	}
	if checkBitSet(m.SensorBitField, 256) {
		retMap[KEY_GAS_CH4] = m.Ch4
	}
	return retMap
}

func (m GasSensorsMeasurement) DirectoryName() string {
	return "gas"
}

func (m GasSensorsMeasurement) DirectoryData() map[string]float32 {
	return m.FloatMap()
}
