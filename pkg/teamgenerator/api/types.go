// api holds data structures serialized to disk
package api

// ClassGrouping is a collection of all groupings for a given class for a given semester
type ClassGrouping struct {
	// Projects is a list of all project groupings for a semester
	Projects []ProjectGrouping `json:"projects"`
}

// ProjectGrouping is a collection of groups that contain all members of a class
type ProjectGrouping struct {
	Name string `json:"name"`
	// Groups hold the grouped students
	Groups []Group `json:"groups"`
}

// Group is a collection of students
type Group struct {
	// Members are the student members of a group
	Members []Student `json:"students"`
}

// Student represents a student in the class
type Student struct {
	FullName string `json:"name"`
	NetID    string `json:"netID"`
}
