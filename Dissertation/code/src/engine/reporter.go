package engine

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func GenerateReport(results []SimulationResult) {
	file, _ := os.Create("report.csv")
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Strategy", "Tasks Done", "Energy Used", "Avg Delay"})

	for _, res := range results {
		writer.Write([]string{
			res.StrategyName,
			strconv.Itoa(res.TotalTasksDone),
			fmt.Sprintf("%.2f", res.TotalEnergyUsed),
			fmt.Sprintf("%.2f", res.AvgTaskDelay),
		})
	}
}
