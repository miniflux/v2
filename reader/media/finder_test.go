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
var medias = map[string]*model.Media{
	"06bead8615c0dea3bbba488423b7358ef8426768e5635960ef86d55a4c2b4426": &model.Media{
		ID:       1,
		URLHash:  "06bead8615c0dea3bbba488423b7358ef8426768e5635960ef86d55a4c2b4426",
		MimeType: "image/png",
		Content:  []byte{},
	},
	"ef1687527395c23b59b3f89e6eb8a4cf530d8631360bdf12db77bdf2c16cac57": &model.Media{
		ID:       2,
		URLHash:  "ef1687527395c23b59b3f89e6eb8a4cf530d8631360bdf12db77bdf2c16cac57",
		MimeType: "image/jpeg",
		Content:  []byte{},
	},
}
var urlParsed = []string{
	"http://sample.org/logo.png",
	"http://sample.org/images/pic.jpg",
	"http://sample.org/images/pic.jpg",
}
var urlRedirected = []string{
	"http://miniflux.app/media/06bead8615c0dea3bbba488423b7358ef8426768e5635960ef86d55a4c2b4426",
	"http://miniflux.app/media/ef1687527395c23b59b3f89e6eb8a4cf530d8631360bdf12db77bdf2c16cac57",
	"http://miniflux.app/media/ef1687527395c23b59b3f89e6eb8a4cf530d8631360bdf12db77bdf2c16cac57",
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

func TestRedirectMedia(t *testing.T) {
	type args struct {
		entry   *model.Entry
		medias  map[string]*model.Media
		baseURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{"test", args{entry, medias, baseURL}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RedirectMedia(tt.args.entry, tt.args.medias, tt.args.baseURL)
			got, _ := ParseDocument(tt.args.entry)
			if !reflect.DeepEqual(got, urlRedirected) {
				t.Errorf("ParseDocument() = %v, want %v", got, urlRedirected)
			}
		})
	}
}
