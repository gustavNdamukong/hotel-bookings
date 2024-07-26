package forms

type errors map[string][]string

// Add adds an error message for a given form field
func (e errors) Add(field, message string) {
	//crete an errors slice & add (append) errors to it as they occur
	e[field] = append(e[field], message)
}

// Get returns first error message for a field
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}

	//return the first index on that error string
	return es[0]
}
