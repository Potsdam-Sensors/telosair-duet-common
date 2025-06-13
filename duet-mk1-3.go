package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK1 Var 3 ~~ */
var DuetTypeMk1Var3 = DuetTypeInfo{
	ExpectedBytes:        58,
	ExpectedStringLen:    14,
	StructInstanceGetter: func() DuetData { return &DuetDataMk1Var3{} },
	TypeAlias:            "Mk1.3",
}

/*
	typedef struct __attribute__((packed, aligned(4))) {
	  uint8_t hardwareVersion;
	  uint8_t sensorVariation;
	  uint8_t sensorStates;
	  uint8_t _; //padding to align on 4, TODO: why not just remove `packed`?
	  uint16_t id;
	  uint16_t co2;
	  uint32_t tvoc;
	  unsigned long sample_time;
	  float temperature;
	  float humidity;
	  float pressure;
	  PMS5003::pms5003_data_t pt1;
	  PMS5003::pms5003_data_t pt2;
	} duet_sensor_reading_mk1_t;
*/
type DuetDataMk1Var3 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Sps Pms5003Measurement

	Si        Si7021Measurement
	Scd       Scd41Measurement
	Mprls     MprlsMeasurement
	Sgp30     Sgp30Measurement
	Sgp40     Sgp40Measurement
	RadioMeta RadioMetadata
}

func (d *DuetDataMk1Var3) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Sps, d.Si, d.Scd, d.Mprls, d.Sgp30, d.Sgp40, DuetSensorState{d.SensorStates}}
}

func (d *DuetDataMk1Var3) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk1Var3) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk1Var3) String() string {
	return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | Si7021 %s | SCD41 %s | MPRLS: %s | SGP's: %s, %s | SPS30: %s | Radio: %s | Errstate %d ]",
		d.SerialNumber, 1, 3, d.UnixSec, d.Si.String(), d.Scd.String(), d.Mprls.String(), d.Sgp30.String(), d.Sgp40, d.Sps.String(),
		d.RadioMeta.String(), d.SensorStates)
}
func (d *DuetDataMk1Var3) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk1Var3
}

func (d *DuetDataMk1Var3) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk1Var3) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk1Var3) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk1Var3) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk1Var3) doPopulateFromSubStrings(splitStr []string) error {
	// Serial Number
	idx := 0
	sn, err := strconv.ParseUint(splitStr[idx], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[idx])
	}
	d.SerialNumber = uint16(sn)
	idx += 1

	// Sample Time
	st, err := strconv.ParseUint(splitStr[idx], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[idx])
	}
	d.SampleTimeMs = uint32(st)
	idx += 1

	// Plantower PMS5003s
	if err := d.Sps.FromSerialString(splitStr[idx]); err != nil {
		return fmt.Errorf("failed to convert PT0 string, %s, to PlantowerData", splitStr[idx])
	}
	idx += 1

	// Temperature
	if temp, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert si temp string, %s, to float32", splitStr[idx])
	} else {
		d.Si.Temp = float32(temp)
	}
	idx += 1

	// Temperature
	if temp, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[idx])
	} else {
		d.Scd.Temp = float32(temp)
	}
	idx += 1

	// Humidity
	if hum, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert si hum string, %s, to float32", splitStr[idx])
	} else {
		d.Si.Hum = float32(hum)
	}

	// Humidity
	if hum, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[idx])
	} else {
		d.Scd.Hum = float32(hum)
	}

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[idx], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[idx])
	} else {
		d.Mprls.Pressure = float32(press)
	}

	// TVOC
	if voc, err := strconv.ParseInt(splitStr[idx], 10, 32); err != nil {
		return fmt.Errorf("failed to convert tvoc string, %s, to int32", splitStr[idx])
	} else {
		d.Sgp30.Tvoc = int32(voc)
	}

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[idx], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[idx])
	} else {
		d.Sgp40.VocIndex = uint32(voc)
	}

	// CO2
	if co2, err := strconv.ParseUint(splitStr[idx], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 index string, %s, to uint16", splitStr[idx])
	} else {
		d.Scd.Co2 = uint16(co2)
	}

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[idx], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[idx])
	} else {
		d.SensorStates = uint8(sensorStates)
	}

	return nil
}

func (d *DuetDataMk1Var3) doPopulateFromBytes(buff []byte) error {
	reader := bytes.NewReader(buff)
	var toss uint8
	pointers := append(
		[]any{&d.SensorStates, &toss, &d.SerialNumber, &d.Scd.Co2, &d.Sgp30.Tvoc, &d.Sgp40.VocIndex, &d.SampleTimeMs,
			&d.Si.Temp, &d.Scd.Temp, &d.Si.Hum, &d.Scd.Hum, &d.Mprls.Pressure},
		[]any(d.Sps.PointerIterable())...,
	)

	for idx := range pointers {
		if err := binary.Read(reader, binary.LittleEndian, pointers[idx]); err != nil {
			return fmt.Errorf("error converting bytes at index %d: %w", idx, err)
		}
	}

	return nil
}
func (d *DuetDataMk1Var3) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     1.3,
		KEY_SERIAL_NUMBER:   d.SerialNumber,
		KEY_DEVICE_ID:       d.SerialNumber,
		KEY_UNIX:            d.UnixSec,
		KEY_ECO2:            0,
		KEY_RAWH2:           0,
		KEY_SENSOR_STATES:   d.SensorStates,
		KEY_CONNECTION_TYPE: d.ConnectionType,
		KEY_LAST_RESET_TIME: d.LastResetUnix,
		KEY_GATEWAY_SERIAL:  gatewaySerial,

		KEY_TVOC:    d.Sgp30.Tvoc,
		"voc_index": d.Sgp40.VocIndex, // TODO
	}
	maps.Copy(ret, d.Sps.ToMap("_t"))
	maps.Copy(ret, d.Sps.ToMap("_b"))
	maps.Copy(ret, d.Sps.ToMap("_m"))
	maps.Copy(ret, d.Si.ToMap())
	maps.Copy(ret, d.Scd.ToMap())
	maps.Copy(ret, d.Mprls.ToMap())
	maps.Copy(ret, d.RadioMeta.ToMap())
	if d.piMcuTempSet {
		ret[KEY_PI_MCU_TEMP] = d.PiMcuTemp
	}

	return ret
}
