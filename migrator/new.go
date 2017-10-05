package migrator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var config *Config

type MigrationFile struct {
	Timestamp      time.Time
	Name           string
	FuncName       string
	OrderingNumber int64
}

func (m *MigrationFile) Unix() int64 {
	return m.Timestamp.Unix()
}

func (m *MigrationFile) FilePath() string {
	return filepath.Join(config.LocalMigrationsPath, m.FileName())
}

func (m *MigrationFile) FileName() string {
	return m.Name + ".go"
}

func newMigrationFile(name string) *MigrationFile {
	name = strings.Replace(name, " ", "_", 100)
	m := &MigrationFile{
		Timestamp: time.Now(),
	}
	secondsInDay := m.Timestamp.Unix() % 86400
	m.Name = fmt.Sprintf("%d_%02d_%02d_%05d_%s", m.Timestamp.Year(), m.Timestamp.Month(), m.Timestamp.Day(), secondsInDay, name)
	m.FuncName = fmt.Sprintf("Migration%d%02d%02d%05d%s", m.Timestamp.Year(), m.Timestamp.Month(), m.Timestamp.Day(), secondsInDay, camelCase(name))
	m.OrderingNumber, _ = strconv.ParseInt(fmt.Sprintf("%d%02d%02d%05d", m.Timestamp.Year(), m.Timestamp.Month(), m.Timestamp.Day(), secondsInDay), 10, 64)
	return m
}

func NewMigrationFile(options *NewOptions) {
	config = LoadConfig()

	m := newMigrationFile(*options.Name)
	renderTemplate("new_migration.tmpl", m)

	updateMainMigrationFile(m)

	fmt.Println("Created migration", m.FilePath())
}

func renderTemplate(templateName string, migration *MigrationFile) {
	b, err := ReadFile(filepath.Join(config.LocalTemplatesPath, templateName))
	if err != nil {
		panic("Could not read template " + templateName + ": " + err.Error())
	}

	parsedTemplate, err := template.New(migration.Name).Parse(string(b))
	if err != nil {
		panic("Could not render template " + templateName + ": " + err.Error())
	}

	buf := bytes.NewBuffer(nil)
	parsedTemplate.Execute(buf, migration)

	if err := WriteFile(migration.FilePath(), buf.Bytes()); err != nil {
		panic("Could not write to file " + migration.FilePath() + ": " + err.Error())
	}
}

func ReadFile(source string) ([]byte, error) {
	sourceFile, err := os.Open(source)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read %s", source)
	}

	defer sourceFile.Close()
	content, err := ioutil.ReadAll(sourceFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read %s", source)
	}

	return content, nil
}

func WriteFile(destName string, content []byte) error {
	dest, err := os.Create(destName)
	if err != nil {
		return errors.Wrapf(err, "Could not create file %s", destName)
	}
	defer dest.Close()

	if _, err := dest.Write(content); err != nil {
		return errors.Wrapf(err, "Could not write to file %s", destName)
	}
	return nil
}

// camelCase converts a name with_underscores to a name with CamelCase.
func camelCase(name string) string {
	result := make([]byte, len(name))
	upCaseNext := true
	j := 0
	for i := 0; i < len(name); i++ {

		if (name[i] >= 'a' && name[i] <= 'z') || (name[i] >= 'A' && name[i] <= 'Z') {
			modifier := strings.ToLower
			if upCaseNext {
				modifier = strings.ToUpper
			}
			result[j] = modifier(string(name[i]))[0]
			j++
			upCaseNext = false
		} else if name[i] >= '0' && name[i] <= '9' {
			result[j] = name[i]
			upCaseNext = true
			j++
		} else {
			upCaseNext = true
		}
	}
	return string(result)[0:j]
}

func updateMainMigrationFile(m *MigrationFile) error {
	migrationFilePath := path.Join(config.LocalMigratorPath, config.MainMigrationFile)
	f, err := parser.ParseFile(token.NewFileSet(), migrationFilePath, nil, parser.ParseComments)
	if err != nil {
		panic("Couldn't open " + migrationFilePath + ": " + err.Error())
	}

	v := visitor{f: f}
	ast.Walk(&v, f)

	if v.RunPos.IsValid() {
		source, err := ReadFile(migrationFilePath)
		if err != nil {
			panic("Can't read main config file `" + migrationFilePath + "`: " + err.Error())
		}

		newSource := bytes.Buffer{}
		newSource.Write(source[0:v.RunPos])
		newSource.WriteString("mig.Register(")
		if config.LocalMigrationsPackage != "" {
			newSource.WriteString(config.LocalMigrationsPackage)
			newSource.WriteString(".")
		}
		newSource.WriteString(m.FuncName)
		newSource.WriteString("())\n\t")
		newSource.Write(source[v.RunPos:])

		err = WriteFile(migrationFilePath, newSource.Bytes())
		if err != nil {
			panic("Can't write main config file `" + migrationFilePath + "`: " + err.Error())
		}
	} else {
		panic("Couldn't find `mig.Run()` in the migrator file: " + migrationFilePath)
	}

	return nil
}

type visitor struct {
	f      *ast.File
	RunPos token.Pos
}

// Visit here specifically searches for "mig.Run" and writes the position to visitor.RunPos.
func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	value := reflect.ValueOf(node).Elem()
	if value.Type().String() == "ast.SelectorExpr" {
		selector := value.Addr().Interface().(*ast.SelectorExpr)
		if selector.Sel.Name == "Run" {
			value2 := reflect.ValueOf(selector.X).Elem()
			if value2.Addr().Interface().(*ast.Ident).Name == "mig" {
				// found the right insertion point.
				v.RunPos = selector.Pos() - 1
				return nil
			}
		}
	}

	return v
}
