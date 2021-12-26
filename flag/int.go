package flag

import "strconv"

type intValue int

func (i *intValue) String() string {
	return strconv.Itoa(int(*i))
}

func (i *intValue) Get() interface{} {
	return int(*i)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = intValue(v)
	return err
}

func Int(name string, value int, usage string) {
	CommandLine.Int(name, value, usage)
}

func IntVar(p *int, name string, value int, usage string) {
	CommandLine.Var(newIntValue(value, p), name, usage)
}

func (f *FlagSet) Int(name string, value int, usage string) *int {
	p := new(int)
	f.IntVar(p, name, value, usage)
	return p
}

func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.Var(newIntValue(value, p), name, usage)
}

func newIntValue(value int, p *int) *intValue {
	*p = value
	return (*intValue)(p)
}
