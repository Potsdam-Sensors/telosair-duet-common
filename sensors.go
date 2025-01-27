package telosairduetcommon

import (
	"fmt"
	"os"
	"path"
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
