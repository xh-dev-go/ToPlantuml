package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"github.com/xh-dev-go/xhUtils/flagUtils"
	"github.com/xh-dev-go/xhUtils/flagUtils/flagBool"
	"os"
	"regexp"
	"strings"
)

type FieldType string

const (
	date FieldType = "Date"
	str  FieldType = "string"
	uuid FieldType = "UUID"
	num  FieldType = "number"
)

func (ft FieldType) toString() string {
	return ft.toString()
}

type Table struct {
	name   string
	fields []Field
}
type Field struct {
	name         string
	fieldType    FieldType
	defaultValue string
}

func (field *Field) defaultVal() string {
	switch field.fieldType {
	case date:
		return "\"20220102\""
	case num:
		return "123"
	case str:
		return "\"string\""
	case uuid:
		return "\"8ce77e6f-7d9b-4e5c-a4e0-2ebead7d0f81\""
	}
	panic("No match found")
}

func ToClass(table Table) string {
	indent := Indentation{}
	var line = fmt.Sprintf("class %s{\n", table.name)
	indent.increase()
	for _, field := range table.fields {
		line = line + fmt.Sprintf("%s%s %s\n", indent.pad(), field.fieldType, field.name)
	}
	line = line + "}"
	return line
}

func ToObject(table Table) string {
	indent := Indentation{}
	var line = fmt.Sprintf("object instance_%s{\n", table.name)
	indent.increase()
	for _, field := range table.fields {
		var defVal = ""
		if field.defaultValue == "" {
			defVal = field.defaultVal()
		}
		line = line + fmt.Sprintf("%s%s = %s\n", indent.pad(), field.name, defVal)
	}
	line = line + "}"
	return line
}

func Parse(msg string) ([]Table, error) {
	var ms = make(map[string][]Field)
	splitMsg := strings.Split(msg, "\n")
	l := len(splitMsg) - 1
	for i, m := range splitMsg {
		m = strings.TrimRight(m, "\r")
		if i == l && m == "" {
			continue
		}

		arr := strings.Split(m, "\t")
		if len(arr) < 3 {
			return nil, errors.New("len not match")
		}

		_, ok := ms[arr[0]]
		if !ok {
			ms[arr[0]] = []Field{}
		}
		fields, _ := ms[arr[0]]
		var ft FieldType

		if strings.HasPrefix(strings.ToUpper(arr[2]), "VARCHAR") {
			arr[2] = "VARCHAR"
		}
		reg, err := regexp.Compile("NUMBER\\(\\d+[ ]?,[ ]?0")
		if err != nil {
			panic(err)
		}
		if reg.MatchString(arr[2]) {
			arr[2] = "LONG"
		}
		switch strings.ToUpper(arr[2]) {
		case "RAW(16)":
			ft = uuid
		case "VARCHAR":
			ft = str
		case "DATE":
			ft = date
		case "LONG":
			ft = num
		default:
			panic(fmt.Sprintf("type not match: %v\n", arr))
		}
		ms[arr[0]] = append(fields, Field{name: arr[1], fieldType: ft})
	}

	var tables []Table
	for i, m := range ms {
		tables = append(tables,
			Table{
				name:   i,
				fields: m,
			},
		)
	}
	return tables, nil
}

type Indentation struct {
	level int
}

func (indent *Indentation) increase() {
	indent.level = indent.level + 1
}

func (indent *Indentation) pad() string {
	var l = ""
	for i := 0; i < indent.level; i++ {
		l += " "
	}
	return l
}

func (indent *Indentation) decrease() {
	if indent.level > 0 {
		indent.level = indent.level - 1
	} else {
		panic("negative indentation")
	}
}

const VERSION = "1.0.0"

func main() {
	toClassParam := flagBool.New("to-class", "to class diagram").BindCmd()
	toObjectparam := flagBool.New("to-object", "to class diagram").BindCmd()
	versionParam := flagUtils.Version().BindCmd()

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	if versionParam.Value() {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if !toObjectparam.Value() && !toClassParam.Value() {
		panic("Expect at least to-class or to-object flag is applied")
	}

	msg, err := clipboard.ReadAll()
	if err != nil {
		panic(err)
	}
	tables, err := Parse(msg)
	if err != nil {
		panic(err)
	}

	var line = "@startuml\n"
	for _, table := range tables {
		if toClassParam.Value() {
			line = line + ToClass(table) + "\n"
		}
		if toObjectparam.Value() {
			line = line + ToObject(table) + "\n"
		}
	}
	line = line + "@enduml"

	clipboard.WriteAll(line)

}
