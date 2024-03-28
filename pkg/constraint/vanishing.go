package constraint

type Vanishing struct {
	Handle string
	Domain Domain
	Expr interface{}
}

func (c Vanishing) GetHandle() string {
	return c.Handle
}
