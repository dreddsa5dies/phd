package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/dreddsa5dies/phd/Dissertation/code/src/models"
	"github.com/dreddsa5dies/phd/Dissertation/code/src/strategies"
)

// Структуры для результатов
type (
	scenarioData map[string]strategies.Metrics // стратегия -> метрики
	allData      map[string]scenarioData       // сценарий -> scenarioData
)

// Результаты анализа
type giniAnalysisResult struct {
	Scenario string
	Strategy string
	Step     int

	// Векторы распределения
	TaskDistribution   []int     // n_i - количество задач на машине
	EnergyDistribution []float64 // ε_i - энергозатраты на машине

	// Коэффициенты Джини для шага
	GiniTasks  float64 // по задачам
	GiniEnergy float64 // по энергии
}

type summaryResult struct {
	Scenario string // наименование сценария
	Strategy string // наименование стратегии

	// Статистики по шагам
	GiniTasksValues  []float64 // по задачам
	GiniEnergyValues []float64 // по энергии

	// Сводные статистики
	AvgGiniTasks  float64
	MedGiniTasks  float64
	AvgGiniEnergy float64
	MedGiniEnergy float64
}

// Функция вычисления коэффициента Джини
func calculateGini(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}

	// Сумма всех элементов
	sumX := 0.0
	for _, v := range values {
		sumX += v
	}

	// Если сумма нулевая, коэффициент Джини = 0
	// (равномерное распределение)
	if sumX == 0 {
		return 0
	}

	// Вычисление суммы модулей разностей
	sumAbsDiff := 0.0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			sumAbsDiff += math.Abs(values[i] - values[j])
		}
	}

	// Формула коэффициента Джини
	gini := sumAbsDiff / (2 * float64(n) * sumX)

	// Ограничиваем значения [0, 1]
	if gini < 0 {
		return 0
	}
	if gini > 1 {
		return 1
	}
	return gini
}

// Анализ одного шага
func analyzeStep(scenario, strategy string, stepMetrics strategies.StepMetrics, allMachines []models.Machine) giniAnalysisResult {
	result := giniAnalysisResult{
		Scenario: scenario,
		Strategy: strategy,
		Step:     stepMetrics.Step,
	}

	machineCount := len(allMachines)
	result.TaskDistribution = make([]int, machineCount)
	result.EnergyDistribution = make([]float64, machineCount)

	machineIndex := make(map[string]int, machineCount)
	for i, machine := range allMachines {
		machineIndex[machine.ID] = i
	}

	// MachineStats: map[machineID]map[taskID]energy
	for machineID, taskEnergyMap := range stepMetrics.MachineStats {
		idx, exists := machineIndex[machineID]
		if !exists {
			log.Printf("Предупреждение: машина %s из MachineStats отсутствует в AllMachines", machineID)
			continue
		}

		// Количество задач = число записей в мапе задач
		result.TaskDistribution[idx] = len(taskEnergyMap)

		// Суммарная энергия = сумма всех значений в мапе
		totalEnergy := 0.0
		for _, energy := range taskEnergyMap {
			totalEnergy += energy
		}
		result.EnergyDistribution[idx] = totalEnergy
	}

	// Преобразуем в []float64 для calculateGini
	taskValues := make([]float64, machineCount)
	for i, count := range result.TaskDistribution {
		taskValues[i] = float64(count)
	}
	energyValues := make([]float64, machineCount)
	copy(energyValues, result.EnergyDistribution)

	result.GiniTasks = calculateGini(taskValues)
	result.GiniEnergy = calculateGini(energyValues)

	return result
}

