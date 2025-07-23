package models

// Machine описывает машину
type Machine struct {
	ID           string          `yaml:"id"`        // уникальный идентификатор
	Equipment    []TypeEquipment `yaml:"equipment"` // тип оборудования
	Energy       float64         `yaml:"energy"`    // запас энергии
	AssignedTask *Task           // выполняемая задача
}

// TypeEquipment - типы оборудования
// Примеры по технике:
//
//	экскаватор : ковш, грейфер, гидромолот, рыхлитель
//	бульдозер : отвал, рыхлитель
//	погрузчик : ковш, захват, вилы (для паллет)
//	снегоуборочная машина : плуг, фреза, щётка, метатель
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
