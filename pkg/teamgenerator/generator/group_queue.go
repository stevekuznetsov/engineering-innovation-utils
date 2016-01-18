package generator

type GroupQueue struct {
	head *groupNode
	tail *groupNode
}

type groupNode struct {
	value *Group
	next  *groupNode
}

func (q *GroupQueue) Enqueue(group *Group) {
	newNode := &groupNode{
		value: group,
		next:  nil,
	}
	if q.tail != nil {
		q.tail.next = newNode
	}
	q.tail = newNode

	if q.head == nil {
		// this is the first element being added
		q.head = q.tail
	}
}

func (q *GroupQueue) Dequeue() *Group {
	if q.head == nil {
		return nil
	}

	oldestNode := q.head
	q.head = oldestNode.next
	return oldestNode.value

}

func (q *GroupQueue) Peek() *Group {
	return q.head.value
}

func (q *GroupQueue) IsEmpty() bool {
	return q.head == nil
}