// Анализ всей стратегии
func analyzeStrategy(scenario, strategy string, metrics strategies.Metrics) summaryResult {
	result := summaryResult{
		Scenario: scenario,
		Strategy: strategy,
	}

	if flag.Lookup("verbose").Value.String() == "true" {
		totalTasks := 0
		for _, step := range metrics.StepMetric {
			totalTasks += step.LenTasksDone
		}
		log.Printf("Сценарий %s, Стратегия %s: всего машин = %d, всего задач = %d", scenario, strategy, len(metrics.AllMachines), totalTasks)
	}

	// Анализируем каждый шаг
	for _, stepMetrics := range metrics.StepMetric {
		stepResult := analyzeStep(scenario, strategy, stepMetrics, metrics.AllMachines)
		result.GiniTasksValues = append(result.GiniTasksValues, stepResult.GiniTasks)
		result.GiniEnergyValues = append(result.GiniEnergyValues, stepResult.GiniEnergy)
	}

	// Вычисляем статистики
	if len(result.GiniTasksValues) > 0 {
		result.AvgGiniTasks = mean(result.GiniTasksValues)
		result.MedGiniTasks = median(result.GiniTasksValues)
	}

	if len(result.GiniEnergyValues) > 0 {
		result.AvgGiniEnergy = mean(result.GiniEnergyValues)
		result.MedGiniEnergy = median(result.GiniEnergyValues)
	}

	return result
}

// Сбор всех результатов
func collectAllResults(allData allData) []summaryResult {
	var results []summaryResult

	for scenario, scenarioData := range allData {
		for strategy, metrics := range scenarioData {
			result := analyzeStrategy(scenario, strategy, metrics)
			results = append(results, result)

			// Детальная информация для отладки (опционально)
			log.Printf("Анализ завершен: Сценарий=%s, Стратегия=%s", scenario, strategy)
			log.Printf("  Коэффициент Джини по задачам: среднее=%.4f, медиана=%.4f",
				result.AvgGiniTasks, result.MedGiniTasks)
			log.Printf("  Коэффициент Джини по энергии: среднее=%.4f, медиана=%.4f",
				result.AvgGiniEnergy, result.MedGiniEnergy)
		}
	}

	return results
}

// Сохранение в CSV
func saveSummaryCSV(results []summaryResult) {
	file, err := os.Create("gini_results.csv")
	if err != nil {
		log.Fatalf("Ошибка создания CSV файла: %v", err)
	}
	defer file.Close()

	// Заголовок CSV
	header := "Сценарий;Стратегия;Среднее значение (задачи);Медиана (задачи);Среднее значение (энергия);Медиана (энергия)\n"
	file.WriteString(header)

	// Данные
	for _, result := range results {
		line := fmt.Sprintf("%s;%s;%.6f;%.6f;%.6f;%.6f\n",
			result.Scenario,
			result.Strategy,
			result.AvgGiniTasks,
			result.MedGiniTasks,
			result.AvgGiniEnergy,
			result.MedGiniEnergy,
		)
		file.WriteString(line)
	}

	log.Println("Результаты сохранены в gini_results.csv")
}

// Дополнительная функция для детального анализа (по шагам)
func saveDetailedAnalysisCSV(allData allData) {
	file, err := os.Create("gini_detailed.csv")
	if err != nil {
		log.Fatalf("Ошибка создания детального CSV файла: %v", err)
	}
	defer file.Close()

	header := "Сценарий;Стратегия;Шаг;Джини (задачи);Джини (энергия)\n"
	file.WriteString(header)

	for scenario, scenarioData := range allData {
		for strategy, metrics := range scenarioData {
			for _, stepMetrics := range metrics.StepMetric {
				stepResult := analyzeStep(scenario, strategy, stepMetrics, metrics.AllMachines)

				line := fmt.Sprintf("%s;%s;%d;%.6f;%.6f\n",
					stepResult.Scenario,
					stepResult.Strategy,
					stepResult.Step,
					stepResult.GiniTasks,
					stepResult.GiniEnergy,
				)
				file.WriteString(line)
			}
		}
	}

	log.Println("Детальные результаты сохранены в gini_detailed.csv")
}

