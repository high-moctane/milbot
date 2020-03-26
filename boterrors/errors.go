package boterrors

import (
	"os"
)

// ErrInvalidEnv は環境変数が変な値になっているときのエラーです。
type ErrInvalidEnv struct {
	key string
}

// NewErrInvalidEnv は ErrInvalidEnv を返します。key に環境変数のキーを
// 設定します。
func NewErrInvalidEnv(key string) *ErrInvalidEnv {
	return &ErrInvalidEnv{key: key}
}

// Error はエラーメッセージを返します。環境変数に key がなかった場合は，
// val == None です。
func (e *ErrInvalidEnv) Error() string {
	val, ok := os.LookupEnv(e.key)

	res := "invalid ENV "
	if ok {
		res += "(key: " + e.key + ", val: " + val + ")"
	} else {
		res += "(key: " + e.key + ", val: Null)"
	}

	return res
}
