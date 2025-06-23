package tasks

type TaskResult interface {
	Data() interface{}
}

type BaseResult struct {
	baseData interface{}
}

func (r *BaseResult) Data() interface{} {
	return r.baseData
}

type DownloadResult struct {
	BaseResult
	Name     string
	Bytes    []byte
	FileType string
}
