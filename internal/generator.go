package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type Field struct {
	Name    string
	Type    string
	GormTag string
	Comment string
}

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
	timeStr := time.Now().Format("20060102150405") // Updated to include seconds
	fileBase := toSnakeCase(structName)
	modelDir := filepath.Join("model")
	os.MkdirAll(modelDir, os.ModePerm)

	modelPath := filepath.Join(modelDir, fileBase+".go")

	modelCode := fmt.Sprintf(`package model

	type %s struct {
	ID uint `+"`gorm:\"primaryKey\"`"+`
	}
`, structName)

	os.WriteFile(modelPath, []byte(modelCode), 0644)
	fmt.Println("✅ Model created:", modelPath)

	// Generate migration
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
	inserInitSchema(structName)
}
func inserInitSchema(structName string) {
	tmpl := `&model.{{.StructName}}{},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"StructName": structName,
	})
	migrationDir := "migration"
	migrationGo := filepath.Join(migrationDir, "migration.go")
	data, err := os.ReadFile(migrationGo)
	if err != nil {
		panic(err)
	}
	insertAfter := "tx.AutoMigrate("
	idx := strings.Index(string(data), insertAfter)
	if idx == -1 {
		panic("Could not find the implementation of AutoMigration")
	}
	insertPos := idx + len(insertAfter)
	updated := string(data[:insertPos]) + "\n\t\t\t\t" + sb.String() + string(data[insertPos:])
	err = os.WriteFile(migrationGo, []byte(updated), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅Auto migration entry inserted")
}

func handleAddColumn(model string, columns ...string) {
	if len(columns) == 0 {
		fmt.Println("❌ No column(s) specified.")
		return
	}

	if !modelExists(model) {
		fmt.Printf("❌ Model '%s' not found in model directory.\n", model)
		return
	}

	timeStr := timestamp()
	now := time.Now().Format("2006-01-02 15:04")
	var structFields []string
	var addCalls []string
	var dropCalls []string
	var fields []Field
	firstCol := ""

	// Process each column provided
	for i, col := range columns {
		// Split column into name and type (and optionally the GORM tag)
		parts := strings.SplitN(col, ":", 3)

		// Default to string type if no type is provided
		field := Field{
			Name:    parts[0],
			Type:    parts[1],
			Comment: "// Added " + now,
		}

		// If a GORM tag is provided
		if len(parts) == 3 {
			field.GormTag = parts[2]
		}

		// Set the first column name for the migration
		if i == 0 {
			firstCol = field.Name
		}

		// Add the field to the list
		fields = append(fields, field)

		// Struct line for migration (considering GORM tag if present)
		tag := ""
		if field.GormTag != "" {
			tag = fmt.Sprintf(" `gorm:\"%s\"`", field.GormTag)
		}

		// Add this field to the struct definition
		structFields = append(structFields, fmt.Sprintf("		%s %s%s %s", field.Name, field.Type, tag, field.Comment))

		// Add migration commands to add and drop the column
		addCalls = append(addCalls, fmt.Sprintf(`		if err := tx.Migrator().AddColumn(%s{}, "%s"); err != nil { return err }`, model, field.Name))
		dropCalls = append(dropCalls, fmt.Sprintf(`		if err := tx.Migrator().DropColumn(&%s{}, "%s"); err != nil { return err }`, model, field.Name))
	}

	// 🔧 1. Update migration.go
	tmpl := `{
				ID: "{{.Timestamp}}",
				Migrate: func(tx *gorm.DB) error {
					type {{.Model}} struct {
			{{.StructFields}}
					}
			{{.AddCalls}}
					return nil
				},
				Rollback: func(tx *gorm.DB) error {
				type {{.Model}} struct {
			{{.StructFields}}
					}
			{{.DropCalls}}
					return nil
				},
			},`
	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp":    timeStr,
		"Model":        model,
		"FirstCol":     firstCol,
		"StructFields": strings.Join(structFields, "\n"),
		"AddCalls":     strings.Join(addCalls, "\n"),
		"DropCalls":    strings.Join(dropCalls, "\n"),
	})
	writeMigration(sb.String())

	// 🏗️ 2. Update actual model file
	updateModelStruct(model, fields)

	fmt.Printf("✅ Added column '%s' to model '%s'.\n", structFields, model)
	fmt.Println("✅ Migration and model updated.")
}

