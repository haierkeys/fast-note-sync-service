package util

import (
	"reflect"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name               string
		content            string
		wantYaml           map[string]interface{}
		wantBody           string
		wantHasFrontmatter bool
	}{
		{
			name: "Standard LF",
			content: "---\ntags: [work]\n---\nHello",
			wantYaml: map[string]interface{}{"tags": []interface{}{"work"}},
			wantBody: "Hello",
			wantHasFrontmatter: true,
		},
		{
			name: "Standard CRLF",
			content: "---\r\ntags: [work]\r\n---\r\nHello",
			wantYaml: map[string]interface{}{"tags": []interface{}{"work"}},
			wantBody: "Hello",
			wantHasFrontmatter: true,
		},
		{
			name: "Mixed Line Endings",
			content: "---\r\ntags: [work]\n---\r\nHello",
			wantYaml: map[string]interface{}{"tags": []interface{}{"work"}},
			wantBody: "Hello",
			wantHasFrontmatter: true,
		},
		{
			name: "No Frontmatter",
			content: "Hello World",
			wantYaml: nil,
			wantBody: "Hello World",
			wantHasFrontmatter: false,
		},
		{
			name: "Empty content",
			content: "",
			wantYaml: nil,
			wantBody: "",
			wantHasFrontmatter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYaml, gotBody, gotHasFrontmatter := ParseFrontmatter(tt.content)
			if gotHasFrontmatter != tt.wantHasFrontmatter {
				t.Errorf("ParseFrontmatter() gotHasFrontmatter = %v, want %v", gotHasFrontmatter, tt.wantHasFrontmatter)
			}
			if !reflect.DeepEqual(gotYaml, tt.wantYaml) {
				t.Errorf("ParseFrontmatter() gotYaml = %v, want %v", gotYaml, tt.wantYaml)
			}
			if gotBody != tt.wantBody {
				t.Errorf("ParseFrontmatter() gotBody = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
