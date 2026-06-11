package telosairduetcommon

import (
	"fmt"
	"strconv"
	"strings"
)

type AlphasenseOpcN3Measurement struct {
	PM1, PM2p5, PM10 float32
	Temp, Rh         float32
	Bins             [24]float32
}

/*
	func (p *Pms5003Measurement) FromSerialString(s string) error {
		// Convert the string into a slice of strings of numbers
		splitStr := strings.Split(strings.Trim(s, "[]"), ",")

		// Make sure the length is correct.
		if len(splitStr) != 9 {
			return fmt.Errorf("expected list of length 9 for PlantowerData. Instead, have %d: %v", len(splitStr), splitStr)
		}

		// Try to convert each token from the slice to a uint16 and store the value in the PlantowerData
		convErr := fmt.Errorf("failed to convert a plantower data point to uint16 from %v", splitStr)
		if val, err := strconv.ParseUint(splitStr[0], 10, 16); err != nil {
			return convErr
		} else {
			p.PM1 = uint16(val)
		}
*/
func (m AlphasenseOpcN3Measurement) PopulateBinsFromString(s string) error {
	splitStr := strings.Split(strings.Trim(s, "[]"), ",")

	if n := len(splitStr); n != 24 {
		return fmt.Errorf("expected list of length 24 for OPCN3 bins, instead have: %d: %v", n, splitStr)
	}

	for i, ss := range splitStr {
		if val, err := strconv.ParseFloat(ss, 32); err != nil {
			return fmt.Errorf("failed to convert a substring to float for OPCN3 bins, \"%s\": %v", ss, err)
		} else {
			m.Bins[i] = float32(val)
		}
	}

	return nil
}

func (m AlphasenseOpcN3Measurement) DirectoryName() string {
	return "alphasense-opc-n3"
}

func (m AlphasenseOpcN3Measurement) DirectoryData() map[string]float32 {
	return map[string]float32{
		"PM1":   m.PM1,
		"PM2.5": m.PM2p5,
		"PM10":  m.PM10,
		"TEMP":  m.Temp,
		"RH":    m.Rh,
	}
}

func (m AlphasenseOpcN3Measurement) String() string {
	return fmt.Sprintf("PM1 %.2f, PM2.5 %.2f, PM10 %.2f, Temp %.2f, RH %.2f", m.PM1, m.PM2p5, m.PM10, m.Temp, m.Rh)
}

func (m AlphasenseOpcN3Measurement) ToMapPm(suff string) map[string]any {
	return map[string]any{
		"pm10" + suff:  m.PM1,
		"pm25" + suff:  m.PM2p5,
		"pm100" + suff: m.PM10,
	}
}

func (m AlphasenseOpcN3Measurement) ToMapTempRh() map[string]any {
	return map[string]any{
		"temp_opcn3": m.Temp,
		"hum_opcn3":  m.Rh,
	}
}

func (m AlphasenseOpcN3Measurement) ToMapBins() map[string]any {
	ret := map[string]any{}
	for i, v := range m.Bins {
		ret[fmt.Sprintf("opc_bin%d", i)] = v
	}

	return ret
}
