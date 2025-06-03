package telosairduetcommon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"maps"
	"strconv"
)

/* ~~ MK1 Var 2 ~~ */
var DuetTypeMk1Var2 = DuetTypeInfo{
	ExpectedBytes:        44,
	ExpectedStringLen:    10,
	StructInstanceGetter: func() DuetData { return &DuetDataMk1Var2{} },
	TypeAlias:            "Mk1.2",
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
type DuetDataMk1Var2 struct {
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
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata
}

func (d *DuetDataMk1Var2) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Sps, d.Si, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}

func (d *DuetDataMk1Var2) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk1Var2) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk1Var2) String() string {
	return fmt.Sprintf("[Duet %d, Type %d.%d | Unix %d | Si7021 %s | MPRLS: %s | SGP: %s | SPS30: %s | Radio: %s | Errstate %d ]",
		d.SerialNumber, 1, 2, d.UnixSec, d.Si.String(), d.Mprls.String(), d.Sgp.String(), d.Sps.String(),
		d.RadioMeta.String(), d.SensorStates)
}
func (d *DuetDataMk1Var2) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk1Var2
}

func (d *DuetDataMk1Var2) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk1Var2) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk1Var2) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk1Var2) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

func (d *DuetDataMk1Var2) doPopulateFromSubStrings(splitStr []string) error {
	// Serial Number
	sn, err := strconv.ParseUint(splitStr[0], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[0])
	}
	d.SerialNumber = uint16(sn)

	// Sample Time
	st, err := strconv.ParseUint(splitStr[1], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[1])
	}
	d.SampleTimeMs = uint32(st)

	// Plantower PMS5003s
	if err := d.Sps.FromSerialString(splitStr[2]); err != nil {
		return fmt.Errorf("failed to convert PT0 string, %s, to PlantowerData", splitStr[2])
	}

	// Temperature
	if temp, err := strconv.ParseFloat(splitStr[3], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[3])
	} else {
		d.Si.Temp = float32(temp)
	}

	// Humidity
	if hum, err := strconv.ParseFloat(splitStr[4], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[4])
	} else {
		d.Si.Hum = float32(hum)
	}

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[5], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[5])
	} else {
		d.Mprls.Pressure = float32(press)
	}

	// TVOC
	if voc, err := strconv.ParseUint(splitStr[6], 10, 32); err != nil {
		return fmt.Errorf("failed to convert tvoc string, %s, to uint32", splitStr[6])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[7], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[7])
	} else {
		d.SensorStates = uint8(sensorStates)
	}

	return nil
}

func (d *DuetDataMk1Var2) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	// Skip buff[1] as it is padding
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[4:8])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[8:12])

	reader := bytes.NewReader(buff[12:24])
	if err := binary.Read(reader, binary.LittleEndian, &d.Si.Temp); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Si.Hum); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &d.Mprls.Pressure); err != nil {
		return fmt.Errorf("error converting bytes to float: %w", err)
	}

	if err := d.Sps.PopulateFromBytes(buff[24:42]); err != nil {
		return fmt.Errorf("error parsing bytes for PT: %w", err)
	}

	return nil
}
func (d *DuetDataMk1Var2) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     1.2,
		KEY_SERIAL_NUMBER:   d.SerialNumber,
		KEY_DEVICE_ID:       d.SerialNumber,
		KEY_UNIX:            d.UnixSec,
		KEY_ECO2:            0,
		KEY_RAWH2:           0,
		KEY_SENSOR_STATES:   d.SensorStates,
		KEY_CONNECTION_TYPE: d.ConnectionType,
		KEY_LAST_RESET_TIME: d.LastResetUnix,
		KEY_GATEWAY_SERIAL:  gatewaySerial,
	}
	maps.Copy(ret, d.Sps.ToMap("_t"))
	maps.Copy(ret, d.Sps.ToMap("_b"))
	maps.Copy(ret, d.Sps.ToMap("_m"))
	maps.Copy(ret, d.Si.ToMap())
	maps.Copy(ret, d.Mprls.ToMap())
	maps.Copy(ret, d.Sgp.ToMap())
	maps.Copy(ret, d.RadioMeta.ToMap())
	if d.piMcuTempSet {
		ret[KEY_PI_MCU_TEMP] = d.PiMcuTemp
	}

	return ret
}
