package flag

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
)

// 用户自定义一个Value类型的变量,转换为Flag类型
type Value interface {
	String() string
	Set(string) error
}

type Getter interface {
	Value
	Get() interface{}
}

type Flag struct {
	Name     string
	Usage    string
	Value    Value
	DefValue string
}

type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota
	ExitOnError
	PanicOnError
)

// 一个Command的Flag抽象
type FlagSet struct {
	// Invoked when parse failed,wrapped by usage()
	// When FlagSet.Usage() is nil,DefaultPrints will be called
	// We can change this behavior by redefine FlagSet.Usage() or var Usage()
	// Once we changed FlagSet.Usage(),the behavior can not be changed by redefine Usage()
	// So we can have a dynamic behavior by redefine Usage() before FlagSet.Usage() be set explicit
	// After the package being imported,FlagSet.Usage() was pointed to DefaultPrints
	// CommandLine.Usage()->commandLineUsage()->Usage()->PrintDefaults
	// We can change FlagSet's Usage by redefine FlagSet.Usage().CommandLine is also a sort of FlagSet,so
	// we can change it's Usage by redefine CommandLine'Usage,but more generally,we change the var Usage() to do this
	Usage func()

	name          string
	parsed        bool
	actual        map[string]*Flag
	formal        map[string]*Flag
	args          []string
	errorHandling ErrorHandling
	output        io.Writer
}

func (f *FlagSet) Name() string {
	return f.name
}

func (f *FlagSet) SetOutput(output io.Writer) {
	f.output = output
}

func (f *FlagSet) Output() io.Writer {
	if f.output == nil {
		return os.Stderr
	}
	return f.output
}

func PrintDefaults() {
	CommandLine.PrintDefaults()
}

func (f *FlagSet) PrintDefaults() {
	f.VisitAll(func(flag *Flag) {
		var b strings.Builder
		fmt.Fprintf(&b, "  -%s", flag.Name)
		name, usage := UnquoteUsage(flag)
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}
		if b.Len() <= 4 {
			b.WriteString("\t")
		} else {
			b.WriteString("\n    \t")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))
		if !isZeroValue(flag, flag.DefValue) {
			if _, ok := flag.Value.(*stringValue); ok {
				fmt.Fprintf(&b, " (default %q)", flag.DefValue)
			} else {
				fmt.Fprintf(&b, " (defaule %v)", flag.DefValue)
			}
		}
		fmt.Fprint(f.Output(), b.String(), "\n")
	})
}

func isZeroValue(flag *Flag, value string) bool {
	typ := reflect.TypeOf(flag.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	return value == z.Interface().(Value).String()
}

func UnquoteUsage(flag *Flag) (name string, usage string) {
	usage = flag.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return
				}
			}
			break
		}
	}
	name = "value"
	switch flag.Value.(type) {
	case boolFlag:
		name = ""
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}
	return
}

func (f *FlagSet) defaultUsage() {
	if f.name == "" {
		fmt.Fprintf(f.Output(), "Usage:\n")
	} else {
		fmt.Fprintf(f.Output(), "Usage of %s:\n", f.name)
	}
	f.PrintDefaults()
}

func NewFlagSet(name string, errorHandling ErrorHandling) *FlagSet {
	f := &FlagSet{
		name:          name,
		errorHandling: errorHandling,
	}
	f.Usage = f.defaultUsage
	return f
}

