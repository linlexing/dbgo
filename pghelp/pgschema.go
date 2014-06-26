package pghelp

type PGSchemaDesc struct {
	MetaProject   string
	Grade         string
	DefaultAction string
}
type PGSchema struct {
	Name string
	Desc *PGSchemaDesc
}
