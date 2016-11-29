package levonCache

import "errors"

var (
	ErrKeyNotFound            = errors.New("key not found")
	ErrKeyNotFoundOrLoadTable = errors.New("key not found loaded")
)
