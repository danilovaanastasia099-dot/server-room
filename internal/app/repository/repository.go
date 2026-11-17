package repository

import (
	"fmt"
)

type Repository struct {
	serverComponents []ServerComponent
}

type ServerComponent struct {
	ID          int
	Name        string
	Description string
	Type        string // CPU, GPU, PSU, STORAGE, NIC, RAM
	TDP         int    // тепловыделение в ваттах
	ImageURL    string // URL изображения из Minio
	VideoURL    string // URL видео из Minio
	Status      string // draft, published, deleted
	Likes       []int  // ID пользователей, поставивших лайк
}

func NewRepository() (*Repository, error) {
	serverComponents := []ServerComponent{
		{
			ID:          1,
			Name:        "Intel Xeon E5-2680 v4",
			Description: "14-ядерный серверный процессор для стоечных решений. Надёжный выбор для виртуализации и вычислительных узлов.",
			Type:        "CPU",
			TDP:         120,
			ImageURL:    "http://localhost:9000/server-images/cpu-xeon.jpg",
			VideoURL:    "http://localhost:9000/server-videos/cpu-xeon.mp4",
			Status:      "published",
			Likes:       []int{1, 2, 3},
		},
		{
			ID:          2,
			Name:        "NVIDIA Tesla V100",
			Description: "Графический ускоритель для вычислений с TDP 250 Вт. Подходит для ML/AI вычислений и вычислительных шардов.",
			Type:        "GPU",
			TDP:         250,
			ImageURL:    "http://localhost:9000/server-images/gpu-tesla.jpg",
			VideoURL:    "http://localhost:9000/server-videos/gpu-tesla.mp4",
			Status:      "published",
			Likes:       []int{1, 2},
		},
		{
			ID:          3,
			Name:        "Блок питания 800W",
			Description: "Серверный блок питания мощностью 800 Вт с высоким КПД и активным PFC.",
			Type:        "PSU",
			TDP:         800,
			ImageURL:    "http://localhost:9000/server-images/psu-800.jpg",
			VideoURL:    "http://localhost:9000/server-videos/psu-800.mp4",
			Status:      "published",
			Likes:       []int{1},
		},
		{
			ID:          4,
			Name:        "AMD EPYC 7742",
			Description: "64-ядерный серверный процессор с TDP 225 Вт. Отличается высокой плотностью вычислений и производительностью на ядро.",
			Type:        "CPU",
			TDP:         225,
			ImageURL:    "http://localhost:9000/server-images/cpu-epyc.jpg",
			VideoURL:    "http://localhost:9000/server-videos/cpu-epyc.mp4",
			Status:      "draft",
			Likes:       []int{},
		},
		{
			ID:          5,
			Name:        "NVIDIA RTX A6000",
			Description: "Профессиональная видеокарта с высокой производительностью рендеринга и TDP 300 Вт.",
			Type:        "GPU",
			TDP:         300,
			ImageURL:    "http://localhost:9000/server-images/gpu-rtx.jpg",
			VideoURL:    "http://localhost:9000/server-videos/gpu-rtx.mp4",
			Status:      "published",
			Likes:       []int{4, 5},
		},
		{
			ID:          6,
			Name:        "Samsung PM1733 3.84TB",
			Description: "NVMe SSD для серверов, высокопроизводительное хранилище с низкой латентностью. TDP условный 12 Вт.",
			Type:        "STORAGE",
			TDP:         12,
			ImageURL:    "http://localhost:9000/server-images/ssd-pm1733.jpg",
			VideoURL:    "http://localhost:9000/server-videos/ssd-pm1733.mp4",
			Status:      "published",
			Likes:       []int{2, 3},
		},
		{
			ID:          7,
			Name:        "Mellanox ConnectX-6",
			Description: "Сетевая карта 200GbE/100GbE для высокоскоростных кластеров. Низкая задержка, высокая пропускная способность.",
			Type:        "NIC",
			TDP:         25,
			ImageURL:    "http://localhost:9000/server-images/nic-connectx6.jpg",
			VideoURL:    "http://localhost:9000/server-videos/nic-connectx6.mp4",
			Status:      "published",
			Likes:       []int{1},
		},
		{
			ID:          8,
			Name:        "Kingston 128GB DDR4 ECC",
			Description: "Модуль оперативной памяти ECC для серверных систем, 128GB, DDR4. Надёжность и совместимость.",
			Type:        "RAM",
			TDP:         8,
			ImageURL:    "http://localhost:9000/server-images/ram-128gb.jpg",
			VideoURL:    "http://localhost:9000/server-videos/ram-128gb.mp4",
			Status:      "published",
			Likes:       []int{1, 2},
		},
	}

	return &Repository{
		serverComponents: serverComponents,
	}, nil
}

