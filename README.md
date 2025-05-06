# 🚀 gomiggen

A lightweight CLI tool for auto-generating GORM models and versioned migrations using [gormigrate](https://github.com/go-gormigrate/gormigrate). Built for fast and structured database evolution in Go projects.

---

## 📦 Installation

Install the CLI globally with:

```bash
go install github.com/mhdaslam790/gomiggen@latest
```

Ensure your Go `bin` directory is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Now you can run `gomiggen` from anywhere in your project.

---

## 🔧 Requirements

- **Go 1.18+**
- [GORM](https://gorm.io)
- [gormigrate](https://github.com/go-gormigrate/gormigrate)

Install required dependencies in your Go module:

```bash
go get -u gorm.io/gorm
go get -u github.com/go-gormigrate/gormigrate/v2
```

---

## 📁 Folder Structure

After using `gomiggen`, your project will include:

```
.
├── model/                      # Auto-generated model structs
│   └── 202505051245_user.go
├── migration/                 
│   └── migration.go            # Auto-managed migration file
└── main.go                     # Your app entry (you wire up NewMigration)
```

---

## 🧪 Usage

Run the CLI to generate migrations and models:

```bash
gomiggen generate <command> <ModelName> [args...]
```

### Available Commands

| Command        | Example                                             | Description                           |
|----------------|-----------------------------------------------------|---------------------------------------|
| `create`       | `gomiggen create User`                              | Create a model + CreateTable migration |
| `add-column`   | `gomiggen add-column User Name:string:"column:Name;size:255;not null"`         | Add a column + migration              |
| `drop-column`  | `gomiggen drop-column User age`                      | Drop a column + rollback              |
| `add-index`    | `gomiggen add-index User email`                      | Create index on column                |
| `drop-index`   | `gomiggen drop-index User email`                     | Drop index from column                |

---

## 📄 Example Inserted Migration

When you run `gomiggen generate add-column User age:int`, the following code is inserted into `migration/migration.go`:

```go
{
	ID: "202505051245",
	Migrate: func(tx *gorm.DB) error {
		type User struct {
			Age int
		}
		return tx.Migrator().AddColumn(User{}, "Age")
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropColumn(User{}, "Age")
	},
},
```

---

## 🏗 Auto Generated migration.go (if missing)

If `migration/migration.go` does not exist, it is created automatically with:

```go
package migration

import (
	"log"
	"yourapp/model"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func NewMigration(db *gorm.DB) {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		// generated migrations inserted here
	})

	m.InitSchema(func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			// generated models inserted here for initSchema
		)
	})

	if err := m.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration ran successfully")
}


```

As you generate more models, they are automatically added to `getModels()`.

---

## 🧼 Features

- ✅ Automatically creates migration file if missing
- ✅ Auto-appends new migrations
- ✅ Dynamically populates `AutoMigrate()`
- ✅ Reversible migrations via Rollback
- ✅ Timestamped file naming for model history

---

## 📌 Tips

- Keep your `model/` and `migration/` folders version-controlled.
- Customize generated structs as needed after creation.
- Run your app's migrations by calling `migration.NewMigration(db)` in `main.go`.

---

## 🧭 Roadmap Ideas

- [ ] Support multiple column changes at once
- [ ] Auto format code with `gofmt`
- [ ] Add `modify-column` support
- [ ] Index existence checks before creating/dropping

---

## 📜 License

MIT © [mhdaslam790](https://github.com/mhdaslam790)

---

## 💬 Feedback & Contributions

Found a bug or want to suggest a feature? Open an [issue](https://github.com/mhdaslam790/gomiggen/issues) or a PR.
