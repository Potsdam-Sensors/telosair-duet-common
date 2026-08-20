package telosairduetcommon

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const Mk4Var24BinBoundsPath = "/var/run/sensor_data/opc_bin_bounds.json"

/* ~~ MK4 Var 24 - One OPC-N3 ~~ */
var DuetTypeMk4Var24 = DuetTypeInfo{
	Major:                4,
	Variant:              24,
	ExpectedBytes:        152,
	ExpectedStringLen:    19,
	StructInstanceGetter: func() DuetData { return &DuetDataMk4Var24{} },
	TypeAlias:            "Mk4.24",
	RunVariantConfig: func(ctx VariantInitContext) error {
		fmt.Printf("[%s] Running variant-specific configuration...\n", "Mk4.24")

		bounds, err := ReadBinBoundaries(ctx.Writer, ctx.Scanner, "B", ctx.Timeout, ctx.IsDebug)
		if err != nil {
			return fmt.Errorf("failed to read bin boundaries: %w", err)
		}

		// Use the variant's own hardcoded path!
		if err := SaveBinBoundariesToFile(Mk4Var24BinBoundsPath, bounds); err != nil {
			fmt.Printf("Error saving bin boundaries: %v\n", err)
		} else {
			fmt.Printf("Successfully wrote bin boundaries to %s\n", Mk4Var24BinBoundsPath)
		}

		if ctx.UploadBootMetadata != nil {
			if err := ctx.UploadBootMetadata(ctx.SerialNumber, bounds); err != nil {
				fmt.Printf("Error sending boot metadata to Kafka: %v\n", err)
			}
		}

		return nil
	},
}

// Add this right below your DuetTypeMk4Var24 definition
var SupportedVariants = []DuetTypeInfo{
	DuetTypeMk4Var24,
	// You can add any future variants here (e.g. Mk5.10)
}

type DuetDataMk4Var24 struct {
	SerialNumber   uint16
	SampleTimeMs   uint32
	UnixSec        uint32
	LastResetUnix  uint32
	SensorStates   uint8
	PoeUsbVoltage  uint8
	ConnectionType int
	PiMcuTemp      float32
	piMcuTempSet   bool

	Opc       AlphasenseOpcN3Measurement
	Scd       Scd41Measurement
	Htu       Htu21Measurement
	TempRh    CombinedTempRhMeasurements
	Mprls     MprlsMeasurement
	Sgp       Sgp40Measurement
	RadioMeta RadioMetadata

	timeResolved bool
}

func (d *DuetDataMk4Var24) TimeResolved() bool {
	return d.timeResolved
}
func (d *DuetDataMk4Var24) MarkTimeResolved(v bool) {
	d.timeResolved = v
}
func (d *DuetDataMk4Var24) Timestamp() uint32 {
	return d.UnixSec
}
func (d *DuetDataMk4Var24) ResolveTime(t uint32) {
	d.UnixSec = t
}

func (d *DuetDataMk4Var24) SensorMeasurements() []SensorMeasurement {
	return []SensorMeasurement{d.Opc, d.TempRh, d.Scd, d.Mprls, d.Sgp, DuetSensorState{d.SensorStates}}
}
func (d *DuetDataMk4Var24) SetRadioData(v RadioMetadata) {
	d.RadioMeta = v
}
func (d *DuetDataMk4Var24) SetPiMcuTemp(val float32) {
	d.PiMcuTemp = val
	d.piMcuTempSet = true
}
func (d *DuetDataMk4Var24) String() string {
	return fmt.Sprintf("[Duet %d, Type 4.24 | Unix %d | %s | HTU: %s | SCD: %s | MPRLS: %s | SGP: %s | OPC: %s | Radio: %s | Errstate %d | PoE Voltage %d]",
		d.SerialNumber, d.UnixSec, d.TempRh.String(), d.Htu.String(), d.Scd.String(), d.Mprls.String(), d.Sgp.String(), d.Opc.String(),
		d.RadioMeta.String(), d.SensorStates, d.PoeUsbVoltage)
}
func (d *DuetDataMk4Var24) GetTypeInfo() DuetTypeInfo {
	return DuetTypeMk4Var24
}

