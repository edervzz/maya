package maya

/*
Marker interface to indicate an entity with auditable fields.

  - created_by (string): field to save user who create registry (read from context.Context.Value("user"))
  - created_at (string/datetime): timestamp when registry is created
  - created_by (string): field to save user who update registry (read from context.Context.Value("user"))
  - created_at (string/datetime): timestamp when registry is updated
*/
type IAuditable interface {
}
