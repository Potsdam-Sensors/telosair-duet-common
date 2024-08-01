package duetcommongo

import (
	"fmt"
)

/* ~~ Keys & Constants ~~ */
const (
	KEY_HTU_TEMP        = "temp_htu"
	KEY_HTU_HUM         = "hum_htu"
	KEY_SCD_TEMP        = "temp_scd"
	KEY_SCD_HUM         = "hum_scd"
	KEY_SCD_CO2         = "co2"
	KEY_SCD_CO2_LEGACY  = "rawethanol"
	KEY_TEMP            = "temp"
	KEY_HUM             = "hum"
	KEY_TVOC            = "tvoc"
	KEY_VOC_INDEX       = "tvoc"
	KEY_MPRLS_PRESSURE  = "pressure"
	KEY_RSSI            = "lastRssi"
	KEY_SNR             = "lastSNR"
	KEY_HOPS            = "hops"
	KEY_DEVICE_TYPE     = "deviceType"
	KEY_SERIAL_NUMBER   = "serial_number"
	KEY_DEVICE_ID       = "device_id"
	KEY_UNIX            = "unix"
	KEY_ECO2            = "eco2"
	KEY_RAWH2           = "rawh2"
	KEY_SENSOR_STATES   = "sensorStates"
	KEY_CONNECTION_TYPE = "connection_type"
	KEY_LAST_RESET_TIME = "lastResetTime"
	KEY_GATEWAY_SERIAL  = "gateway_serial"
	KEY_POE_USB_VOLTAGE = "poe_usb_voltage"

	CONNECTION_TYPE_LORA_GATEWAY = 0
	CONNECTION_TYPE_LORAWAN      = 1 // TODO: is this true? Unused I think now
	CONNECTION_TYPE_USB_SERIAL   = 2
)

func validateConnectionType(t int) error {
	switch t {
	case CONNECTION_TYPE_LORA_GATEWAY, CONNECTION_TYPE_LORAWAN, CONNECTION_TYPE_USB_SERIAL:
		return nil
	default:
		return fmt.Errorf("unable to validate connection type given: %d", t)
	}
}

// /* ~~ Type Checking ~~ */
// // TODO: PT & Gas stuff
// var expectedTypeCheckers = map[string]func(any) bool{
// 	KEY_DEVICE_TYPE:     isFloat,
// 	KEY_SERIAL_NUMBER:   isInteger,
// 	KEY_DEVICE_ID:       isInteger,
// 	KEY_TEMP:            isFloat,
// 	KEY_HUM:             isFloat,
// 	KEY_MPRLS_PRESSURE:  isFloat,
// 	KEY_TVOC:            isInteger,
// 	KEY_SCD_CO2_LEGACY:  isInteger,
// 	KEY_SCD_CO2:         isInteger,
// 	KEY_UNIX:            isInteger,
// 	KEY_ECO2:            isInteger,
// 	KEY_RAWH2:           isInteger,
// 	KEY_RSSI:            isInteger,
// 	KEY_SNR:             isInteger,
// 	KEY_HOPS:            isInteger,
// 	KEY_SENSOR_STATES:   isInteger,
// 	KEY_CONNECTION_TYPE: isInteger,
// 	KEY_LAST_RESET_TIME: isInteger,
// 	KEY_GATEWAY_SERIAL:  isString,
// 	KEY_POE_USB_VOLTAGE: isInteger,
// 	KEY_HTU_TEMP:        isFloat,
// 	KEY_HTU_HUM:         isFloat,
// 	KEY_SCD_TEMP:        isFloat,
// 	KEY_SCD_HUM:         isFloat,
// }

// func isInteger(value interface{}) bool {
// 	v := reflect.ValueOf(value)
// 	switch v.Kind() {
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
// 		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
// 		return true
// 	default:
// 		return false
// 	}
// }

// func isFloat(value interface{}) bool {
// 	v := reflect.ValueOf(value)
// 	switch v.Kind() {
// 	case reflect.Float32, reflect.Float64:
// 		return true
// 	default:
// 		return false
// 	}
// }

// func isString(value interface{}) bool {
// 	v := reflect.ValueOf(value)
// 	switch v.Kind() {
// 	case reflect.String:
// 		return true
// 	default:
// 		return false
// 	}
// }
