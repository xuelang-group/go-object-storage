package common

type ErrorProcessor interface {
	Match(e error) bool
	Process(e error) ObjectStorageError
	SetNext(ErrorProcessor)
}

type BaseErrorProcessor struct {
	next ErrorProcessor
}

func (p *BaseErrorProcessor) SetNext(next ErrorProcessor) {
	p.next = next
}

func (p *BaseErrorProcessor) ProcessNext(err error) ObjectStorageError {
	if p.next != nil {
		return p.next.Process(err)
	}
	return nil
}
