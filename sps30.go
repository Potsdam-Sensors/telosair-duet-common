package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Sps30Measurement = Pms5003Measurement

type Sps30FloatMeasurement struct {
	PM1, PM2p5, PM10                    float32
	PN0p3, PN0p5, PN1, PN2p5, PN5, PN10 float32
}

func (m *Sps30FloatMeasurement) String() string {
	return fmt.Sprintf("PM 1: %.1f, 2.5: %.1f, 10: %.1f | PN 0.3: %.1f, 0.5: %.1f, 1: %.1f, 2.5: %.1f, 5: %.1f, 10: %.1f",
		m.PM1, m.PM2p5, m.PM10, m.PN0p3, m.PN0p5, m.PN1, m.PN2p5, m.PN5, m.PN10)
}

/*
Convert the sample to a map, adding the suffix to the end of each key.
*/
func (m *Sps30FloatMeasurement) ToMap(suffix string) map[string]any {
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
Take the raw uint16s encoded as bytes and convert them, LittleEndian, to fill a PMS5003 measurement.
*/
func (m *Sps30FloatMeasurement) PopulateFromBytesReader(reader io.Reader) error {
	for p := range m.PointerIterable() {
		if err := binary.Read(reader, binary.LittleEndian, p); err != nil {
			return fmt.Errorf("failed to read a float: %w", err)
		}
	}
	return nil
}

/*
Take the raw uint16s encoded as bytes and convert them, LittleEndian, to fill a PMS5003 measurement.
*/
func (m *Sps30FloatMeasurement) PopulateFromBytes(buff []byte) error {
	if len(buff) < 36 {
		return fmt.Errorf("expected >%d bytes, got %d", 36, len(buff))
	}
	reader := bytes.NewReader(buff)
	return m.PopulateFromBytesReader(reader)
}

func (p *Sps30FloatMeasurement) PointerIterable() []any {
	return []any{&p.PM1, &p.PM2p5, &p.PM10, &p.PN0p3, &p.PN0p5, &p.PN1, &p.PN2p5, &p.PN5, &p.PN10}
}

func (p *Sps30FloatMeasurement) FromSerialString(s string) error {
	// Convert the string into a slice of strings of numbers
	splitStr := strings.Split(strings.Trim(s, "[]"), ",")

	// Make sure the length is correct.
	if len(splitStr) != 9 {
		return fmt.Errorf("expected list of length 9 for PlantowerData. Instead, have %d: %v", len(splitStr), splitStr)
	}

	// Try to convert each token from the slice to a float32
	convErr := fmt.Errorf("failed to convert a sps30 (float) data point to float32 from %v", splitStr)
	if val, err := strconv.ParseFloat(splitStr[0], 32); err != nil {
		return convErr
	} else {
		p.PM1 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[1], 32); err != nil {
		return convErr
	} else {
		p.PM2p5 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[2], 32); err != nil {
		return convErr
	} else {
		p.PM10 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[3], 32); err != nil {
		return convErr
	} else {
		p.PN0p3 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return convErr
	} else {
		p.PN0p5 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return convErr
	} else {
		p.PN1 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[6], 32); err != nil {
		return convErr
	} else {
		p.PN2p5 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[7], 32); err != nil {
		return convErr
	} else {
		p.PN5 = float32(val)
	}

	if val, err := strconv.ParseFloat(splitStr[8], 32); err != nil {
		return convErr
	} else {
		p.PN10 = float32(val)
	}

	return nil
}

func (p Sps30FloatMeasurement) DirectoryName() string {
	return "sps30"
}
func (m Sps30FloatMeasurement) DirectoryData() map[string]float32 {
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
