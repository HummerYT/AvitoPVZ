package models

import "testing"

func TestIsPVZCity(t *testing.T) {
	type args struct {
		city PVZCity
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "true",
			args: args{
				city: CityMoscow,
			},
			want: true,
		},
		{
			name: "false",
			args: args{
				city: "Moscow",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPVZCity(tt.args.city); got != tt.want {
				t.Errorf("IsPVZCity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTypeProduct(t *testing.T) {
	type args struct {
		product string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "true",
			args: args{
				product: "обувь",
			},
			want: true,
		},
		{
			name: "false",
			args: args{
				product: "product",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTypeProduct(tt.args.product); got != tt.want {
				t.Errorf("IsTypeProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}
