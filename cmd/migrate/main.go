package main

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"server_components/internal/app/ds"
	"server_components/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	dbURL := dsn.FromEnv()
	fmt.Println("Connecting to PostgreSQL for migration:", dbURL)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("Failed to connect database: %v", err)
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.ServerComponent{},
		&ds.ComponentLike{},
	)
	if err != nil {
		logrus.Fatalf("Can't migrate DB: %v", err)
	}
	fmt.Println("Tables migrated successfully: users, server_components, component_likes")

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_single_draft ON server_components (creator_id) WHERE status = 'черновик';")

	var count int64
	db.Model(&ds.User{}).Count(&count)
	if count == 0 {
		fmt.Println("Seeding initial users...")
		users := []ds.User{
			{ID: 1, Login: "admin", Password: "password123", IsModerator: true},
			{ID: 2, Login: "operator", Password: "password123", IsModerator: false},
			{ID: 3, Login: "engineer", Password: "password123", IsModerator: false},
		}
		db.Create(&users)
		db.Exec("SELECT setval('users_id_seq', (SELECT COALESCE(MAX(id), 1) FROM users));")
	}

	var compCount int64
	db.Model(&ds.ServerComponent{}).Count(&compCount)
	if compCount == 0 {
		fmt.Println("Seeding initial server components (with published, draft, deleted)...")
		now := time.Now()
		past := now.Add(-24 * time.Hour)

		components := []ds.ServerComponent{
			{
				ID:            1,
				Name:          "Intel Xeon E5-2680 v4",
				Description:   "14-ядерный серверный процессор для стоечных серверов. Надёжное решение для виртуализации и вычислительных узлов.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/cpu-xeon.jpg",
				VideoURL:      "http://localhost:9000/server-videos/cpu-xeon.mp4",
				ComponentType: "CPU",
				TDP:           120,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
			{
				ID:            2,
				Name:          "NVIDIA Tesla V100",
				Description:   "Графический ускоритель для вычислений с TDP 250 Вт. Предназначен для задач искусственного интеллекта и глубокого обучения.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/gpu-tesla.jpg",
				VideoURL:      "http://localhost:9000/server-videos/gpu-tesla.mp4",
				ComponentType: "GPU",
				TDP:           250,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
			{
				ID:            3,
				Name:          "Блок питания 800W",
				Description:   "Резервируемый серверный блок питания мощностью 800 Вт стандарта 80 PLUS Platinum с горячей заменой.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/psu-800.jpg",
				VideoURL:      "http://localhost:9000/server-videos/psu-800.mp4",
				ComponentType: "PSU",
				TDP:           800,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
			{
				ID:            4,
				Name:          "Mellanox ConnectX-6 200GbE",
				Description:   "Двухпортовый сетевой адаптер 200Gb/s HDR InfiniBand и Ethernet для сверхбыстрых кластеров и СХД.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/nic-connectx6.jpg",
				VideoURL:      "http://localhost:9000/server-videos/nic-connectx6.mp4",
				ComponentType: "NIC",
				TDP:           25,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
			{
				ID:            5,
				Name:          "Samsung PM1733 3.84TB NVMe",
				Description:   "Серверный твердотельный накопитель NVMe U.3 PCIe Gen4 с высокой скоростью чтения/записи для корпоративных баз данных.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/ssd-pm1733.jpg",
				VideoURL:      "http://localhost:9000/server-videos/ssd-pm1733.mp4",
				ComponentType: "STORAGE",
				TDP:           18,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
			{
				ID:            6,
				Name:          "DDR4 ECC Reg 128GB",
				Description:   "Модуль оперативной памяти серверного класса с коррекцией ошибок ECC Registered для максимальной стабильности под нагрузкой.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/ram-128gb.jpg",
				VideoURL:      "http://localhost:9000/server-videos/ram-128gb.mp4",
				ComponentType: "RAM",
				TDP:           15,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
			{
				ID:            7,
				Name:          "AMD EPYC 7742",
				Description:   "64-ядерный серверный процессор с TDP 225 Вт. Высокая плотность ядер для облачных дата-центров.",
				Status:        "опубликован",
				ImageURL:      "http://localhost:9000/server-images/cpu-epyc.jpg",
				VideoURL:      "http://localhost:9000/server-videos/cpu-epyc.mp4",
				ComponentType: "CPU",
				TDP:           225,
				CreatedAt:     now,
				CreatorID:     1,
				FormedAt:      &now,
			},
			{
				ID:            8,
				Name:          "NVIDIA RTX A6000",
				Description:   "Профессиональная видеокарта для серверов рендеринга и VDI рабочих станций.",
				Status:        "удален",
				ImageURL:      "http://localhost:9000/server-images/gpu-rtx.jpg",
				VideoURL:      "http://localhost:9000/server-videos/gpu-rtx.mp4",
				ComponentType: "GPU",
				TDP:           300,
				CreatedAt:     past,
				CreatorID:     1,
				FormedAt:      &past,
			},
		}
		db.Create(&components)
		db.Exec("SELECT setval('server_components_id_seq', (SELECT MAX(id) FROM server_components));")
	}

	var likesCount int64
	db.Model(&ds.ComponentLike{}).Count(&likesCount)
	if likesCount == 0 {
		fmt.Println("Seeding initial component likes (M-M)...")
		likes := []ds.ComponentLike{
			{UserID: 1, ComponentID: 1},
			{UserID: 2, ComponentID: 1},
			{UserID: 3, ComponentID: 1},
			{UserID: 1, ComponentID: 2},
			{UserID: 2, ComponentID: 2},
			{UserID: 1, ComponentID: 3},
			{UserID: 2, ComponentID: 4},
			{UserID: 3, ComponentID: 4},
			{UserID: 1, ComponentID: 5},
			{UserID: 2, ComponentID: 6},
		}
		db.Create(&likes)
	}

	fmt.Println("Migration and initial data seeding completed successfully!")
}