func handleDropColumn(model string, colNames ...string) {
	if len(colNames) == 0 {
		fmt.Println("❌ No columns specified to drop.")
		return
	}

	var rollback []string
	for _, col := range colNames {
		rollback = append(rollback, fmt.Sprintf(`		// Re-add column %s manually if needed`, col))
	}

	tmpl := `{
			ID: "{{.Timestamp}}",
			Migrate: func(tx *gorm.DB) error {
			{{.DropLines}}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
			{{.Rollback}}
				return nil
			},
			},`

	var dropLines string
	for _, col := range colNames {
		dropLines += fmt.Sprintf(`		if err := tx.Migrator().DropColumn(&model.%s{}, "%s"); err != nil {
		return err
	}
`, model, col)
	}

	t := template.Must(template.New("mig").Parse(tmpl))
	var sb strings.Builder
	t.Execute(&sb, map[string]string{
		"Timestamp": timestamp(),
		"DropLines": dropLines,
		"Rollback":  strings.Join(rollback, "\n"),
	})
	writeMigration(sb.String())
}

func handleAddIndex(model, col string) {

	if !modelExists(model) {
		fmt.Printf("❌ Model '%s' not found in model directory.\n", model)
		return
	}
	table := toSnakeCase(model) + "s"
	idx := "idx_" + col
	lowerCaseColumn := strings.ToLower(col)

	migrationDir := "migration"
	migrationGo := filepath.Join(migrationDir, "migration.go")
	data, err := os.ReadFile(migrationGo)

	if err != nil {
		fmt.Println("❌ Failed to read model directory: ", err)
		return
	}
	stringPattern := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", idx, table, lowerCaseColumn)
	if strings.Contains(string(data), stringPattern) {
		fmt.Printf("❌ Index already exist for %s the on %s", table, col)
		return
	}
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
		"Col":       lowerCaseColumn,
	})
	writeMigration(sb.String())

	// Insert index into the table
	modelDir := "model"
	fileName := fmt.Sprintf("%s.go", model)
	modelGo := filepath.Join(modelDir, fileName)
	modelData, err := os.ReadFile(modelGo)
	if err != nil {
		fmt.Printf("❌ Model '%s' not found in the directory", modelGo)
		return
	}

	lines := strings.Split(string(modelData), "\n")
	updated := []string{}
	var fieldFound bool = false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, col+" ") && !strings.Contains(line, "struct") {
			if strings.Contains(line, "`gorm:\"") {
				line = strings.Replace(line, "`gorm:\"", "`gorm:\"index;", 1)
			} else {
				line = strings.Replace(line, col, fmt.Sprintf("%s %s `gorm:\"index;\"`", col, strings.Split(trimmed, " ")[1]), 1)
			}
			fieldFound = true
		}
		updated = append(updated, line)
	}
	if fieldFound {
		err = os.WriteFile(modelGo, []byte(strings.Join(updated, "\n")), 0644)
		if err != nil {
			fmt.Printf("❌ Failed to write updated model: %v\n", err)
			return
		}
		fmt.Printf("✅ Index tag added to column '%s' in model '%s'\n", col, model)
	} else {
		fmt.Printf("⚠️ Field '%s' not found in model '%s'. No changes made to struct.\n", col, model)
	}
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

func updateModelStruct(structName string, fields []Field) {
	modelDir := "model"
	files, err := os.ReadDir(modelDir)
	if err != nil {
		fmt.Println("❌ Failed to read model directory:", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			path := filepath.Join(modelDir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			content := string(data)
			if strings.Contains(content, fmt.Sprintf("type %s struct {", structName)) {
				lines := strings.Split(content, "\n")
				var updated []string
				inserted := false
				for _, line := range lines {
					updated = append(updated, line)
					if strings.Contains(line, fmt.Sprintf("type %s struct {", structName)) && !inserted {
						for _, f := range fields {
							if f.GormTag != "" {
								updated = append(updated, fmt.Sprintf("	%s %s `gorm:\"%s\"` // Added %s", f.Name, f.Type, f.GormTag, time.Now().Format("2006-01-02 15:04")))
							} else {
								updated = append(updated, fmt.Sprintf("	%s %s // Added %s", (f.Name), f.Type, time.Now().Format("2006-01-02 15:04")))
							}
						}
						inserted = true
					}
				}
				os.WriteFile(path, []byte(strings.Join(updated, "\n")), 0644)
				fmt.Println("✅ Model updated:", path)
				return
			}
		}
	}
	fmt.Println("❌ Model struct not found:", structName)
}

func modelExists(structName string) bool {
	modelDir := "model"
	files, err := os.ReadDir(modelDir)
	if err != nil {
		return false
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			content, err := os.ReadFile(filepath.Join(modelDir, file.Name()))
			if err == nil && strings.Contains(string(content), fmt.Sprintf("type %s struct {", structName)) {
				return true
			}
		}
	}
	return false
}
