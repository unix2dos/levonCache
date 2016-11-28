package levonCache

var (
	ErrKeyNotFound            = error.New("key not found")
	ErrKeyNotFoundOrLoadTable = error.New("key not found loaded")
)
