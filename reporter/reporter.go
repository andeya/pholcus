package reporter

type Reporter interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	send(string)
	Stop()
	Run()
}
