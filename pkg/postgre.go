package pkg

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func AutoMigrateDB(db *gorm.DB, models ...interface{}) {
	// db.Migrator().DropTable(models...)
	err := db.AutoMigrate(models...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database migrated successfully!")
}
