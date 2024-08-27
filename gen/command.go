package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
	"unicode"
)

func IsValidModuleName(name string) error {
	if !IsSnakeCase(name) {
		return fmt.Errorf("IsValidModuleName(%q): module name must be in snake_case", name)
	}

	if err := ModuleFileExists(name); err != nil {

		return err
	}

	if err := ModuleFileHasCommandDispatcher(name); err != nil {
		return err
	}

	return nil
}

func ModuleFileExists(name string) error {
	if _, err := os.Stat(name + ".go"); err != nil {
		return fmt.Errorf("ModuleFileExists(%q): %w", name+".go", err)
	}
	return nil
}

func ModuleFileHasCommandDispatcher(name string) error {
	filename := name + ".go"
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ModuleFileHasCommandDispatcher(%q): %w", filename, err)
	}
	defer f.Close()

	var buf [4096]byte
	n, err := f.Read(buf[:])

	if err != nil {
		return fmt.Errorf("ModuleFileHasCommandDispatcher(%q): %w", filename, err)
	}

	if !bytes.Contains(buf[:n], []byte("HandleCommand(cmd Command) error")) {
		return fmt.Errorf("ModuleFileHasCommandDispatcher(%q): missing HandleCommand method", filename)
	}
	return nil
}

func IsValidCommandName(name string) error {
	if !IsCamelCase(name) {
		return fmt.Errorf("IsValidCommandName(%q): command name must be in CamelCase", name)
	}
	return nil
}

func IsCamelCase(name string) bool {
	for i, r := range name {
		if i == 0 && !unicode.IsUpper(r) {
			return false
		}
		if i > 0 && !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func IsSnakeCase(name string) bool {
	for i, r := range name {
		if i == 0 && !unicode.IsLower(r) {
			return false
		}
		if i > 0 && !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func must(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const commandTemplate = `package main

import (
	"errors"
	"fmt"
	"time"
)

type {{.CommandName}} struct {
}

func (cmd *{{.CommandName}}) CommandName() string { return "{{.CommandName}}" }

func init() {
	DefaultCommandRegistry.Register("{{.CommandName}}", func() Command { return new({{.CommandName}}) })
}

func (self *{{.ModuleReceiver}}) handle{{.CommandName}}(cmd *{{.CommandName}}) error {
  return errors.New("not implemented")
}
`

func RenderCommandTemplate(buf *bytes.Buffer, templateData map[string]any) error {
	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		return fmt.Errorf("RenderCommandTemplate: %w", err)
	}
	if err := tmpl.Execute(buf, templateData); err != nil {
		return fmt.Errorf("RenderCommandTemplate: %w", err)
	}
	return nil
}

func ToTitleCase(name string) string {
	return string(unicode.ToUpper(rune(name[0]))) + name[1:]
}

func ToSnakeCase(name string) string {
	var buf bytes.Buffer
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func WriteCommandFile(filename string, buf *bytes.Buffer) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("WriteCommandFile(%q): %w", filename, err)
	}
	defer f.Close()
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("WriteCommandFile(%q): %w", filename, err)
	}
	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: command <module-name><command-name>")
		os.Exit(1)
	}
	moduleName := os.Args[1]
	commandName := os.Args[2]

	must(IsValidModuleName(moduleName))
	must(IsValidCommandName(commandName))

	buf := bytes.Buffer{}
	templateData := map[string]any{
		"ModuleName":     moduleName,
		"ModuleReceiver": ToTitleCase(moduleName),
		"CommandName":    commandName,
	}
	must(RenderCommandTemplate(&buf, templateData))

	filename := fmt.Sprintf("./%s_%s.go", moduleName, ToSnakeCase(commandName))

	must(WriteCommandFile(filename, &buf))
	fmt.Println(filename)
}
