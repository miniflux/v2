// Code generated by go generate; DO NOT EDIT.

package template // import "miniflux.app/template"

var templateCommonMap = map[string]string{
	"entry_pagination": `{{ define "entry_pagination" }}
<div class="pagination">
    <div class="pagination-prev">
        {{ if .prevEntry }}
            <a href="{{ .prevEntryRoute }}{{ if .searchQuery }}?q={{ .searchQuery }}{{ end }}" title="{{ .prevEntry.Title }}" data-page="previous">{{ t "pagination.previous" }}</a>
        {{ else }}
            {{ t "pagination.previous" }}
        {{ end }}
    </div>

    <div class="pagination-next">
        {{ if .nextEntry }}
            <a href="{{ .nextEntryRoute }}{{ if .searchQuery }}?q={{ .searchQuery }}{{ end }}" title="{{ .nextEntry.Title }}" data-page="next">{{ t "pagination.next" }}</a>
        {{ else }}
            {{ t "pagination.next" }}
        {{ end }}
    </div>
</div>
{{ end }}`,
	"feed_list": `{{ define "feed_list" }}
    <div class="items">
        {{ range .feeds }}
        <article class="item {{ if ne .ParsingErrorCount 0 }}feed-parsing-error{{ end }}">
            <div class="item-header">
                <span class="item-title">
                    {{ if .Icon }}
                        <img src="{{ route "icon" "iconID" .Icon.IconID }}" width="16" height="16" loading="lazy" alt="{{ .Title }}">
                    {{ end }}
                    {{ if .Disabled }} 🚫 {{ end }}
                    <a href="{{ route "feedEntries" "feedID" .ID }}">{{ .Title }}</a>
                </span>
                <span class="feed-entries-counter">
                    (<span title="{{ t "page.feeds.unread_counter" }}">{{ .UnreadCount }}</span>/<span title="{{ t "page.feeds.read_counter" }}">{{ .ReadCount }}</span>)
                </span>
                <span class="category">
                    <a href="{{ route "categoryEntries" "categoryID" .Category.ID }}">{{ .Category.Title }}</a>
                </span>
            </div>
            <div class="item-meta">
                <ul>
                    <li>
                        <a href="{{ .SiteURL | safeURL  }}" title="{{ .SiteURL }}" target="_blank" rel="noopener noreferrer" referrerpolicy="no-referrer" data-original-link="true">{{ domain .SiteURL }}</a>
                    </li>
                    <li>
                        {{ t "page.feeds.last_check" }} <time datetime="{{ isodate .CheckedAt }}" title="{{ isodate .CheckedAt }}">{{ elapsed $.user.Timezone .CheckedAt }}</time>
                    </li>
                </ul>
                <ul>
                    <li>
                        <a href="{{ route "refreshFeed" "feedID" .ID }}">{{ t "menu.refresh_feed" }}</a>
                    </li>
                    <li>
                        <a href="{{ route "editFeed" "feedID" .ID }}">{{ t "menu.edit_feed" }}</a>
                    </li>
                    <li>
                        <a href="#"
                            data-confirm="true"
                            data-label-question="{{ t "confirm.question" }}"
                            data-label-yes="{{ t "confirm.yes" }}"
                            data-label-no="{{ t "confirm.no" }}"
                            data-label-loading="{{ t "confirm.loading" }}"
                            data-url="{{ route "removeFeed" "feedID" .ID }}">{{ t "action.remove" }}</a>
                    </li>
                </ul>
            </div>
            {{ if ne .ParsingErrorCount 0 }}
                <div class="parsing-error">
                    <strong title="{{ .ParsingErrorMsg }}" class="parsing-error-count">{{ plural "page.feeds.error_count" .ParsingErrorCount .ParsingErrorCount }}</strong>
                    - <small class="parsing-error-message">{{ .ParsingErrorMsg }}</small>
                </div>
            {{ end }}
        </article>
        {{ end }}
    </div>
{{ end }}`,
	"feed_menu": `{{ define "feed_menu" }}
<ul>
    <li>
        <a href="{{ route "feeds" }}">{{ t "menu.feeds" }}</a>
    </li>
    <li>
        <a href="{{ route "addSubscription" }}">{{ t "menu.add_feed" }}</a>
    </li>
    <li>
        <a href="{{ route "export" }}">{{ t "menu.export" }}</a>
    </li>
    <li>
        <a href="{{ route "import" }}">{{ t "menu.import" }}</a>
    </li>
    <li>
        <a href="{{ route "refreshAllFeeds" }}">{{ t "menu.refresh_all_feeds" }}</a>
    </li>
</ul>
{{ end }}`,
	"item_meta": `{{ define "item_meta" }}
<div class="item-meta">
    <ul>
        <li>
            <a href="{{ route "feedEntries" "feedID" .entry.Feed.ID }}" title="{{ .entry.Feed.SiteURL }}">{{ truncate .entry.Feed.Title 35 }}</a>
        </li>
        <li>
            <time datetime="{{ isodate .entry.Date }}" title="{{ isodate .entry.Date }}">{{ elapsed .user.Timezone .entry.Date }}</time>
        </li>
        {{ if .hasSaveEntry }}
            <li>
                <a href="#"
                    title="{{ t "entry.save.title" }}"
                    data-save-entry="true"
                    data-save-url="{{ route "saveEntry" "entryID" .entry.ID }}"
                    data-label-loading="{{ t "entry.state.saving" }}"
                    data-label-done="{{ t "entry.save.completed" }}"
                    >{{ t "entry.save.label" }}</a>
            </li>
        {{ end }}
        <li>
            <a href="{{ .entry.URL | safeURL  }}" target="_blank" rel="noopener noreferrer" referrerpolicy="no-referrer" data-original-link="true">{{ t "entry.original.label" }}</a>
        </li>
        {{ if .entry.CommentsURL }}
            <li>
                <a href="{{ .entry.CommentsURL | safeURL  }}" title="{{ t "entry.comments.title" }}" target="_blank" rel="noopener noreferrer" referrerpolicy="no-referrer" data-comments-link="true">{{ t "entry.comments.label" }}</a>
            </li>
        {{ end }}
        <li>
            <a href="#"
                data-toggle-bookmark="true"
                data-bookmark-url="{{ route "toggleBookmark" "entryID" .entry.ID }}"
                data-label-loading="{{ t "entry.state.saving" }}"
                data-label-star="☆&nbsp;{{ t "entry.bookmark.toggle.on" }}"
                data-label-unstar="★&nbsp;{{ t "entry.bookmark.toggle.off" }}"
                data-value="{{ if .entry.Starred }}star{{ else }}unstar{{ end }}"
                >{{ if .entry.Starred }}★&nbsp;{{ t "entry.bookmark.toggle.off" }}{{ else }}☆&nbsp;{{ t "entry.bookmark.toggle.on" }}{{ end }}</a>
        </li>
        <li>
            <a href="#"
                title="{{ t "entry.status.title" }}"
                data-toggle-status="true"
                data-label-read="✔&#xfe0e;&nbsp;{{ t "entry.status.read" }}"
                data-label-unread="✘&nbsp;{{ t "entry.status.unread" }}"
                data-value="{{ if eq .entry.Status "read" }}read{{ else }}unread{{ end }}"
                >{{ if eq .entry.Status "read" }}✘&nbsp;{{ t "entry.status.unread" }}{{ else }}✔&#xfe0e;&nbsp;{{ t "entry.status.read" }}{{ end }}</a>
        </li>
    </ul>
</div>
{{ end }}
`,
	"layout": `{{ define "base" }}
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>{{template "title" .}} - Miniflux</title>

    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no">
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-title" content="Miniflux">
    <link rel="manifest" href="{{ route "webManifest" }}" crossorigin="use-credentials"/>

    <meta name="robots" content="noindex,nofollow">
    <meta name="referrer" content="no-referrer">
    <meta name="google" content="notranslate">

    <!-- Favicons -->
    <link rel="icon" type="image/png" sizes="16x16" href="{{ route "appIcon" "filename" "favicon-16.png" }}">
    <link rel="icon" type="image/png" sizes="32x32" href="{{ route "appIcon" "filename" "favicon-32.png" }}">

    <!-- Android icons -->
    <link rel="icon" type="image/png" sizes="128x128" href="{{ route "appIcon" "filename" "icon-128.png" }}">
    <link rel="icon" type="image/png" sizes="192x192" href="{{ route "appIcon" "filename" "icon-192.png" }}">

    <!-- iOS icons -->
    <link rel="apple-touch-icon" sizes="120x120" href="{{ route "appIcon" "filename" "icon-120.png" }}">
    <link rel="apple-touch-icon" sizes="152x152" href="{{ route "appIcon" "filename" "icon-152.png" }}">
    <link rel="apple-touch-icon" sizes="167x167" href="{{ route "appIcon" "filename" "icon-167.png" }}">
    <link rel="apple-touch-icon" sizes="180x180" href="{{ route "appIcon" "filename" "icon-180.png" }}">

    {{ if .csrf }}
        <meta name="X-CSRF-Token" value="{{ .csrf }}">
    {{ end }}

    <meta name="theme-color" content="{{ theme_color .theme }}">
    <link rel="stylesheet" type="text/css" href="{{ route "stylesheet" "name" .theme }}?{{ .theme_checksum }}">

    <script type="text/javascript" src="{{ route "javascript" "name" "app" }}?{{ .app_js_checksum }}" defer></script>
    <script type="text/javascript" src="{{ route "javascript" "name" "sw" }}?{{ .sw_js_checksum }}" defer id="service-worker-script"></script>
</head>
<body
    data-entries-status-url="{{ route "updateEntriesStatus" }}"
    {{ if .user }}{{ if not .user.KeyboardShortcuts }}data-disable-keyboard-shortcuts="true"{{ end }}{{ end }}>
    <div class="toast-wrap">
        <span class="toast-msg"></span>
    </div>
    {{ if .user }}
    <header class="header">
        <nav>
            <div class="logo">
                <a href="{{ route "unread" }}">Mini<span>flux</span></a>
            </div>
            <ul>
                <li {{ if eq .menu "unread" }}class="active"{{ end }} title="{{ t "tooltip.keyboard_shortcuts" "g u" }}">
                    <a href="{{ route "unread" }}" data-page="unread">{{ t "menu.unread" }}
                      {{ if gt .countUnread 0 }}
                          <span class="unread-counter-wrapper">(<span class="unread-counter">{{ .countUnread }}</span>)</span>
                      {{ end }}
                    </a>
                </li>
                <li {{ if eq .menu "starred" }}class="active"{{ end }} title="{{ t "tooltip.keyboard_shortcuts" "g b" }}">
                    <a href="{{ route "starred" }}" data-page="starred">{{ t "menu.starred" }}</a>
                </li>
                <li {{ if eq .menu "history" }}class="active"{{ end }} title="{{ t "tooltip.keyboard_shortcuts" "g h" }}">
                    <a href="{{ route "history" }}" data-page="history">{{ t "menu.history" }}</a>
                </li>
                <li {{ if eq .menu "feeds" }}class="active"{{ end }} title="{{ t "tooltip.keyboard_shortcuts" "g f" }}">
                    <a href="{{ route "feeds" }}" data-page="feeds">{{ t "menu.feeds" }}
                      {{ if gt .countErrorFeeds 0 }}
                          <span class="error-feeds-counter-wrapper">(<span class="error-feeds-counter">{{ .countErrorFeeds }}</span>)</span>
                      {{ end }}
                    </a>
                </li>
                <li {{ if eq .menu "categories" }}class="active"{{ end }} title="{{ t "tooltip.keyboard_shortcuts" "g c" }}">
                    <a href="{{ route "categories" }}" data-page="categories">{{ t "menu.categories" }}</a>
                </li>
                <li {{ if eq .menu "settings" }}class="active"{{ end }} title="{{ t "tooltip.keyboard_shortcuts" "g s" }}">
                    <a href="{{ route "settings" }}" data-page="settings">{{ t "menu.settings" }}</a>
                </li>
                <li>
                    <a href="{{ route "logout" }}" title="{{ t "tooltip.logged_user" .user.Username }}">{{ t "menu.logout" }}</a>
                </li>
            </ul>
            <div class="search">
                <div class="search-toggle-switch {{ if $.searchQuery }}has-search-query{{ end }}">
                    <a href="#" data-action="search">&laquo;&nbsp;{{ t "search.label" }}</a>
                </div>
                <form action="{{ route "searchEntries" }}" class="search-form {{ if $.searchQuery }}has-search-query{{ end }}">
                    <input type="search" name="q" id="search-input" placeholder="{{ t "search.placeholder" }}" {{ if $.searchQuery }}value="{{ .searchQuery }}"{{ end }} required>
                </form>
            </div>
        </nav>
    </header>
    {{ end }}
    {{ if .flashMessage }}
        <div class="flash-message alert alert-success">{{ .flashMessage }}</div>
    {{ end }}
    {{ if .flashErrorMessage }}
        <div class="flash-error-message alert alert-error">{{ .flashErrorMessage }}</div>
    {{ end }}
    <main>
        {{template "content" .}}
    </main>
    <template id="keyboard-shortcuts">
        <div id="modal-left">
            <a href="#" class="btn-close-modal">x</a>
            <h3>{{ t "page.keyboard_shortcuts.title" }}</h3>

            <div class="keyboard-shortcuts">
                <p>{{ t "page.keyboard_shortcuts.subtitle.sections" }}</p>
                <ul>
                    <li>{{ t "page.keyboard_shortcuts.go_to_unread" }} = <strong>g + u</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_starred" }} = <strong>g + b</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_history" }} = <strong>g + h</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_feeds" }} = <strong>g + f</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_categories" }} = <strong>g + c</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_settings" }} = <strong>g + s</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.show_keyboard_shortcuts" }} = <strong>?</strong></li>
                </ul>

                <p>{{ t "page.keyboard_shortcuts.subtitle.items" }}</p>
                <ul>
                    <li>{{ t "page.keyboard_shortcuts.go_to_previous_item" }} = <strong>p</strong>, <strong>k</strong>, <strong>◄</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_next_item" }} = <strong>n</strong>, <strong>j</strong>, <strong>►</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_feed" }} = <strong>g + f</strong></li>
                </ul>

                <p>{{ t "page.keyboard_shortcuts.subtitle.pages" }}</p>
                <ul>
                    <li>{{ t "page.keyboard_shortcuts.go_to_previous_page" }} = <strong>h</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_next_page" }} = <strong>l</strong></li>
                </ul>

                <p>{{ t "page.keyboard_shortcuts.subtitle.actions" }}</p>
                <ul>
                    <li>{{ t "page.keyboard_shortcuts.open_item" }} = <strong>o</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.open_original" }} = <strong>v</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.open_original_same_window" }} = <strong>V</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.open_comments" }} = <strong>c</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.open_comments_same_window" }} = <strong>C</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.toggle_read_status" }} = <strong>m</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.mark_page_as_read" }} = <strong>A</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.download_content" }} = <strong>d</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.toggle_bookmark_status" }} = <strong>f</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.save_article" }} = <strong>s</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.remove_feed" }} = <strong>#</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.go_to_search" }} = <strong>/</strong></li>
                    <li>{{ t "page.keyboard_shortcuts.close_modal" }} = <strong>Esc</strong></li>
                </ul>
            </div>
        </div>
    </template>
</body>
</html>
{{ end }}
`,
	"pagination": `{{ define "pagination" }}
<div class="pagination">
    <div class="pagination-prev">
        {{ if .ShowPrev }}
            <a href="{{ .Route }}{{ if gt .PrevOffset 0 }}?offset={{ .PrevOffset }}{{ if .SearchQuery }}&amp;q={{ .SearchQuery }}{{ end }}{{ else }}{{ if .SearchQuery }}?q={{ .SearchQuery }}{{ end }}{{ end }}" data-page="previous">{{ t "pagination.previous" }}</a>
        {{ else }}
            {{ t "pagination.previous" }}
        {{ end }}
    </div>

    <div class="pagination-next">
        {{ if .ShowNext }}
            <a href="{{ .Route }}?offset={{ .NextOffset }}{{ if .SearchQuery }}&amp;q={{ .SearchQuery }}{{ end }}" data-page="next">{{ t "pagination.next" }}</a>
        {{ else }}
            {{ t "pagination.next" }}
        {{ end }}
    </div>
</div>
{{ end }}
`,
	"settings_menu": `{{ define "settings_menu" }}
<ul>
    <li>
        <a href="{{ route "settings" }}">{{ t "menu.settings" }}</a>
    </li>
    <li>
        <a href="{{ route "integrations" }}">{{ t "menu.integrations" }}</a>
    </li>
    <li>
        <a href="{{ route "sessions" }}">{{ t "menu.sessions" }}</a>
    </li>
    {{ if .user.IsAdmin }}
        <li>
            <a href="{{ route "users" }}">{{ t "menu.users" }}</a>
        </li>
        <li>
            <a href="{{ route "createUser" }}">{{ t "menu.add_user" }}</a>
        </li>
    {{ end }}
    <li>
        <a href="{{ route "about" }}">{{ t "menu.about" }}</a>
    </li>
</ul>
{{ end }}`,
}

var templateCommonMapChecksums = map[string]string{
	"entry_pagination": "4faa91e2eae150c5e4eab4d258e039dfdd413bab7602f0009360e6d52898e353",
	"feed_list":        "db406e7cb81292ce1d974d63f63270384a286848b2e74fe36bf711b4eb5717dd",
	"feed_menu":        "318d8662dda5ca9dfc75b909c8461e79c86fb5082df1428f67aaf856f19f4b50",
	"item_meta":        "061107d7adf9b7430b5f28a0f452a294df73c3b16c19274e9ce6f7cdf71064a4",
	"layout":           "a1f67b8908745ee4f9cee6f7bbbb0b242d4dcc101207ad4a9d67242b45683299",
	"pagination":       "3386e90c6e1230311459e9a484629bc5d5bf39514a75ef2e73bbbc61142f7abb",
	"settings_menu":    "78e5a487ede18610b23db74184dab023170f9e083cc0625bc2c874d1eea1a4ce",
}
