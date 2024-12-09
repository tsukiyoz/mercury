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

func Int(key string, value int) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Int64(key string, value int64) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Int64Slice(key string, value []int64) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Int32(key string, value int32) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Bool(key string, value bool) Field {
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
