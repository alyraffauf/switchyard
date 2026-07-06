package main

type actionItem struct {
	name string
	exec string
}

type browserItem struct {
	name string
	id   string
	exec string
}

func (i actionItem) FilterValue() string { return i.name }
func (i actionItem) Title() string       { return i.name }
func (i actionItem) Description() string { return "" }

func (i browserItem) FilterValue() string { return i.name }
func (i browserItem) Title() string       { return i.name }
func (i browserItem) Description() string { return i.id }
