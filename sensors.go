package telosairduetcommon

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type SensorMeasurement interface {
	DirectoryName() string
	DirectoryData() map[string]float32
}

func StoreSensorData(m SensorMeasurement, dir string) error {
	folderPath := path.Join(dir, m.DirectoryName())
	filenamesToValues := m.DirectoryData()

	// Ensure the directory exists, or create it.
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", folderPath, err)
	}

	// Write each measurement into its own file.
	for filename, value := range filenamesToValues {
		content := fmt.Sprintf("%v\n", value)
		filePath := path.Join(folderPath, filename)

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %q: %w", filePath, err)
		}
	}

	return nil
}

// StoreDeviceType writes the device type string to a file at the root of the sensor directory.
func StoreDeviceType(deviceType string, dir string) error {
	// Ensure the base directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	deviceType = strings.TrimPrefix(deviceType, "Mk")
	deviceType = strings.TrimPrefix(deviceType, "mk")

	filePath := path.Join(dir, "device_type")
	content := fmt.Sprintf("%s\n", deviceType)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write device_type file %q: %w", filePath, err)
	}

	return nil
}
