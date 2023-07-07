package maya

type Migration struct {
	ID   string   // ID migration. e.g. 00000001
	Up   []string // slice of DDL for creation
	Down []string // slice of DDL for drop
}

type ByID []Migration

func (a ByID) Len() int           { return len(a) }
func (a ByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
