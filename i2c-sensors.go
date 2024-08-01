package telosairduetcommon

import "fmt"

/* ~~ Temp & Rh ~~ */
type TempRhMeasurement interface {
	Temperature() float32
	Humidity() float32
	ToMap() map[string]any
}

type CombinedTempRhMeasurements struct {
	Temp, Hum float32
}

func (m CombinedTempRhMeasurements) String() string {
	return fmt.Sprintf("%.1fC, %.1f%", m.Temp, m.Hum)
}
func (m CombinedTempRhMeasurements) Temperature() float32 {
	return m.Temp
}
func (m CombinedTempRhMeasurements) Humidity() float32 {
	return m.Hum
}
func (m CombinedTempRhMeasurements) ToMap() map[string]any {
	return map[string]any{
		KEY_TEMP: m.Temp,
		KEY_HUM:  m.Hum,
	}
}

func CombineTempRhMeasurements(m1 TempRhMeasurement, m2 TempRhMeasurement, m3 *CombinedTempRhMeasurements) {
	m3.Temp = (m1.Temperature() + m2.Temperature()) / 2
	m3.Hum = (m1.Humidity() + m2.Humidity()) / 2
}

/* ~~ HTU ~~ */
type Htu21Measurement struct {
	Temp, Hum float32
}

func (m Htu21Measurement) String() string {
	return fmt.Sprintf("%.1fC %.1f%", m.Temp, m.Hum)
}
func (m Htu21Measurement) ToMap() map[string]any {
	return map[string]any{
		KEY_HTU_TEMP: m.Temp,
		KEY_HTU_HUM:  m.Hum,
	}
}
func (m Htu21Measurement) Temperature() float32 {
	return m.Temp
}
func (m Htu21Measurement) Humidity() float32 {
	return m.Hum
}

/* ~~ SCD 41 ~~ */
type Scd41Measurement struct {
	Temp, Hum float32
	Co2       uint16
}

func (m Scd41Measurement) String() string {
	return fmt.Sprintf("%.1fC, %.1f%, %dpm", m.Temp, m.Hum, m.Co2)
}
func (m Scd41Measurement) ToMap() map[string]any {
	return map[string]any{
		KEY_SCD_TEMP:       m.Temp,
		KEY_SCD_HUM:        m.Hum,
		KEY_SCD_CO2:        m.Co2,
		KEY_SCD_CO2_LEGACY: m.Co2,
	}
}
func (m Scd41Measurement) Temperature() float32 {
	return m.Temp
}
func (m Scd41Measurement) Humidity() float32 {
	return m.Hum
}

/* ~~ SG30 ~~ */
type Sgp30Measurement struct {
	Tvoc int32
}

func (m Sgp30Measurement) String() string {
	return fmt.Sprintf("%dppb", m.Tvoc)
}
func (m *Sgp30Measurement) ToMap() map[string]any {
	return map[string]any{
		KEY_TVOC: m.Tvoc,
	}
}

/* ~~ SGP40 ~~ */
type Sgp40Measurement struct {
	VocIndex uint32
}

func (m Sgp40Measurement) String() string {
	return fmt.Sprintf("%d", m.VocIndex)
}

func (m *Sgp40Measurement) ToMap() map[string]any {
	return map[string]any{
		KEY_VOC_INDEX: m.VocIndex,
	}
}

/* ~~~ MPRLS ~~ */
type MprlsMeasurement struct {
	Pressure float32
}

func (m *MprlsMeasurement) String() string {
	return fmt.Sprintf("%.1fkPa", m.Pressure)
}

func (m *MprlsMeasurement) ToMap() map[string]any {
	return map[string]any{
		KEY_MPRLS_PRESSURE: m.Pressure,
	}
}
