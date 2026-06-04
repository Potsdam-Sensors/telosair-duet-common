package telosairduetcommon

import "fmt"

type AlphasenseOpcN3Measurement struct {
	PM1, PM2p5, PM10 float32
	Temp, Rh         float32
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
