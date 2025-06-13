package telosairduetcommon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

/*
This data type is of the Plantower PMS5003 OPC
*/
type Pms5003Measurement struct {
	PM1, PM2p5, PM10                    uint16
	PN0p3, PN0p5, PN1, PN2p5, PN5, PN10 uint16
}

func (m *Pms5003Measurement) String() string {
	return fmt.Sprintf("PM 1: %d, 2.5: %d, 10: %d | PN 0.3: %d, 0.5: %d, 1: %d, 2.5: %d, 5: %d, 10: %d",
		m.PM1, m.PM2p5, m.PM10, m.PN0p3, m.PN0p5, m.PN1, m.PN2p5, m.PN5, m.PN10)
}

/*
Convert the sample to a map, adding the suffix to the end of each key.
*/
func (m *Pms5003Measurement) ToMap(suffix string) map[string]any {
	return map[string]any{
		"pm10" + suffix:  m.PM1,
		"pm25" + suffix:  m.PM2p5,
		"pm100" + suffix: m.PM10,

		"pn03" + suffix:  m.PN0p3,
		"pn05" + suffix:  m.PN0p5,
		"pn10" + suffix:  m.PN1,
		"pn25" + suffix:  m.PN2p5,
		"pn50" + suffix:  m.PN5,
		"pn100" + suffix: m.PN10,
	}
}

/*
Merge two PMS5003 measurement valus, using mean if tha ratio is less than 2:1, min otherwise.
*/
func MergePTValue(v1 uint16, v2 uint16) uint16 {
	if v1*v2 == 0 {
		return 0
	}
	if ratio := float32(v1) / float32(v2); (ratio < .5) || (ratio > 2) {
		if v1 <= v2 {
			return v1
		}
		return v2
	}
	return (v1 + v2) / 2

}

/*
Use MergePTValue() to combine two PMS5003 measurements to represent the Duet's actual reported concentration info.
*/
func MergePT(pt1 *Pms5003Measurement, pt2 *Pms5003Measurement, ptResult *Pms5003Measurement) error {
	if (pt1 == nil) || (pt2 == nil) || (ptResult == nil) {
		return errors.New("an arg was nil")
	}
	ptResult.PM1 = MergePTValue(pt1.PM1, pt2.PM1)
	ptResult.PM2p5 = MergePTValue(pt1.PM2p5, pt2.PM2p5)
	ptResult.PM10 = MergePTValue(pt1.PM10, pt2.PM10)

	ptResult.PN0p3 = MergePTValue(pt1.PN0p3, pt2.PN0p3)
	ptResult.PN0p5 = MergePTValue(pt1.PN0p5, pt2.PN0p5)
	ptResult.PN1 = MergePTValue(pt1.PN1, pt2.PN1)
	ptResult.PN2p5 = MergePTValue(pt1.PN2p5, pt2.PN2p5)
	ptResult.PN5 = MergePTValue(pt1.PN5, pt2.PN5)
	ptResult.PN10 = MergePTValue(pt1.PN10, pt2.PN10)
	return nil
}

/*
Take the raw uint16s encoded as bytes and convert them, LittleEndian, to fill a PMS5003 measurement.
*/
func (m *Pms5003Measurement) PopulateFromBytes(buff []byte) error {
	if len(buff) < 18 {
		return fmt.Errorf("expected >%d bytes, got %d", 18, len(buff))
	}
	m.PM1 = binary.LittleEndian.Uint16(buff[0:2])
	m.PM2p5 = binary.LittleEndian.Uint16(buff[2:4])
	m.PM10 = binary.LittleEndian.Uint16(buff[4:6])
	m.PN0p3 = binary.LittleEndian.Uint16(buff[6:8])
	m.PN0p5 = binary.LittleEndian.Uint16(buff[8:10])
	m.PN1 = binary.LittleEndian.Uint16(buff[10:12])
	m.PN2p5 = binary.LittleEndian.Uint16(buff[12:14])
	m.PN5 = binary.LittleEndian.Uint16(buff[14:16])
	m.PN10 = binary.LittleEndian.Uint16(buff[16:18])
	return nil
}

func (p *Pms5003Measurement) PointerIterable() [9]*uint16 {
	return [9]*uint16{&p.PM1, &p.PM2p5, &p.PM10, &p.PN0p3, &p.PN0p5, &p.PN1, &p.PN2p5, &p.PN5, &p.PN10}
}

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

	if val, err := strconv.ParseUint(splitStr[1], 10, 16); err != nil {
		return convErr
	} else {
		p.PM2p5 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[2], 10, 16); err != nil {
		return convErr
	} else {
		p.PM10 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[3], 10, 16); err != nil {
		return convErr
	} else {
		p.PN0p3 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[4], 10, 16); err != nil {
		return convErr
	} else {
		p.PN0p5 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[5], 10, 16); err != nil {
		return convErr
	} else {
		p.PN1 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[6], 10, 16); err != nil {
		return convErr
	} else {
		p.PN2p5 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[7], 10, 16); err != nil {
		return convErr
	} else {
		p.PN5 = uint16(val)
	}

	if val, err := strconv.ParseUint(splitStr[8], 10, 16); err != nil {
		return convErr
	} else {
		p.PN10 = uint16(val)
	}

	return nil
}

func (p Pms5003Measurement) DirectoryName() string {
	return "pms5003"
}
func (m Pms5003Measurement) DirectoryData() map[string]float32 {
	return map[string]float32{
		"pm1":   float32(m.PM1),
		"pm2p5": float32(m.PM2p5),
		"pm10":  float32(m.PM10),

		"pn0p3": float32(m.PN0p3),
		"pn0p5": float32(m.PN0p5),
		"pn1":   float32(m.PN1),
		"pn2p5": float32(m.PN2p5),
		"pn5":   float32(m.PN5),
		"pn10":  float32(m.PN10),
	}
}