func (d *DuetDataMk4Var24) SetConnectionType(ct int) {
	d.ConnectionType = ct
}
func (d *DuetDataMk4Var24) SetTimeRadio(unixSecRecieved uint32) error {
	if (d.RadioMeta.RadioSentTimeMs < d.SampleTimeMs) || (unixSecRecieved*d.RadioMeta.RadioSentTimeMs*unixSecRecieved == 0) {
		return fmt.Errorf("incompatible timekeeping parameters: unix: %d, radio sent ms: %d, sample ms: %d", unixSecRecieved, d.RadioMeta.RadioSentTimeMs, d.SampleTimeMs)
	}
	d.UnixSec = unixSecRecieved - ((d.RadioMeta.RadioSentTimeMs - d.SampleTimeMs) / 1000)
	return nil
}

func (d *DuetDataMk4Var24) SetTimeSerial(unixSecRecieved uint32) {
	d.UnixSec = unixSecRecieved
}

func (d *DuetDataMk4Var24) RecalculateLastResetUnix() {
	d.LastResetUnix = d.UnixSec - (d.SampleTimeMs / 1000)
}

/*
Error converting data point, 4 24 4404 1831248 27.06 24.52 30.38 28.62 33.65 25.89 404.82 100 426 0.27 0.51 0.59 39 0
: failed to populate for type Mk4.24: failed to convert voc index string, 25.89, to uint32
*/
func (d *DuetDataMk4Var24) doPopulateFromSubStrings(splitStr []string) error {
	cur := 0
	// Serial Number
	sn, err := strconv.ParseUint(splitStr[cur], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to convert DuetSerialNumber string, %s, to uint32", splitStr[cur])
	}
	d.SerialNumber = uint16(sn)
	cur++

	// Sample Time
	st, err := strconv.ParseUint(splitStr[cur], 10, 32)
	if err != nil {
		return fmt.Errorf("failed to convert SampleTime string, %s, to uint32", splitStr[cur])
	}
	d.SampleTimeMs = uint32(st)
	cur++

	// Temperatures (1 & 2 & 3)
	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert htu temp string, %s, to float32", splitStr[cur])
	} else {
		d.Htu.Temp = float32(temp)
	}
	cur++

	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert scd temp string, %s, to float32", splitStr[cur])
	} else {
		d.Scd.Temp = float32(temp)
	}
	cur++

	if temp, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc temp string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.Temp = float32(temp)
	}
	cur++

	// Humidities (1 & 2 & 3)
	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert htu hum string, %s, to float32", splitStr[cur])
	} else {
		d.Htu.Hum = float32(hum)
	}
	cur++

	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert scd hum string, %s, to float32", splitStr[cur])
	} else {
		d.Scd.Hum = float32(hum)
	}
	cur++

	if hum, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc hum string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.Rh = float32(hum)
	}
	cur++

	// Pressure
	if press, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert pressure string, %s, to float32", splitStr[cur])
	} else {
		d.Mprls.Pressure = float32(press)
	}
	cur++

	// VOC Index
	if voc, err := strconv.ParseUint(splitStr[cur], 10, 32); err != nil {
		return fmt.Errorf("failed to convert voc index string, %s, to uint32", splitStr[cur])
	} else {
		d.Sgp.VocIndex = uint32(voc)
	}
	cur++

	// CO2
	if co2, err := strconv.ParseUint(splitStr[cur], 10, 16); err != nil {
		return fmt.Errorf("failed to convert co2 string, %s, to uint32", splitStr[cur])
	} else {
		d.Scd.Co2 = uint16(co2)
	}
	cur++

	// OPC-N3 PM1, 2.5, 10 & Bins
	if pm1, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc pm1 string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.PM1 = float32(pm1)
	}
	cur++

	if pm2p5, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc pm2.5 string, %s, to float32", splitStr[cur])
	} else {
		d.Opc.PM2p5 = float32(pm2p5)
	}
	cur++

	if pm10, err := strconv.ParseFloat(splitStr[cur], 32); err != nil {
		return fmt.Errorf("failed to convert opc pm10 string, %s, to float32", splitStr[cur])
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
		return fmt.Errorf("failed to convert voltage string, %s, to uint8", splitStr[cur])
	} else {
		d.PoeUsbVoltage = uint8(voltage)
	}
	cur++

	// Sensor States
	if sensorStates, err := strconv.ParseUint(splitStr[cur], 10, 8); err != nil {
		return fmt.Errorf("failed to convert states string, %s, to uint8", splitStr[cur])
	} else {
		d.SensorStates = uint8(sensorStates)
	}
	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)
	cur++

	return nil
}
func (d *DuetDataMk4Var24) doPopulateFromBytes(buff []byte) error {
	d.SensorStates = buff[0]
	d.PoeUsbVoltage = buff[1]
	d.SerialNumber = binary.LittleEndian.Uint16(buff[2:4])
	d.Scd.Co2 = binary.LittleEndian.Uint16(buff[4:6])
	d.Sgp.VocIndex = binary.LittleEndian.Uint32(buff[6:10])
	d.SampleTimeMs = binary.LittleEndian.Uint32(buff[10:14])

	reader := bytes.NewReader(buff[14:34])
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

	CombineTempRhMeasurements(d.Htu, d.Scd, &d.TempRh)

	return nil
}
func (d *DuetDataMk4Var24) ToMap(gatewaySerial string) map[string]any {
	ret := map[string]any{
		KEY_DEVICE_TYPE:     4.24,
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
	maps.Copy(ret, d.Opc.ToMapPm("_t"))
	maps.Copy(ret, d.Opc.ToMapPm("_b"))
	maps.Copy(ret, d.Opc.ToMapPm("_m"))
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
func ReadBinBoundaries(
	writer io.Writer,
	scanner *bufio.Scanner,
	command string,
	timeout time.Duration,
	isDebug bool,
) ([]float32, error) {

	if isDebug {
		return []float32{
			0.35, 0.46, 0.66, 1.0, 1.3, 1.7,
			2.3, 3.0, 4.0, 5.2, 6.5, 8.0, 10.0,
		}, nil
	}

	if _, err := writer.Write([]byte(command + "\n")); err != nil {
		return nil, fmt.Errorf("failed to write query command: %w", err)
	}

	fmt.Printf("[BinBounds] Sending command: %q\n", command+"\n")

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("[BinBounds] Scanner error: %v\n", err)
			} else {
				fmt.Printf("[BinBounds] Scanner returned false with no error\n")
			}

			// The scanner can't recover itself after Scan() returns false.
			// We need to reconstruct it from the underlying reader.
			reader, ok := writer.(io.Reader)
			if !ok {
				return nil, fmt.Errorf(
					"serial writer does not implement io.Reader",
				)
			}

			*scanner = *bufio.NewScanner(reader)

			time.Sleep(50 * time.Millisecond)
			continue
		}

		line := strings.TrimSpace(scanner.Text())

		fmt.Printf("[BinBounds] Received: %q\n", line)

		if !strings.Contains(line, "DIAMETERS:") {
			continue
		}

		parts := strings.Split(line, ";")

		var diametersStr string

		for _, part := range parts {
			if strings.HasPrefix(part, "DIAMETERS:") {
				diametersStr = strings.TrimPrefix(part, "DIAMETERS:")
				break
			}
		}

		if diametersStr == "" {
			return nil, fmt.Errorf(
				"received DIAMETERS line but payload was empty: %s",
				line,
			)
		}

		rawBounds := strings.Split(diametersStr, ",")
		bounds := make([]float32, 0, len(rawBounds))

		for _, raw := range rawBounds {
			val, err := strconv.ParseFloat(strings.TrimSpace(raw), 32)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid bin value %q: %w",
					raw,
					err,
				)
			}

			bounds = append(bounds, float32(val)/100.0)
		}

		return bounds, nil
	}

	return nil, fmt.Errorf(
		"timed out waiting for OPC bin boundaries after %v",
		timeout,
	)
}

