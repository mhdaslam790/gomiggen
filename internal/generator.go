package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func timestamp() string {
	return time.Now().Format("200601021504")
}

func PrintUsage() {
	fmt.Println(`Usage:
  miggen create <ModelName>
  miggen add-column <ModelName> <column:type>
  miggen drop-column <ModelName> <column>
  miggen add-index <ModelName> <column>
  miggen drop-index <ModelName> <column>`)
}

func HandleAction(args []string) {
	if len(args) < 2 {
		PrintUsage()
		return
	}
	action := args[0]
	model := args[1]

	switch action {
	case "create":
		handleCreateModel(model)
	case "add-column":
		if len(args) < 3 {
			fmt.Println("Missing column:type argument")
			return
		}
		handleAddColumn(model, args[2])
	case "drop-column":
		if len(args) < 3 {
			fmt.Println("Missing column name")
			return
		}
		handleDropColumn(model, args[2])
	case "add-index":
		if len(args) < 3 {
			fmt.Println("Missing column name")
			return
		}
		handleAddIndex(model, args[2])
	case "drop-index":
		if len(args) < 3 {
			fmt.Println("Missing column name")
			return
		}
		handleDropIndex(model, args[2])
	default:
		fmt.Println("Unknown action:", action)
		PrintUsage()
	}
}

func writeMigration(content string) {
	migrationDir := "migration"
	migrationGo := filepath.Join(migrationDir, "migration.go")

	if _, err := os.Stat(migrationGo); os.IsNotExist(err) {
		// Create directory if needed
		os.MkdirAll(migrationDir, os.ModePerm)

		// Initial migration.go structure
		initialContent := `package migration

		import (
			"log"
			"your/module/path/model"
		
			"github.com/go-gormigrate/gormigrate/v2"
			"gorm.io/gorm"
		)
		
		func NewMigration(db *gorm.DB) {
			m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
				// migrations inserted here
			})
		
			m.InitSchema(func(tx *gorm.DB) error {
				err := tx.AutoMigrate(
					getModels()...,
				)
				if err != nil {
					log.Fatal(err)
				}
				return nil
			})
		
			if err := m.Migrate(); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
			log.Println("Migration did run successfully")
		}
		
		func getModels() []interface{} {
			return []interface{}{
				// &model.User{},
				// Add other models here
			}
		}
		`

		err = os.WriteFile(migrationGo, []byte(initialContent), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Println("✅ Created migration.go with initial structure.")
	}

	// Insert new migration into the array
	data, err := os.ReadFile(migrationGo)
	if err != nil {
		panic(err)
	}
	insertAfter := "[]*gormigrate.Migration{"
	idx := strings.Index(string(data), insertAfter)
	if idx == -1 {
		panic("Could not find migration array")
	}
	insertPos := idx + len(insertAfter)
	updated := string(data[:insertPos]) + "\n" + content + string(data[insertPos:])
	err = os.WriteFile(migrationGo, []byte(updated), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ Migration entry inserted.")
}

func handleCreateModel(structName string) {
	timeStr := timestamp()
	fileBase := toSnakeCase(structName)
	modelDir := filepath.Join("model")
	os.MkdirAll(modelDir, os.ModePerm)

	modelPath := filepath.Join(modelDir, timeStr+"_"+fileBase+".go")
	modelCode := fmt.Sprintf(`package model

type %s struct {
	ID uint `+"`gorm:\"primaryKey\"`"+`
	// Add your fields here
}
`, structName)
	os.WriteFile(modelPath, []byte(modelCode), 0644)
	fmt.Println("✅ Model created:", modelPath)

	// Step 1: Add migration block
	tmpl := `{
ID: "{{.Timestamp}}",
Migrate: func(tx *gorm.DB) error {
	return tx.Migrator().CreateTable(&model.{{.StructName}}{})
},
Rollback: func(tx *gorm.DB) error {
	return tx.Migrator().DropTable("{{.TableName}}")
},
},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp":  timeStr,
		"StructName": structName,
		"TableName":  fileBase + "s",
	})
	writeMigration(sb.String())

	// Step 2: Add model to getModels()
	migPath := filepath.Join("migration", "migration.go")
	data, err := os.ReadFile(migPath)
	if err != nil {
		panic(err)
	}

	// Find the line with `return []interface{}{`
	insertLine := "&model." + structName + "{},"
	lines := strings.Split(string(data), "\n")
	var modified []string
	inserted := false
	for _, line := range lines {
		modified = append(modified, line)
		if !inserted && strings.TrimSpace(line) == "return []interface{}{" {
			modified = append(modified, "\t\t"+insertLine)
			inserted = true
		}
	}

	if inserted {
		err := os.WriteFile(migPath, []byte(strings.Join(modified, "\n")), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Println("✅ Model added to AutoMigrate:", insertLine)
	} else {
		fmt.Println("⚠️ Could not find getModels() to insert AutoMigrate line")
	}
}

func handleAddColumn(model, colType string) {
	parts := strings.SplitN(colType, ":", 2)
	colName := parts[0]
	typeGo := "string"
	if len(parts) > 1 {
		typeGo = parts[1]
	}
	tmpl := `{
ID: "{{.Timestamp}}",
Migrate: func(tx *gorm.DB) error {
	type T struct {
		{{.ColName}} {{.ColType}}
	}
	return tx.Migrator().AddColumn(&model.{{.Model}}{}, "{{.ColName}}")
},
Rollback: func(tx *gorm.DB) error {
	return tx.Migrator().DropColumn(&model.{{.Model}}{}, "{{.ColName}}")
},
},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp": timestamp(),
		"Model":     model,
		"ColName":   colName,
		"ColType":   typeGo,
	})
	writeMigration(sb.String())
}

func handleDropColumn(model, colName string) {
	tmpl := `{
ID: "{{.Timestamp}}",
Migrate: func(tx *gorm.DB) error {
	return tx.Migrator().DropColumn(&model.{{.Model}}{}, "{{.ColName}}")
},
Rollback: func(tx *gorm.DB) error {
	// You may need to re-add the column manually
	return nil
},
},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp": timestamp(),
		"Model":     model,
		"ColName":   colName,
	})
	writeMigration(sb.String())
}

func handleAddIndex(model, col string) {
	table := toSnakeCase(model) + "s"
	idx := "idx_" + col
	tmpl := `{
				ID: "{{.Timestamp}}",
				Migrate: func(tx *gorm.DB) error {
					return tx.Exec("CREATE INDEX {{.Index}} ON {{.Table}} ({{.Col}})").Error
				},
				Rollback: func(tx *gorm.DB) error {
					return tx.Exec("DROP INDEX {{.Index}} ON {{.Table}}")
				},
			},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp": timestamp(),
		"Table":     table,
		"Index":     idx,
		"Col":       col,
	})
	writeMigration(sb.String())
}

func handleDropIndex(model, col string) {
	table := toSnakeCase(model) + "s"
	idx := "idx_" + col
	tmpl := `{
		ID: "{{.Timestamp}}",
		Migrate: func(tx *gorm.DB) error {
			return tx.Exec("DROP INDEX {{.Index}} ON {{.Table}}")
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Exec("CREATE INDEX {{.Index}} ON {{.Table}} ({{.Col}})").Error
		},
		},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp": timestamp(),
		"Table":     table,
		"Index":     idx,
		"Col":       col,
	})
	writeMigration(sb.String())
}
