package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	postgre_models "github.com/Vilamuzz/yota-backend/models"
)

func main() {
	// Disable logging for gorm so it doesn't pollute the schema output
	log.SetOutput(io.Discard)

	// Get all models
	models := postgre_models.GetAllModels()

	// Generate the SQL schema string
	stmts, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}

	// Atlas expects the SQL statements to be printed to stdout
	fmt.Print(stmts)
}