// // Notice the updated signature: we drop portPath and accept writer & scanner
// func ReadBinBoundaries(writer io.Writer, scanner *bufio.Scanner, command string, timeout time.Duration, isDebug bool) ([]float32, error) {
// 	if isDebug {
// 		return []float32{0.35, 0.46, 0.66, 1.0, 1.3, 1.7, 2.3, 3.0, 4.0, 5.2, 6.5, 8.0, 10.0}, nil
// 	}

// 	type response struct {
// 		data []float32
// 		err  error
// 	}
// 	ch := make(chan response, 1)

// 	go func() {
// 		cmd := []byte(command + "\n")
// 		fmt.Printf("[BinBounds] Sending command: %q\n", cmd)

// 		// 1. Send query command to daughterboard using the shared writer
// 		_, err := writer.Write(cmd)
// 		if err != nil {
// 			ch <- response{err: fmt.Errorf("failed to write query command: %w", err)}
// 			return
// 		}

// 		// 2. Scan lines continuously using the SHARED scanner
// 		// (We do not call bufio.NewScanner(f) here anymore!)
// 		for scanner.Scan() {
// 			line := strings.TrimSpace(scanner.Text())

// 			// Skip boot splash lines (e.g. "init ok!")
// 			if !strings.HasPrefix(line, "B;CONF") && !strings.Contains(line, "DIAMETERS:") {
// 				continue
// 			}