// GetPublishedServerComponents возвращает все опубликованные компоненты
func (r *Repository) GetPublishedServerComponents() ([]ServerComponent, error) {
	var result []ServerComponent
	for _, component := range r.serverComponents {
		if component.Status == "published" {
			result = append(result, component)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("нет опубликованных компонентов")
	}
	return result, nil
}

// GetServerComponentByID возвращает компонент по ID
func (r *Repository) GetServerComponentByID(id int) (ServerComponent, error) {
	for _, component := range r.serverComponents {
		if component.ID == id && component.Status != "deleted" {
			return component, nil
		}
	}
	return ServerComponent{}, fmt.Errorf("компонент не найден")
}

// GetNextServerComponent возвращает следующий опубликованный компонент после указанного ID
func (r *Repository) GetNextServerComponent(id int) (ServerComponent, error) {
	for i, component := range r.serverComponents {
		if component.ID == id {
			for j := i + 1; j < len(r.serverComponents); j++ {
				if r.serverComponents[j].Status == "published" {
					return r.serverComponents[j], nil
				}
			}
			break
		}
	}
	return ServerComponent{}, fmt.Errorf("следующий компонент не найден")
}

// GetDraftServerComponent возвращает черновик
func (r *Repository) GetDraftServerComponent() (ServerComponent, error) {
	for _, component := range r.serverComponents {
		if component.Status == "draft" {
			return component, nil
		}
	}
	return ServerComponent{}, fmt.Errorf("черновик не найден")
}

// FilterServerComponentsByTDP фильтрует компоненты по минимальному TDP
func (r *Repository) FilterServerComponentsByTDP(minTDP int) ([]ServerComponent, error) {
	var result []ServerComponent
	for _, component := range r.serverComponents {
		if component.Status == "published" && component.TDP >= minTDP {
			result = append(result, component)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("компоненты не найдены")
	}
	return result, nil
}

// CalculateTotalTDP вычисляет общее тепловыделение опубликованных компонентов
func (r *Repository) CalculateTotalTDP() int {
	total := 0
	for _, component := range r.serverComponents {
		if component.Status == "published" {
			total += component.TDP
		}
	}
	return total
}

// ToggleLike переключает лайк пользователя на компоненте
func (r *Repository) ToggleLike(componentID, userID int) (bool, error) {
	for index, component := range r.serverComponents {
		if component.ID != componentID || component.Status == "deleted" {
			continue
		}
		for likeIndex, likedUserID := range component.Likes {
			if likedUserID == userID {
				r.serverComponents[index].Likes = append(component.Likes[:likeIndex], component.Likes[likeIndex+1:]...)
				return false, nil
			}
		}
		r.serverComponents[index].Likes = append(component.Likes, userID)
		return true, nil
	}
	return false, fmt.Errorf("компонент не найден")
}

// AddServerComponent добавляет новый компонент
func (r *Repository) AddServerComponent(component ServerComponent) {
	component.ID = len(r.serverComponents) + 1
	component.Status = "published"
	r.serverComponents = append(r.serverComponents, component)
}

// ConvertToBTU конвертирует ватты в BTU/час
func ConvertToBTU(watts int) float64 {
	return float64(watts) * 3.412141633
}
