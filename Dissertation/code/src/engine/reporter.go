package engine

import (
	"encoding/json"
	"os"
)

// GenerateReport записывает отчет о прогоне стратегии
// в файл формата json
func GenerateReport(results any) error {
	file, err := os.Create("report.json")
	if err != nil {
		return err
	}
	defer file.Close()

	// Используем json.Encoder для записи в файл
	encoder := json.NewEncoder(file)
	// Форматирование с отступами (опционально)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(results)
	if err != nil {
		return err
	}

	return nil
}
