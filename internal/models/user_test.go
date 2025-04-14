package models

import "testing"

func TestIsUserRole(t *testing.T) {
	type args struct {
		role UserRole
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test IsUserRole",
			args: args{
				role: RoleEmployee,
			},
			want: true,
		},
		{
			name: "Test not IsUserRole",
			args: args{
				role: "employeer",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUserRole(tt.args.role); got != tt.want {
				t.Errorf("IsUserRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
