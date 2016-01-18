package generator

import "testing"

func TestEnqueue(t *testing.T) {
	queue := GroupQueue{}

	group := &Group{DesiredSize: 10}
	queue.Enqueue(group)
	if queue.head.value != group || queue.tail.value != group {
		t.Fatalf("queue failed to add a group when empty")
	}

	secondGroup := &Group{DesiredSize: 20}
	queue.Enqueue(secondGroup)
	if queue.head.value != group || queue.tail.value != secondGroup {
		t.Fatalf("queue failed to add a group when non-empty")
	}
}

func TestDequeue(t *testing.T) {
	group := &Group{DesiredSize: 10}
	secondGroup := &Group{DesiredSize: 20}
	queue := GroupQueue{
		head: &groupNode{
			value: group,
			next: &groupNode{
				value: secondGroup,
				next:  nil,
			},
		},
	}

	if actual, expected := queue.Dequeue(), group; actual != expected {
		t.Fatalf("failed to dequeue item correctly,\n\texpected:\n\t%v\n\tgot:\n\t%v", actual, expected)
	}

	if actual, expected := queue.Dequeue(), secondGroup; actual != expected {
		t.Fatalf("failed to dequeue second item correctly,\n\texpected:\n\t%v\n\tgot:\n\t%v", actual, expected)
	}

	if actual := queue.Dequeue(); actual != nil {
		t.Fatalf("failed to dequeue last item correctly,\n\texpected:\n\t%v\n\tgot:\n\t%v", actual, nil)
	}

}

func TestPeek(t *testing.T) {
	queue := GroupQueue{}

	group := &Group{DesiredSize: 10}
	queue.Enqueue(group)
	if actual, expected := queue.Peek(), group; actual != expected {
		t.Fatalf("failed to peek correctly")
	}

	if actual, expected := queue.head.value, group; actual != expected {
		t.Fatalf("peek changed head value,\n\texpected:\n\t%v\n\tgot:\n\t%v", actual, expected)
	}
}
