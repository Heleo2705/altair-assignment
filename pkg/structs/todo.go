package structs

type TodoItem struct {
	Id        string
	Item      string
	ItemOrder int
}

type TodoItemList struct {
	Items []TodoItem
	Count int
}
