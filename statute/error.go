package statute

import "fmt"

// 返回了一个错误功能码
func newReturnedAbnormalFuncCode(funcCode byte) *ReturnedAbnormalFuncCode {
	return &ReturnedAbnormalFuncCode{funcCode: funcCode}
}

var _ error = (*ReturnedAbnormalFuncCode)(nil)

// ReturnedAbnormalFuncCode 返回了一个错误功能码
type ReturnedAbnormalFuncCode struct {
	funcCode byte
}

func (r *ReturnedAbnormalFuncCode) Error() string {
	return fmt.Sprintf("returned abnormal function code:%d", r.funcCode)
}

func (r *ReturnedAbnormalFuncCode) GetFuncCode() byte {
	return r.funcCode
}
