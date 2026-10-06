package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK4 Var 31 - 1 OPC-N3, 1 SPS, 1 PT ~~ */
var DuetTypeMk4Var31 = DuetTypeInfo{
	ExpectedBytes:        188,
	ExpectedStringLen:    21,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var31{} },
	TypeAlias:            "Mk4.31",
}

type DuetDataMk4Var31 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Sps       Pms5003Measurement
	Pt        Pms5003Measurement
	PtM       Pms5003Measurement
	Opc       AlphasenseOpcN3Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	timeResolved bool
}

func (d *DuetDataMk4Var31) TimeResolved() bool {
	return d.timeResolved
}

func (d *DuetDataMk4Var31) MarkTimeResolved(v bool) {
	d.timeResolved = v
}

func (d *DuetDataMk4Var31) Timestamp() uint32 {
	return d.UnixSec
}

func (d *DuetDataMk4Var31) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk4Var31) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{
		d.Pt,
		d.Sps,
		// d.PtM,
		d.TempRh,
		d.Scd,
		d.Mprls,
		d.Sgp,
		d.Opc,
		DuetSensorState{d.SensorStates},
	}
}

func (d *DuetDataMk4Var31) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}

func (d *DuetDataMk4Var31) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}

func (d *DuetDataMk4Var31) String() string {
	return fmt.Sprintf(
		"[Duet %d, Type 4.31 | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | PT: %s | SPS: %s | OPC: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber,
		d.UnixSec,
		d.TempRh.String(),
		d.Htu.String(),
		d.Scd.String(),
		d.Mprls.String(),
		d.Sgp.String(),
		d.Pt.String(),
		d.Sps.String(),
		d.Opc.String(),
		d.RadioMeta.String(),
		d.SensorStates,
		d.PoeUsbVoltage,
	)
}

func (d *DuetDataMk4Var31) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var31
}

func (d *DuetDataMk4Var31) SetConnectionType(ct int) {
	d.ConnectionType = ct
}

func (d *DuetDataMk4Var31) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) ||
		(unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf(
			"incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d",
			unixSecRecieved,
			d.RadioMeta.RadioSentTimeMs,
			d.SampleTimeMs,
		)
	}

	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var31) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var31) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk4Var31) doPopulateFromSubStrings(splitStr []string) error {
	cur := 0

	// Serial Number
	sn, err := strconv.ParseUint(splitStr[cur], 10, 16)
	if err != nil {
		return fmt.Errorf(
			"failed to convert DuetSerialNumber string, %s, to uint32",
			splitStr[cur],
		)
	}
	d.SerialNumber = uint16(sn)
	cur++

	// Sample Time
	st, err := strconv.ParseUint(splitStr[cur], 10, 32)
	if err != nil {
		return fmt.Errorf(
			"failed to convert SampleTime string, %s, to uint32",
			splitStr[cur],
		)
	}
	d.SampleTimeMs = uint32(st)
	cur++

	// SPS30, converted by firmware to PMS5003-shaped data
	if err := d.Sps.FromSerialString(splitStr[cur]); err != nil {
		return fmt.Errorf(
			"failed to convert SPS string, %s, to PlantowerData",
			splitStr[cur],
		)
	}
	cur++

	// PMS5003 / PT
	if err := d.Pt.FromSerialString(splitStr[cur]); err != nil {
		return fmt.Errorf(
			"failed to convert PT string, %s, to PlantowerData",
			splitStr[cur],
		)
	}
	cur++

	// if err := MergePT(&d.Pt, &d.Sps, &d.PtM); err != nil {
	// 	return fmt.Errorf("failed to merge pt: %w", err)
	// }

	// Temperatures
	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert htu temp string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Htu.Temp = float32(temp)
	}
	cur++

	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert scd temp string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Scd.Temp = float32(temp)
	}
	cur++

	// Humidities
	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert htu hum string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Htu.Hum = float32(hum)
	}
	cur++

	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert scd hum string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Scd.Hum = float32(hum)
	}
	cur++

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert pressure string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Mprls.Pressure = float32(press)
	}
	cur++

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[cur], 10, 32); err != nil {
		return fmt.Errorf(
			"failed to convert voc index string, %s, to uint32",
			splitStr[cur],
		)
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}
	cur++

	// CO2
	if co2, err := strconv.ParseUint(splitStr[cur], 10, 16); err != nil {
		return fmt.Errorf(
			"failed to convert co2 string, %s, to uint16",
			splitStr[cur],
		)
	} else {
		d.Scd.Co2 = uint16(co2)
	}
	cur++

	// OPC-N3 temperature and humidity
	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert opc temp string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Opc.Temp = float32(temp)
	}
	cur++

	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert opc hum string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Opc.Rh = float32(hum)
	}
	cur++

	// OPC-N3 PM1, PM2.5, PM10
	if pm1, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert opc pm1 string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Opc.PM1 = float32(pm1)
	}
	cur++

	if pm2p5, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert opc pm2.5 string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Opc.PM2p5 = float32(pm2p5)
	}
	cur++

	if pm10, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf(
			"failed to convert opc pm10 string, %s, to float32",
			splitStr[cur],
		)
	} else {
		d.Opc.PM10 = float32(pm10)
	}
	cur++

	if err := d.Opc.PopulateBinsFromString(splitStr[cur]); err != nil {
		return fmt.Errorf("failed to convert opcn3 bins: %v", err)
	}
	cur++

	// PoE / USB Voltage
	if voltage, err := strconv.ParseUint(splitStr[cur], 10, 8); err != nil {
		return fmt.Errorf(
			"failed to convert voltage string, %s, to uint8",
			splitStr[cur],
		)
	} else {
		d.PoeUsbVoltage = uint8(voltage)
	}
	cur++

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[cur], 10, 8); err != nil {
		return fmt.Errorf(
			"failed to convert states string, %s, to uint8",
			splitStr[cur],
		)
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	cur++

	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}

