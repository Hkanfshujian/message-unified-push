package rbac_service

import "testing"

func TestAppendBuiltinProfilePermissions(t *testing.T) {
	input := []string{"dashboard:view", "profile:settings:view"}
	result := appendBuiltinProfilePermissions(input)

	hasView := false
	hasEdit := false
	hasDashboard := false
	seen := make(map[string]struct{})
	for _, code := range result {
		if _, ok := seen[code]; ok {
			t.Fatalf("unexpected duplicated permission: %s", code)
		}
		seen[code] = struct{}{}
		if code == "profile:settings:view" {
			hasView = true
		}
		if code == "profile:settings:edit" {
			hasEdit = true
		}
		if code == "dashboard:view" {
			hasDashboard = true
		}
	}

	if !hasView {
		t.Fatalf("missing profile:settings:view")
	}
	if !hasEdit {
		t.Fatalf("missing profile:settings:edit")
	}
	if !hasDashboard {
		t.Fatalf("missing existing permission dashboard:view")
	}
}
