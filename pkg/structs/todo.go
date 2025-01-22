package structs

type TodoItem struct {
	Id    string
	Item  string
	Order int
}

type TodoItemList struct {
	Items []TodoItem
	Count int
}
