package logger

type Field struct {
	Key   string
	Value any
}

func String(key string, value string) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}
