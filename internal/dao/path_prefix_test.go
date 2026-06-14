package dao

import "testing"

func TestIsPathWithinPrefix_UsesPathBoundary(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{name: "child", path: "Projects/Archive/todo.md", prefix: "Projects", want: true},
		{name: "exact path is not child", path: "Projects", prefix: "Projects", want: false},
		{name: "sibling prefix", path: "Projects-old/todo.md", prefix: "Projects", want: false},
		{name: "underscore is literal", path: "fooXbar/todo.md", prefix: "foo_bar", want: false},
		{name: "percent is literal", path: "fooooo/bar/todo.md", prefix: "foo%/bar", want: false},
		{name: "trim slashes", path: "/Projects/Archive/todo.md/", prefix: "/Projects/", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPathWithinPrefix(tt.path, tt.prefix); got != tt.want {
				t.Fatalf("isPathWithinPrefix(%q, %q) = %v, want %v", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}