func (d *DuetDataMk4Var31) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	d.PoeUsbVoltage = buff[1]
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Scd.Co2 = binary.LittleEndian.Uint16(buff[4:6])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[6:10])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[10:14])

	reader := bytes.NewReader(buff[14:54])

	if err := binary.Read(reader, binary.LittleEndian, &d.Htu.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Scd.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Htu.Hum); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Scd.Hum); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Mprls.Pressure); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.Rh); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.PM1); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.PM2p5); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Opc.PM10); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}

	// OPC bins occupy bytes 54:150.
	// The existing Var24 binary parser does not currently populate
	// the bins from the binary payload, so retain that behavior here.

	// SPS30 converted to PMS5003-shaped data.
	if err := d.Sps.PopulateFromBytes(buff[150:168]); err != nil {
		return fmt.Errorf("error parsing bytes for sps: %w", err)
	}

	// PMS5003.
	if err := d.Pt.PopulateFromBytes(buff[168:186]); err != nil {
		return fmt.Errorf("error parsing bytes for pt: %w", err)
	}

	// if err := MergePT(&d.Pt, &d.Sps, &d.PtM); err != nil {
	// 	return fmt.Errorf("failed to merge pt: %w", err)
	// }

	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}

func (d *DuetDataMk4Var31) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.31,
		KEY_SERIAL_NUMBER:   d.SerialNumber,
		KEY_DEVICE_ID:       d.SerialNumber,
		KEY_UNIX:            d.UnixSec,
		KEY_ECO2:            0,
		KEY_RAWH2:           0,
		KEY_SENSOR_STATES:   d.SensorStates,
		KEY_CONNECTION_TYPE: d.ConnectionType,
		KEY_LAST_RESET_TIME: d.LastResetUnix,
		KEY_GATEWAY_SERIAL:  gatewaySerial,
		KEY_POE_USB_VOLTAGE: d.PoeUsbVoltage,
	}

	maps.Copy(ret, d.Pt.ToMap("_t"))
	maps.Copy(ret, d.Sps.ToMap("_b"))
	maps.Copy(ret, d.Opc.ToMapPm("_m"))

	// maps.Copy(ret, d.Opc.ToMapPm("_t"))
	// maps.Copy(ret, d.Opc.ToMapPm("_b"))
	// maps.Copy(ret, d.Opc.ToMapPm("_m"))
	// maps.Copy(ret, d.Opc.ToMapPm("_opc"))
	maps.Copy(ret, d.Opc.ToMapBins())
	maps.Copy(ret, d.Opc.ToMapTempRh())

	maps.Copy(ret, d.Htu.ToMap())
	maps.Copy(ret, d.Scd.ToMap())
	maps.Copy(ret, d.TempRh.ToMap())
	maps.Copy(ret, d.Mprls.ToMap())
	maps.Copy(ret, d.Sgp.ToMap())
	maps.Copy(ret, d.RadioMeta.ToMap())

	if d.piMcuTempSet {
		ret[KEY_PI_MCU_TEMP] = d.PiMcuTemp
	}

	return ret
}
