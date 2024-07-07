package core

type Output struct {
	callBack     func(msg string)
	respCallBack func(url string, msg string)
}

func NewOutput(callBack func(msg string), respCallBack func(url string, msg string)) *Output {

	return &Output{
		callBack:     callBack,
		respCallBack: respCallBack,
	}
}

func (o *Output) WriteToFile(msg string) {
	o.callBack(msg)
}
