package deduce

// Engine abstracts a deductive query engine with a Datalog/Mangle API.
// A future build-tagged implementation can use github.com/google/mangle.
type Engine interface {
	Exec(query string) (string, error)
}

// PlaceholderEngine echoes the query and points to the facts file location.
type PlaceholderEngine struct{ FactsPath string }

func (p *PlaceholderEngine) Exec(query string) (string, error) {
	return "(placeholder) facts=" + p.FactsPath + " query=" + query, nil
}

