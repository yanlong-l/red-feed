package logger

func String(key, val string) Field {
	return Field{
		Key:   key,
		Value: val,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}

func Int64(key string, value int64) Field {
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
