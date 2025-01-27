package telosairduetcommon

import (
	"testing"
)

func TestSensorsImplementSensor(t *testing.T) {
	// Compile-time checks:
	// If HTU21DF or SCD41 do NOT implement sensors.Sensor,
	// the code won’t compile, and you'll get an error.
	var _ SensorMeasurement = &CombinedTempRhMeasurements{}
	var _ SensorMeasurement = &Htu21Measurement{}
	var _ SensorMeasurement = &Scd41Measurement{}
	var _ SensorMeasurement = &Sgp30Measurement{}
	var _ SensorMeasurement = &Sgp40Measurement{}
	var _ SensorMeasurement = &MprlsMeasurement{}

}
