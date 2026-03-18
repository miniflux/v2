// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package mediaproxy // import "miniflux.app/v2/internal/mediaproxy"

import (
	"os"
	"testing"

	"miniflux.app/v2/internal/config"
)

func TestRewriteDocumentWithRelativeProxyURL_None_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_None_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="http://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><audio src="http://website/folder/audio.mp3"></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_None_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="http://website/folder/video.mp4"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_None_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_None_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_None_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="http://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><audio src="http://website/folder/audio.mp3"></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_None_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://website/folder/video.mp4"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_None_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "none")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_HttpOnly_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_HttpOnly_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="http://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><audio src="/proxy/t5HoIOMfOlUs1_lCnhvaMI0sUz2_-gqWs_DyRevKIG0=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2F1ZGlvLm1wMw=="></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_HttpOnly_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="/proxy/lKmvyYMkjI4iV7yxQqcYwJHWzMvJmjJZKl7VASyxEZ8=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3ZpZGVvLm1wNA=="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_HttpOnly_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="/proxy/lKmvyYMkjI4iV7yxQqcYwJHWzMvJmjJZKl7VASyxEZ8=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3ZpZGVvLm1wNA==" poster="/proxy/YEEe0bAqTYpNrLijb25xYUNRFQsTPv5LlBikuDScPuo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3Bvc3Rlci5wbmc="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_HttpOnly_Image_Poster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="http://website/folder/video.mp4" poster="/proxy/YEEe0bAqTYpNrLijb25xYUNRFQsTPv5LlBikuDScPuo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3Bvc3Rlci5wbmc="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_HttpOnly_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="http://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><img src="http://localhost/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_HttpOnly_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="http://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><audio src="http://localhost/proxy/t5HoIOMfOlUs1_lCnhvaMI0sUz2_-gqWs_DyRevKIG0=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2F1ZGlvLm1wMw=="></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_HttpOnly_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://localhost/proxy/lKmvyYMkjI4iV7yxQqcYwJHWzMvJmjJZKl7VASyxEZ8=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3ZpZGVvLm1wNA=="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_HttpOnly_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="http://website/folder/video.mp4" poster="http://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://localhost/proxy/lKmvyYMkjI4iV7yxQqcYwJHWzMvJmjJZKl7VASyxEZ8=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3ZpZGVvLm1wNA==" poster="http://localhost/proxy/YEEe0bAqTYpNrLijb25xYUNRFQsTPv5LlBikuDScPuo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL3Bvc3Rlci5wbmc="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_All_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><img src="/proxy/LdPNR1GBDigeeNp2ArUQRyZsVqT_PWLfHGjYFrrWWIY=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_All_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="https://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><audio src="/folder/proxy/EmBTvmU5B17wGuONkeknkptYopW_Tl6Y6_W8oYbN_Xs=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9hdWRpby5tcDM="></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_All_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_All_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4" poster="https://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ=" poster="/proxy/u-yLZEYDELx9OlU9to8bt13iysttOWfYpqRfmQYkm3U=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9wb3N0ZXIucG5n"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_All_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><img src="http://localhost/proxy/LdPNR1GBDigeeNp2ArUQRyZsVqT_PWLfHGjYFrrWWIY=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_All_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="https://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><audio src="http://localhost/proxy/EmBTvmU5B17wGuONkeknkptYopW_Tl6Y6_W8oYbN_Xs=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9hdWRpby5tcDM="></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_All_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://localhost/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_All_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4" poster="https://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://localhost/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ=" poster="http://localhost/proxy/u-yLZEYDELx9OlU9to8bt13iysttOWfYpqRfmQYkm3U=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9wb3N0ZXIucG5n"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_BasePath_All_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><img src="/folder/proxy/LdPNR1GBDigeeNp2ArUQRyZsVqT_PWLfHGjYFrrWWIY=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_BasePath_All_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="https://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><audio src="/folder/proxy/EmBTvmU5B17wGuONkeknkptYopW_Tl6Y6_W8oYbN_Xs=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9hdWRpby5tcDM="></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_BasePath_All_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="/folder/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_BasePath_All_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4" poster="https://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><video src="/folder/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ=" poster="/folder/proxy/u-yLZEYDELx9OlU9to8bt13iysttOWfYpqRfmQYkm3U=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9wb3N0ZXIucG5n"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_BasePath_All_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><img src="http://example.org:88/folder/proxy/LdPNR1GBDigeeNp2ArUQRyZsVqT_PWLfHGjYFrrWWIY=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_BasePath_All_Audio(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "audio")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><audio src="https://website/folder/audio.mp3"></audio></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><audio src="http://example.org:88/folder/proxy/EmBTvmU5B17wGuONkeknkptYopW_Tl6Y6_W8oYbN_Xs=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9hdWRpby5tcDM="></audio></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_BasePath_All_Video(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://example.org:88/folder/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ="></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_BasePath_All_VideoPoster(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "video")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("BASE_URL", "http://example.org:88/folder/")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><video src="https://website/folder/video.mp4" poster="https://website/folder/poster.png"></video></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><video src="http://example.org:88/folder/proxy/rg7VlAFvCFDe4kxg3YJRgtty6AblMwBVGXsn0WWl89k=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci92aWRlby5tcDQ=" poster="http://example.org:88/folder/proxy/u-yLZEYDELx9OlU9to8bt13iysttOWfYpqRfmQYkm3U=/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9wb3N0ZXIucG5n"></video></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithRelativeProxyURL_CustomMediaProxy_All_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("MEDIA_PROXY_CUSTOM_URL", "https://proxy-example/proxy")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)
	expected := `<p><img src="https://proxy-example/proxy/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestRewriteDocumentWithAbsoluteProxyURL_CustomMediaProxy_All_Image(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")
	os.Setenv("MEDIA_PROXY_CUSTOM_URL", "https://proxy-example/proxy")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Config parsing failure: %v`, err)
	}

	input := `<p><img src="https://website/folder/image.png" alt="Test"/></p>`
	output := RewriteDocumentWithAbsoluteProxyURL(input)
	expected := `<p><img src="https://proxy-example/proxy/aHR0cHM6Ly93ZWJzaXRlL2ZvbGRlci9pbWFnZS5wbmc=" alt="Test"/></p>`

	if expected != output {
		t.Errorf(`Not expected output: got %q instead of %q`, output, expected)
	}
}

