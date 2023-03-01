package server

import (
	"reflect"
	"testing"
)

func TestMux_findRoute(t *testing.T) {
	tests := []struct {
		name  string
		found bool
	}{
		{
			name:  "/services/admin/",
			found: true,
		},
		{
			name:  "/services/admin/users",
			found: true,
		},
		{
			name:  "/services",
			found: false,
		},
		{
			name:  "/",
			found: false,
		},
		{
			name:  "/services/admin/cache",
			found: true,
		},
	}

	// Set up the test routes
	m := NewRouter("testing")
	m.NewRoute("/services/admin/users/", nil)
	m.NewRoute("/services/admin/cache/", nil)
	m.NewRoute("/services/admin/", nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.findRoute(tt.name, "")
			if (got != nil) != tt.found {
				t.Errorf("Mux.findRoute() = %v, want %v", got != nil, tt.found)
			}
		})
	}
}

func TestRoute_makeMap(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    map[string]interface{}
	}{
		{
			path:    "/services/sample",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": false, "name": "", "field": ""},
		},
		{
			path:    "/services/sample/users",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": true, "name": "", "field": ""},
		},
		{
			path:    "/services/sample/users/",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": true, "name": "", "field": ""},
		},
		{
			path:    "/services/sample/users/mary",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": true, "name": "mary", "field": ""},
		},
		{
			path:    "/services/sample/users/mary/",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": true, "name": "mary", "field": ""},
		},
		{
			path:    "/services/sample/users/mary/age",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": true, "name": "mary", "field": "age"},
		},
		{
			path:    "/services/sample/users/mary/age/",
			pattern: "/services/sample/users/{{name}}/{{field}}",
			want:    map[string]interface{}{"services": true, "sample": true, "users": true, "name": "mary", "field": "age"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			r := &Route{
				pattern: tt.pattern,
			}
			if got := r.makeMap(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Route.makeMap() = %v, want %v", got, tt.want)
			}
		})
	}
}