func (f *FlagSet) parseOne() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}
	s := f.args[0]
	if len(s) < 2 || s[0] != '-' {
		return false, nil
	}
	numMinuses := 1
	if s[1] == '-' {
		numMinuses++
		if len(s) == 2 {
			f.args = f.args[1:]
			return false, nil
		}
	}
	name := s[numMinuses:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		f.args = f.args[1:]
		return false, f.failf("bad flag syntax:%s", s)
	}
	// has a flag
	f.args = f.args[1:]
	hasValue := false
	value := ""
	// 如果-name=value格式则将name切割,否则value=""
	for i := 1; i < len(name); i++ {
		if name[i] == '=' {
			hasValue = true
			value = name[i+1:]
			name = name[0:i]
			break
		}
	}
	m := f.formal
	flag, alreadythere := m[name]
	// 不存在flag时有一种特殊情况就是help,打印Usage
	if !alreadythere {
		if name == "help" || name == "h" {
			f.usage()
			return false, ErrHelp
		}
		return false, f.failf("flag provided but not defined: -%s", name)
	}
	if fv, ok := flag.Value.(boolFlag); ok && fv.IsBoolFlag() {
		if hasValue {
			if err := fv.Set(value); err != nil {
				return false, f.failf("invaild bool value %q for -%s:%v", value, name, err)
			}
		} else {
			if err := fv.Set(value); err != nil {
				return false, f.failf("invaild bool flag %s:%v", name, err)
			}
		}
	} else {
		if !hasValue && len(f.args) > 0 {
			hasValue = true
			value, f.args = f.args[0], f.args[1:]
		}
		if !hasValue {
			return false, f.failf("flag needs an argument:-%s", name)
		}
		if err := fv.Set(value); err != nil {
			return false, f.failf("invaild bool value %q for -%s:%v", value, name, err)
		}
		if f.actual == nil {
			f.actual = make(map[string]*Flag)
		}
	}
	f.actual[name] = flag
	return true, nil
}

// translate a string slice has format like :
// -f1 arg1 --f2 arg2 -f3=arg3 --f4=arg4 -f5 --f6
// unlike Parse() use os.Args[1:],this method can parse flag comes from http request or file
func (f *FlagSet) Parse(arguments []string) error {
	f.parsed = true
	f.args = arguments
	for {
		seen, err := f.parseOne()
		if seen == true {
			continue
		}
		if err == nil {
			break // no flag can be parsed
		}
		switch f.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			if err == ErrHelp {
				os.Exit(0)
			}
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

// use commandline as default
func Parse() {
	CommandLine.Parse(os.Args[1:])
}

func sortFlags(flags map[string]*Flag) []*Flag {
	result := make([]*Flag, len(flags))
	i := 0
	for _, flag := range flags {
		result[i] = flag
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func VisitAll(fn func(*Flag)) {
	CommandLine.VisitAll(fn)
}

func (f *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range sortFlags(f.formal) {
		fn(flag)
	}
}

func Visit(fn func(*Flag)) {
	CommandLine.Visit(fn)
}

func (f *FlagSet) Visit(fn func(*Flag)) {
	for _, flag := range sortFlags(f.actual) {
		fn(flag)
	}
}

func (f *FlagSet) failf(format string, a ...interface{}) error {
	msg := f.sprintf(format, a...)
	f.usage()
	return errors.New(msg)
}

// invoked when parse failed
func (f *FlagSet) usage() {
	if f.Usage == nil {
		f.defaultUsage()
	} else {
		f.Usage()
	}
}

func (f *FlagSet) sprintf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
	return msg
}

func (f *FlagSet) Parsed() bool {
	return f.parsed
}

func (f *FlagSet) Init(name string, errorHandling ErrorHandling) {
	f.name = name
	f.errorHandling = errorHandling
}

func Parsed() bool {
	return CommandLine.Parsed()
}

var Usage = func() {
	fmt.Fprintf(CommandLine.Output(), "usage of %s:\n", os.Args[0])
	PrintDefaults()
}

func commandLineUsage() {
	Usage()
}

func init() {
	CommandLine.Usage = commandLineUsage
}

func Var(value Value, name string, usage string) {
	CommandLine.Var(value, name, usage)
}

func (f *FlagSet) Var(value Value, name string, usage string) {
	if strings.HasPrefix(name, "-") {
		panic(f.sprintf("flag %q begins with -", name))
	} else if strings.Contains(name, "=") {
		panic(f.sprintf("flag %q contains =", name))
	}
	flag := &Flag{name, usage, value, value.String()}
	_, alreadythere := f.formal[name]
	if alreadythere {
		var msg string
		if f.name == "" {
			msg = f.sprintf("flag redefined:%s", name)
		} else {
			msg = f.sprintf("%s flag refefined:%s", f.name, name)
		}
		panic(msg)
	}
	if f.formal == nil {
		f.formal = make(map[string]*Flag)
	}
	f.formal[name] = flag
}

var CommandLine = NewFlagSet(os.Args[0], ExitOnError)
