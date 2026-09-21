package mysql

import "testing"

func TestWithClientFoundRows(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		want    string
		wantErr bool
	}{
		{
			name: "clientFoundRowsが付与される",
			dsn:  "user:pass@tcp(localhost:3306)/test?parseTime=true",
			want: "user:pass@tcp(localhost:3306)/test?clientFoundRows=true&parseTime=true",
		},
		{
			name: "既にclientFoundRows=falseが指定されていてもtrueに上書きされる",
			dsn:  "user:pass@tcp(localhost:3306)/test?clientFoundRows=false&parseTime=true",
			want: "user:pass@tcp(localhost:3306)/test?clientFoundRows=true&parseTime=true",
		},
		{
			name:    "不正なDSNはエラーになる",
			dsn:     "not-a-valid-dsn",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := withClientFoundRows(tc.dsn)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("withClientFoundRows() error = %v", err)
			}
			if got != tc.want {
				t.Errorf("withClientFoundRows() = %q, want %q", got, tc.want)
			}
		})
	}
}
