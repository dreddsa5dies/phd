package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/strategies"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type (
	// ключ - имя стратегии
	ScenarioData map[string]strategies.Metrics

	// ключ - сценарий: "equal", "max", "min"
	AllData map[string]ScenarioData
)

var (
	bigFontSize   = 22
	largeFontSize = 18
	meanFontSize  = 16
)

// Парсинг времени в микросекунды (безопасно с UTF-8)
func parseTime(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	var unit string
	var valueStr string

	// Используем TrimSuffix для безопасного удаления суффиксов
	switch {
	case strings.HasSuffix(s, "ns"):
		unit = "ns"
		valueStr = strings.TrimSpace(strings.TrimSuffix(s, "ns"))
	case strings.HasSuffix(s, "µs"): // Unicode-символ µ (U+00B5)
		unit = "µs"
		valueStr = strings.TrimSpace(strings.TrimSuffix(s, "µs"))
	case strings.HasSuffix(s, "us"): // fallback: если вдруг пришло как "us"
		unit = "µs"
		valueStr = strings.TrimSpace(strings.TrimSuffix(s, "us"))
	case strings.HasSuffix(s, "ms"):
		unit = "ms"
		valueStr = strings.TrimSpace(strings.TrimSuffix(s, "ms"))
	case strings.HasSuffix(s, "s"):
		unit = "s"
		valueStr = strings.TrimSpace(strings.TrimSuffix(s, "s"))
	default:
		unit = "µs"
		valueStr = s
	}

	if valueStr == "" {
		return 0
	}

	f, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0
	}

	// Конвертируем всё в микросекунды (µs)
	switch unit {
	case "ns":
		return f / 1000.0 // нано - микро
	case "µs", "us":
		return f
	case "ms":
		return f * 1000.0 // милли - микро
	case "s":
		return f * 1_000_000.0 // секунды - микро
	default:
		return f
	}
}

// Получение количества выполненных задач на шаге
func getDoneCount(metrics []strategies.StepMetrics, step int) int {
	if step >= len(metrics) || step < 0 {
		return 0
	}
	m := metrics[step]
	if len(m.TasksDone) > 0 {
		return len(m.TasksDone)
	}
	return m.LenTasksDone
}

// Медиана
func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	n := len(values)
	if n%2 == 1 {
		return values[n/2]
	}
	return (values[n/2-1] + values[n/2]) / 2
}

// Сбор агрегированных результатов
type ResultRow struct {
	Scenario         string
	Strategy         string
	DoneTasks        int
	EnergyUsed       float64
	TotalEnergy      float64
	TotalTimeUS      float64
	PercentDone      float64
	AvgEnergyPerTask float64
}

func collectAllResults(data AllData) []ResultRow {
	var results []ResultRow
	scenarios := []string{"equal", "max", "min"}
	for _, scenario := range scenarios {
		scenarioData, ok := data[scenario]
		if !ok {
			continue
		}
		for strategy, metrics := range scenarioData {
			done := getDoneCount(metrics.StepMetric, len(metrics.StepMetric)-1)
			totalTime := parseTime(metrics.TotalTime)
			energyUsed := 0.0
			if len(metrics.StepMetric) > 0 {
				energyUsed = metrics.StepMetric[len(metrics.StepMetric)-1].EnergyUsed
			}
			percentDone := 0.0
			if metrics.TotalTasks > 0 {
				percentDone = float64(done) / float64(metrics.TotalTasks) * 100
			}
			avgEnergyPerTask := 0.0
			if done > 0 {
				avgEnergyPerTask = energyUsed / float64(done)
			}
			results = append(results, ResultRow{
				Scenario:         scenario,
				Strategy:         strategy,
				DoneTasks:        done,
				EnergyUsed:       energyUsed,
				TotalEnergy:      metrics.TotalEnergyMachines,
				TotalTimeUS:      totalTime,
				PercentDone:      percentDone,
				AvgEnergyPerTask: avgEnergyPerTask,
			})
		}
	}
	return results
}

