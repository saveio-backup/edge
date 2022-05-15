package dsp

import "testing"

func TestIsDirEmpty(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			args: args{name: "/Users/smallyu/work/test/temp"},
			want: false,
		},
		{
			args: args{name: "/Users/smallyu/work/test/temp/empty"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsDirEmpty(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDirEmpty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsDirEmpty() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddPrefixToFile(t *testing.T) {
	prefix := "test"
	prefixBuf := []byte(prefix)
	output := "/Users/smallyu/work/test/file/aaa.test"
	origin := "/Users/smallyu/work/test/file/aaa"
	err := AddPrefixToFile(prefixBuf, output, origin)
	if err != nil {
		t.Error(err)
	}
}
