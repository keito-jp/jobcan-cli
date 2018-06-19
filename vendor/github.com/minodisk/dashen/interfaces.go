package dashen

type Callback func()

type MACCallbacksMap map[string][]Callback

type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}