func TestMediaProxyWithIncorrectCustomMediaProxy(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_CUSTOM_URL", "http://:8080example.com")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err == nil {
		t.Fatalf(`Incorrect proxy URL silently accepted (MEDIA_PROXY_CUSTOM_URL=%q): %q`, os.Getenv("MEDIA_PROXY_CUSTOM_URL"), config.Opts.MediaCustomProxyURL())
	}
}

func TestMediaProxyFilterWithImageSrcset(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<p><img src="http://website/folder/image.png" srcset="http://website/folder/image2.png 656w, http://website/folder/image3.png 360w" alt="test"></p>`
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" srcset="/proxy/aY5Hb4urDnUCly2vTJ7ExQeeaVS-52O7kjUr2v9VrAs=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, /proxy/QgAmrJWiAud_nNAsz3F8OTxaIofwAiO36EDzH_YfMzo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMy5wbmc= 360w" alt="test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestMediaProxyFilterWithEmptySrcset(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<p><img src="http://website/folder/image.png" srcset="" alt="test"></p>`
	expected := `<p><img src="/proxy/okK5PsdNY8F082UMQEAbLPeUFfbe2WnNfInNmR9T4WA=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlLnBuZw==" srcset="" alt="test"/></p>`
	output := RewriteDocumentWithRelativeProxyURL(input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterWithPictureSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<picture><source srcset="http://website/folder/image2.png 656w,   http://website/folder/image3.png 360w, https://website/some,image.png 2x"></picture>`
	expected := `<picture><source srcset="/proxy/aY5Hb4urDnUCly2vTJ7ExQeeaVS-52O7kjUr2v9VrAs=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, /proxy/QgAmrJWiAud_nNAsz3F8OTxaIofwAiO36EDzH_YfMzo=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMy5wbmc= 360w, /proxy/ZIw0hv8WhSTls5aSqhnFaCXlUrKIqTnBRaY0-NaLnds=/aHR0cHM6Ly93ZWJzaXRlL3NvbWUsaW1hZ2UucG5n 2x"/></picture>`
	output := RewriteDocumentWithRelativeProxyURL(input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestProxyFilterOnlyNonHTTPWithPictureSource(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "http-only")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")
	os.Setenv("MEDIA_PROXY_PRIVATE_KEY", "test")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<picture><source srcset="http://website/folder/image2.png 656w, https://website/some,image.png 2x"></picture>`
	expected := `<picture><source srcset="/proxy/aY5Hb4urDnUCly2vTJ7ExQeeaVS-52O7kjUr2v9VrAs=/aHR0cDovL3dlYnNpdGUvZm9sZGVyL2ltYWdlMi5wbmc= 656w, https://website/some,image.png 2x"/></picture>`
	output := RewriteDocumentWithRelativeProxyURL(input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestMediaProxyWithImageDataURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<img src="data:image/gif;base64,test">`
	expected := `<img src="data:image/gif;base64,test"/>`
	output := RewriteDocumentWithRelativeProxyURL(input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestMediaProxyWithImageSourceDataURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("MEDIA_PROXY_MODE", "all")
	os.Setenv("MEDIA_PROXY_RESOURCE_TYPES", "image")

	var err error
	parser := config.NewConfigParser()
	config.Opts, err = parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	input := `<picture><source srcset="data:image/gif;base64,test"/></picture>`
	expected := `<picture><source srcset="data:image/gif;base64,test"/></picture>`
	output := RewriteDocumentWithRelativeProxyURL(input)

	if expected != output {
		t.Errorf(`Not expected output: got %s`, output)
	}
}

func TestShouldProxifyURLWithMimeType(t *testing.T) {
	testCases := []struct {
		name                    string
		mediaURL                string
		mediaMimeType           string
		mediaProxyOption        string
		mediaProxyResourceTypes []string
		expected                bool
	}{
		{
			name:                    "Empty URL should not be proxified",
			mediaURL:                "",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "Data URL should not be proxified",
			mediaURL:                "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
			mediaMimeType:           "image/png",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "HTTP URL with all mode and matching MIME type should be proxified",
			mediaURL:                "http://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                true,
		},
		{
			name:                    "HTTPS URL with all mode and matching MIME type should be proxified",
			mediaURL:                "https://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                true,
		},
		{
			name:                    "HTTP URL with http-only mode and matching MIME type should be proxified",
			mediaURL:                "http://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "http-only",
			mediaProxyResourceTypes: []string{"image"},
			expected:                true,
		},
		{
			name:                    "HTTPS URL with http-only mode should not be proxified",
			mediaURL:                "https://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "http-only",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "URL with none mode should not be proxified",
			mediaURL:                "http://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "none",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "URL with matching MIME type should be proxified",
			mediaURL:                "http://example.com/video.mp4",
			mediaMimeType:           "video/mp4",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"video"},
			expected:                true,
		},
		{
			name:                    "URL with non-matching MIME type should not be proxified",
			mediaURL:                "http://example.com/video.mp4",
			mediaMimeType:           "video/mp4",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "URL with multiple resource types and matching MIME type should be proxified",
			mediaURL:                "http://example.com/audio.mp3",
			mediaMimeType:           "audio/mp3",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image", "audio", "video"},
			expected:                true,
		},
		{
			name:                    "URL with multiple resource types but non-matching MIME type should not be proxified",
			mediaURL:                "http://example.com/document.pdf",
			mediaMimeType:           "application/pdf",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image", "audio", "video"},
			expected:                false,
		},
		{
			name:                    "URL with empty resource types should not be proxified",
			mediaURL:                "http://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{},
			expected:                false,
		},
		{
			name:                    "Relative URL should not be proxified",
			mediaURL:                "/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "Protocol-relative URL should not be proxified",
			mediaURL:                "//cdn.example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "Unsupported scheme should not be proxified",
			mediaURL:                "ftp://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "Blob URL should not be proxified",
			mediaURL:                "blob:https://example.com/123",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                false,
		},
		{
			name:                    "URL with partial MIME type match should be proxified",
			mediaURL:                "http://example.com/image.jpg",
			mediaMimeType:           "image/jpeg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                true,
		},
		{
			name:                    "URL with uppercase MIME type should be proxified",
			mediaURL:                "http://example.com/image.jpg",
			mediaMimeType:           "Image/JPEG",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"image"},
			expected:                true,
		},
		{
			name:                    "URL with audio MIME type and audio resource type should be proxified",
			mediaURL:                "http://example.com/song.ogg",
			mediaMimeType:           "audio/ogg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio"},
			expected:                true,
		},
		{
			name:                    "URL with mixed-case audio MIME type should be proxified",
			mediaURL:                "http://example.com/song.ogg",
			mediaMimeType:           "AuDiO/Ogg",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"audio"},
			expected:                true,
		},
		{
			name:                    "URL with video MIME type and video resource type should be proxified",
			mediaURL:                "http://example.com/movie.webm",
			mediaMimeType:           "video/webm",
			mediaProxyOption:        "all",
			mediaProxyResourceTypes: []string{"video"},
			expected:                true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldProxifyURLWithMimeType(tc.mediaURL, tc.mediaMimeType, tc.mediaProxyOption, tc.mediaProxyResourceTypes)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for URL: %s, MIME type: %s, proxy option: %s, resource types: %v",
					tc.expected, result, tc.mediaURL, tc.mediaMimeType, tc.mediaProxyOption, tc.mediaProxyResourceTypes)
			}
		})
	}
}
