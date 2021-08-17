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
			arg:  TypeDefinition("bang", StructType),
			want: "T(bang struct)",
		},
		{
			name: "struct",
			arg:  StructType,
			want: "T(struct)",
		},
		{
			name: "array",
			arg:  NewArrayFromArray(IntType, []interface{}{1, 2, 3}),
			want: "[1, 2, 3]",
		},
		{
			name: "array type",
			arg:  Array(IntType),
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