// 			parts := strings.Split(line, ";")
// 			var diametersStr string
// 			for _, part := range parts {
// 				if strings.HasPrefix(part, "DIAMETERS:") {
// 					diametersStr = strings.TrimPrefix(part, "DIAMETERS:")
// 					break
// 				}
// 			}

// 			if diametersStr == "" {
// 				ch <- response{err: fmt.Errorf("received B;CONF line but could not locate DIAMETERS payload: %s", line)}
// 				return
// 			}

// 			// Split the extracted DIAMETERS substring: "35,46,66,100..."
// 			rawBounds := strings.Split(diametersStr, ",")
// 			bounds := make([]float32, 0, len(rawBounds))

// 			for _, raw := range rawBounds {
// 				val, err := strconv.ParseFloat(strings.TrimSpace(raw), 32)
// 				if err != nil {
// 					ch <- response{err: fmt.Errorf("invalid bin value '%s': %w", raw, err)}
// 					return
// 				}
// 				// Convert integer diameters to micrometers (e.g., 35 -> 0.35 um)
// 				bounds = append(bounds, float32(val)/100.0)
// 			}

// 			ch <- response{data: bounds}
// 			return
// 		}

// 		if err := scanner.Err(); err != nil {
// 			ch <- response{err: err}
// 			return
// 		}

// 		ch <- response{err: fmt.Errorf("EOF reached without finding bin boundaries")}
// 	}()

// 	select {
// 	case res := <-ch:
// 		return res.data, res.err
// 	case <-time.After(timeout):
// 		return nil, fmt.Errorf("timed out waiting for bin boundaries after %v", timeout)
// 	}
// }

// SaveBinBoundariesToFile writes the parsed bin boundaries to a local JSON file.
func SaveBinBoundariesToFile(filePath string, bounds []float32) error {
	data, err := json.MarshalIndent(bounds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bin boundaries: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write bin boundaries to %s: %w", filePath, err)
	}

	return nil
}