// Сохранение CSV
func saveSummaryCSV(results []ResultRow) {
	f, err := os.Create("summary_results.csv")
	if err != nil {
		log.Fatalf("Ошибка создания CSV: %v", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	headers := []string{"Сценарий", "Стратегия", "Выполнено задач", "Использовано энергии", "Всего энергии", "% выполнения", "Ср. энергия на задачу", "Общее время (µs)"}
	if err := w.Write(headers); err != nil {
		log.Fatalf("Ошибка записи заголовка: %v", err)
	}

	for _, r := range results {
		record := []string{
			r.Scenario,
			r.Strategy,
			strconv.Itoa(r.DoneTasks),
			fmt.Sprintf("%.2f", r.EnergyUsed),
			fmt.Sprintf("%.2f", r.TotalEnergy),
			fmt.Sprintf("%.1f%%", r.PercentDone),
			fmt.Sprintf("%.2f", r.AvgEnergyPerTask),
			fmt.Sprintf("%.3f", r.TotalTimeUS),
		}
		if err := w.Write(record); err != nil {
			log.Printf("Ошибка записи строки: %v", err)
		}
	}

	fmt.Println("CSV сохранён: summary_results.csv")
}

// Группированный bar-чарт по полю
func plotGroupedBarChart(results []ResultRow, field, _, ylabel string, getValue func(ResultRow) float64) {
	scenarios := []string{"equal", "max", "min"}
	strategiesMap := make(map[string]bool)
	for _, r := range results {
		strategiesMap[r.Strategy] = true
	}
	var strategies []string
	for s := range strategiesMap {
		strategies = append(strategies, s)
	}
	sort.Strings(strategies)

	data := make(map[string]map[string]float64)
	for _, sc := range scenarios {
		data[sc] = make(map[string]float64)
	}
	for _, r := range results {
		data[r.Scenario][r.Strategy] = getValue(r)
	}

	p := plot.New()

	// Настройка увеличенных шрифтов
	p.Y.Label.Text = ylabel
	p.Y.Label.TextStyle.Font.Size = font.Length(largeFontSize)
	p.X.Tick.Label.Rotation = 0.5
	p.X.Tick.Label.Font.Size = font.Length(meanFontSize)
	p.Y.Tick.Label.Font.Size = font.Length(meanFontSize)
	p.X.Padding = 30

	// Увеличенная ширина столбцов
	width := vg.Points(40)
	// Уменьшенный отступ между группами
	sep := vg.Points(10)

	barCharts := make([]plot.Plotter, 0)

	for i, strategy := range strategies {
		values := make(plotter.Values, len(scenarios))
		for j, sc := range scenarios {
			values[j] = data[sc][strategy]
		}
		bars, err := plotter.NewBarChart(values, width)
		if err != nil {
			continue
		}
		color := plotutil.Color(i)
		bars.Color = color
		// Плотное размещение: меньше смещение
		bars.Offset = vg.Length(i)*(width+vg.Points(5)) - sep/2
		barCharts = append(barCharts, bars)

		// Добавляем в легенду с цветом
		p.Legend.Add(strategy, bars)
	}

	for _, bc := range barCharts {
		p.Add(bc)
	}
	p.NominalX("equal", "max", "min")
	p.Legend.Top = true
	p.Legend.TextStyle.Font.Size = font.Length(bigFontSize)

	if err := p.Save(1000, 600, "bar_"+strings.ReplaceAll(field, " ", "_")+".png"); err != nil {
		log.Printf("Ошибка сохранения bar_%s.png: %v", field, err)
	} else {
		fmt.Printf("График сохранён: bar_%s.png\n", strings.ReplaceAll(field, " ", "_"))
	}
}

// Bar-чарты: среднее и медиана выполненных задач по шагам
func plotMeanMedianBarCharts(data AllData) {
	scenarios := []string{"equal", "max", "min"}
	strategiesMap := make(map[string]bool)

	type Stats struct {
		Mean   float64
		Median float64
	}
	stats := make(map[string]map[string]Stats)

	for _, scenario := range scenarios {
		scenarioData, ok := data[scenario]
		if !ok {
			continue
		}
		stats[scenario] = make(map[string]Stats)
		for strategy, metrics := range scenarioData {
			var values []float64
			for step := 0; step < len(metrics.StepMetric); step++ {
				done := float64(getDoneCount(metrics.StepMetric, step))
				values = append(values, done)
			}
			if len(values) == 0 {
				values = []float64{0}
			}

			var sum float64
			for _, v := range values {
				sum += v
			}
			mean := sum / float64(len(values))
			medianVal := median(values)

			strategiesMap[strategy] = true
			stats[scenario][strategy] = Stats{Mean: mean, Median: medianVal}
		}
	}

	var strategies []string
	for s := range strategiesMap {
		strategies = append(strategies, s)
	}
	sort.Strings(strategies)

	width := vg.Points(40)
	sep := vg.Points(10)

	// Среднее
	{
		p := plot.New()

		// Настройка увеличенных шрифтов
		p.Y.Label.Text = "Среднее (по шагам)"
		p.Y.Label.TextStyle.Font.Size = font.Length(largeFontSize)
		p.X.Tick.Label.Rotation = 0.5
		p.X.Tick.Label.Font.Size = font.Length(meanFontSize)
		p.Y.Tick.Label.Font.Size = font.Length(meanFontSize)
		p.X.Padding = 30

		barCharts := make([]plot.Plotter, 0)

		for i, strategy := range strategies {
			values := make(plotter.Values, len(scenarios))
			for j, sc := range scenarios {
				values[j] = stats[sc][strategy].Mean
			}
			bars, err := plotter.NewBarChart(values, width)
			if err != nil {
				continue
			}
			color := plotutil.Color(i)
			bars.Color = color
			bars.Offset = vg.Length(i)*(width+vg.Points(5)) - sep/2
			barCharts = append(barCharts, bars)

			p.Legend.Add(strategy, bars)
		}

		for _, bc := range barCharts {
			p.Add(bc)
		}
		p.NominalX("equal", "max", "min")
		p.Legend.Top = true
		p.Legend.TextStyle.Font.Size = font.Length(bigFontSize)

		if err := p.Save(1000, 600, "bar_done_tasks_mean.png"); err != nil {
			log.Printf("Ошибка сохранения bar_done_tasks_mean.png: %v", err)
		} else {
			fmt.Println("График сохранён: bar_done_tasks_mean.png")
		}
	}

	// Медиана
	{
		p := plot.New()

		// Настройка увеличенных шрифтов
		p.Y.Label.Text = "Медиана (по шагам)"
		p.Y.Label.TextStyle.Font.Size = font.Length(largeFontSize)
		p.X.Tick.Label.Rotation = 0.5
		p.X.Tick.Label.Font.Size = font.Length(meanFontSize)
		p.Y.Tick.Label.Font.Size = font.Length(meanFontSize)
		p.X.Padding = 30

		barCharts := make([]plot.Plotter, 0)

		for i, strategy := range strategies {
			values := make(plotter.Values, len(scenarios))
			for j, sc := range scenarios {
				values[j] = stats[sc][strategy].Median
			}
			bars, err := plotter.NewBarChart(values, width)
			if err != nil {
				continue
			}
			color := plotutil.Color(i)
			bars.Color = color
			bars.Offset = vg.Length(i)*(width+vg.Points(5)) - sep/2
			barCharts = append(barCharts, bars)

			p.Legend.Add(strategy, bars)
		}

		for _, bc := range barCharts {
			p.Add(bc)
		}
		p.NominalX("equal", "max", "min")
		p.Legend.Top = true
		p.Legend.TextStyle.Font.Size = font.Length(bigFontSize)

		if err := p.Save(1000, 600, "bar_done_tasks_median.png"); err != nil {
			log.Printf("Ошибка сохранения bar_done_tasks_median.png: %v", err)
		} else {
			fmt.Println("График сохранён: bar_done_tasks_median.png")
		}
	}
}

// Основная функция
func main() {
	// Чтение файла
	b, err := os.ReadFile("report.json")
	if err != nil {
		log.Fatalf("Ошибка чтения report.json: %v", err)
	}

	var all AllData
	if err := json.Unmarshal(b, &all); err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	// Сбор данных
	results := collectAllResults(all)

	// Сохранение CSV
	saveSummaryCSV(results)

	// Группированные bar-чарты
	plotGroupedBarChart(results, "Done Tasks", "Выполненные задачи", "Количество", func(r ResultRow) float64 {
		return float64(r.DoneTasks)
	})

	plotGroupedBarChart(results, "Energy Used", "Использованная энергия", "Энергия", func(r ResultRow) float64 {
		return r.EnergyUsed
	})

	plotGroupedBarChart(results, "Total Time", "Общее время выполнения", "Время (µs)", func(r ResultRow) float64 {
		return r.TotalTimeUS
	})

	// Среднее и медиана по шагам
	plotMeanMedianBarCharts(all)

	fmt.Println("Анализ завершён. Файлы: summary_results.csv, bar_*.png")
}