// Основная функция
func main() {
	// Определение флагов
	filePath := flag.String("file", "report.json", "Путь к файлу report.json")
	showHelp := flag.Bool("help", false, "Показать справку")
	verbose := flag.Bool("verbose", false, "Подробный вывод")

	flag.Parse()

	// Вывод справки
	if *showHelp {
		printHelp()
		return
	}

	// Проверка существования файла
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalf("Файл не найден: %s", *filePath)
	}

	if *verbose {
		log.Printf("Начинаю анализ файла: %s", *filePath)
	}

	// Чтение файла
	b, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Ошибка чтения %s: %v", *filePath, err)
	}

	if *verbose {
		log.Printf("Файл прочитан успешно, размер: %d байт", len(b))
	}

	var all allData
	if err := json.Unmarshal(b, &all); err != nil {
		log.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	if *verbose {
		log.Printf("JSON распарсен успешно")
		log.Printf("Найдено сценариев: %d", len(all))
		totalStrategies := 0
		for scenario, data := range all {
			totalStrategies += len(data)
			log.Printf("  Сценарий '%s': %d стратегий", scenario, len(data))
		}
		log.Printf("Всего стратегий: %d", totalStrategies)
	}

	// Сбор данных
	results := collectAllResults(all)

	if *verbose {
		log.Printf("Анализ завершен, обработано %d результатов", len(results))
	}

	// Сохранение CSV
	saveSummaryCSV(results)
	saveDetailedAnalysisCSV(all)

	fmt.Println("Анализ завершён. Файлы: gini_results.csv, gini_detailed.csv")

	// Дополнительная статистика
	if *verbose {
		printAdditionalStats(results)
	}
}

// Функция для отображения справки
func printHelp() {
	fmt.Println("Анализатор статистической справедливости распределения задач")
	fmt.Println("============================================================")
	fmt.Println()
	fmt.Println("Использование:")
	fmt.Println("  analyzer [опции]")
	fmt.Println()
	fmt.Println("Опции:")
	fmt.Println("  --file string     Путь к файлу report.json (по умолчанию: report.json)")
	fmt.Println("  --help            Показать эту справку")
	fmt.Println("  --verbose         Подробный вывод процесса анализа")
	fmt.Println()
	fmt.Println("Примеры:")
	fmt.Println("  analyzer --file=./data/report.json")
	fmt.Println("  analyzer --help")
	fmt.Println()
	fmt.Println("Выходные файлы:")
	fmt.Println("  gini_results.csv    - сводные результаты по стратегиям")
	fmt.Println("  gini_detailed.csv   - детальные результаты по шагам")
	fmt.Println()
	fmt.Println("Метрики расчета:")
	fmt.Println("  Коэффициент Джини по количеству задач")
	fmt.Println("  Коэффициент Джини по энергозатратам")
	fmt.Println("  Среднее и медианное значения по всем шагам")
}

// Дополнительная функция для вывода статистики
func printAdditionalStats(results []summaryResult) {
	fmt.Println("\n *** СТАТИСТИЧЕСКАЯ СПРАВКА ***")

	// Группируем по сценариям
	scenarios := make(map[string][]summaryResult)
	for _, r := range results {
		scenarios[r.Scenario] = append(scenarios[r.Scenario], r)
	}

	for scenario, list := range scenarios {
		fmt.Printf("\nСценарий: %s\n", scenario)
		// Заголовок с компактными метками
		fmt.Printf("%-25s %8s %8s %8s %8s\n",
			"Стратегия", "Ср.зад", "Мед.зад", "Ср.эн", "Мед.эн")
		fmt.Println(strings.Repeat("-", 65))

		for _, result := range list {
			fmt.Printf("%-25s %8.4f %8.4f %8.4f %8.4f\n",
				result.Strategy,
				result.AvgGiniTasks,
				result.MedGiniTasks,
				result.AvgGiniEnergy,
				result.MedGiniEnergy)
		}
	}

	fmt.Println("\nИнтерпретация коэффициентов Джини:")
	fmt.Println("  0.0–0.3: Высокая справедливость")
	fmt.Println("  0.3–0.6: Умеренная неравномерность")
	fmt.Println("  0.6–0.8: Высокая неравномерность")
	fmt.Println("  0.8–1.0: Экстремальная неравномерность")
}

// Вспомогательные функции статистики
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Создаем копию для сортировки
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		// Четное количество элементов
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	// Нечетное количество элементов
	return sorted[n/2]
}
