package telosairduetcommon

import (
	"log"
	"testing"
)

func TestCombineTempRhMeasurements(t *testing.T) {
	m1 := Htu21Measurement{
		Temp: 10,
		Hum:  20,
	}
	m2 := Scd41Measurement{
		Temp: 20,
		Hum:  30,
	}
	m3 := CombinedTempRhMeasurements{}
	expected := CombinedTempRhMeasurements{
		Temp: 15,
		Hum:  25,
	}
	CombineTempRhMeasurements(m1, m2, &m3)
	log.Print(m3)

	if m3 != expected {
		t.Errorf(" = %v; want %v", m3, expected)
	}
}
