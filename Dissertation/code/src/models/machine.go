package models

// Machine описывает машину
type Machine struct {
	// уникальный идентификатор
	ID string `json:"id" yaml:"id"`
	// тип оборудования
	Equipment []TypeEquipment `json:"equipment" yaml:"equipment"`
	// запас энергии
	Energy float64 `json:"energy" yaml:"energy"`
	// выполняемая задача
	AssignedTask *Task
}

// TypeEquipment - типы оборудования
// Примеры по технике:
//
// экскаватор : ковш, грейфер, рыхлитель (1, 3, 9)
// бульдозер : отвал, рыхлитель (2, 9)
// погрузчик : ковш, захват, вилы (для паллет) (1, 7, 14)
// снегоуборочная машина : плуг, фреза, щётка, метатель (6, 5, 4, 15)
// автогрейдер : нож-отвал (11)
// виброплита : трамбовочное устройство, виброплита (8, 13)
// конвейер : конвейер ленточный (10)
// скрепер : скребок (12)
type TypeEquipment uint

const (
	Bucket     TypeEquipment = iota + 1 // ковш
	Dump                                // отвал
	Grapple                             // грейфер
	Brushes                             // щетки
	Cutter                              // фреза
	Plow                                // плуг
	Catch                               // захват
	Tamp                                // трамбовочное устройство
	Ripper                              // рыхлитель
	Scraper                             // скребок
	Conveyor                            // конвейер ленточный
	Knife                               // нож-отвал автогрейдера
	VCompactor                          // виброплита
	Pitchfork                           // вилы (для паллет)
	Thrower                             // метатель
)

func IntersectTypeEquipment(in TypeEquipment, t []TypeEquipment) bool {
	for i := range t {
		if t[i] == in {
			return true
		}
	}
	return false
}

func TotalEnergyMachines(in []Machine) float64 {
	total := 0
	for i := range in {
		total += int(in[i].Energy)
	}
	return float64(total)
}
