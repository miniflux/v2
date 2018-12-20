package media

import (
	"reflect"
	"testing"

	"miniflux.app/model"
)

var baseURL = "http://miniflux.app"
var entry = &model.Entry{
	URL: "http://sample.org/articles/1000.htm",
	Content: `
		<body>
		<img src="http://sample.org/logo.png">
		<img src="/images/pic.jpg">
		<img src="../images/pic.jpg">
		</body>
	`,
}

var urlParsed = []string{
	"http://sample.org/logo.png",
	"http://sample.org/images/pic.jpg",
	"http://sample.org/images/pic.jpg",
}

func TestParseDocument(t *testing.T) {
	type args struct {
		entry *model.Entry
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{"test", args{entry}, urlParsed, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDocument(tt.args.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDocument() = %v, want %v", got, tt.want)
			}
		})
	}
}
