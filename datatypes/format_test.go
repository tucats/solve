package datatypes

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want string
	}{
		{
			name: "float",
			arg:  3.14,
			want: "3.14",
		},
		{
			name: "struct type",
			arg:  UserType("bang", StructType),
			want: "T(bang)",
		},
		{
			name: "struct",
			arg:  StructType,
			want: "T(struct)",
		},
		{
			name: "array",
			arg:  NewFromArray(IntType, []interface{}{1, 2, 3}),
			want: "[1, 2, 3]",
		},
		{
			name: "array type",
			arg:  ArrayOfType(IntType),
			want: "T([]int)",
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Format(tt.arg); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
