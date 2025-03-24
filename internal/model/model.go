package model

//"gorm.io/driver/sqlite"

type Predicate string

var (
	// Eq =
	Eq = Predicate("=")
	// Neq <>
	Neq = Predicate("<>")
	// Gt >
	Gt = Predicate(">")
	// Egt >=
	Egt = Predicate(">=")
	// Lt <
	Lt = Predicate("<")
	// Elt <=
	Elt  = Predicate("<=")
	Like = Predicate("LIKE")
)
