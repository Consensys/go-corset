package constraint

type Vanishing struct {
	Handle string
	Domain Domain
	Expr   any
}

func (c *Vanishing) GetHandle() string {
	return c.Handle
}
