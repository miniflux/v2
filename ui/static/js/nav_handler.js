class NavHandler {
    setFocusToSearchInput(event) {
        event.preventDefault();
        event.stopPropagation();

        let toggleSwitchElement = document.querySelector(".search-toggle-switch");
        if (toggleSwitchElement) {
            toggleSwitchElement.style.display = "none";
        }

        let searchFormElement = document.querySelector(".search-form");
        if (searchFormElement) {
            searchFormElement.style.display = "block";
        }

        let searchInputElement = document.getElementById("search-input");
        if (searchInputElement) {
            searchInputElement.focus();
            searchInputElement.value = "";
        }
    }

    showKeyboardShortcuts() {
        let template = document.getElementById("keyboard-shortcuts");
        if (template !== null) {
            ModalHandler.open(template.content);
        }
    }

    markPageAsRead() {
        let items = DomHelper.getVisibleElements(".items .item");
        let entryIDs = [];

        items.forEach((element) => {
            element.classList.add("item-status-read");
            entryIDs.push(parseInt(element.dataset.id, 10));
        });

        if (entryIDs.length > 0) {
            EntryHandler.updateEntriesStatus(entryIDs, "read", () => {
                // This callback make sure the Ajax request reach the server before we reload the page.
                this.goToPage("next", true);
            });
        }
    }

    saveEntry() {
        if (this.isListView()) {
            let currentItem = document.querySelector(".current-item");
            if (currentItem !== null) {
                let saveLink = currentItem.querySelector("a[data-save-entry]");
                if (saveLink) {
                    EntryHandler.saveEntry(saveLink);
                }
            }
        } else {
            let saveLink = document.querySelector("a[data-save-entry]");
            if (saveLink) {
                EntryHandler.saveEntry(saveLink);
            }
        }
    }

    fetchOriginalContent() {
        if (! this.isListView()){
            let link = document.querySelector("a[data-fetch-content-entry]");
            if (link) {
                EntryHandler.fetchOriginalContent(link);
            }
        }
    }

    toggleEntryStatus() {
        if (! this.isListView()) {
            EntryHandler.toggleEntryStatus(document.querySelector(".entry"));
            return;
        }

        let currentItem = document.querySelector(".current-item");
        if (currentItem !== null) {
            // The order is important here,
            // On the unread page, the read item will be hidden.
            this.goToNextListItem();
            EntryHandler.toggleEntryStatus(currentItem);
        }
    }

    toggleBookmark() {
        if (! this.isListView()) {
            this.toggleBookmarkLink(document.querySelector(".entry"));
            return;
        }

        let currentItem = document.querySelector(".current-item");
        if (currentItem !== null) {
            this.toggleBookmarkLink(currentItem);
        }
    }

    toggleBookmarkLink(parent) {
        let bookmarkLink = parent.querySelector("a[data-toggle-bookmark]");
        if (bookmarkLink) {
            EntryHandler.toggleBookmark(bookmarkLink);
        }
    }

    openOriginalLink() {
        let entryLink = document.querySelector(".entry h1 a");
        if (entryLink !== null) {
            DomHelper.openNewTab(entryLink.getAttribute("href"));
            return;
        }

        let currentItemOriginalLink = document.querySelector(".current-item a[data-original-link]");
        if (currentItemOriginalLink !== null) {
            DomHelper.openNewTab(currentItemOriginalLink.getAttribute("href"));

            // Move to the next item and if we are on the unread page mark this item as read.
            let currentItem = document.querySelector(".current-item");
            this.goToNextListItem();
            EntryHandler.markEntryAsRead(currentItem);
        }
    }

    openSelectedItem() {
        let currentItemLink = document.querySelector(".current-item .item-title a");
        if (currentItemLink !== null) {
            window.location.href = currentItemLink.getAttribute("href");
        }
    }

    unsubscribeFromFeed() {
        let unsubscribeLinks = document.querySelectorAll("[data-action=remove-feed]");
        if (unsubscribeLinks.length === 1) {
            let unsubscribeLink = unsubscribeLinks[0];
            FeedHandler.unsubscribe(unsubscribeLink.dataset.url, () => {
                if (unsubscribeLink.dataset.redirectUrl) {
                    window.location.href = unsubscribeLink.dataset.redirectUrl;
                } else {
                    window.location.reload();
                }
            });
        }
    }

    /**
     * @param {string} page Page to redirect to.
     * @param {boolean} fallbackSelf Refresh actual page if the page is not found.
     */
    goToPage(page, fallbackSelf) {
        let element = document.querySelector("a[data-page=" + page + "]");

        if (element) {
            document.location.href = element.href;
        } else if (fallbackSelf) {
            window.location.reload();
        }
    }

    goToPrevious() {
        if (this.isListView()) {
            this.goToPreviousListItem();
        } else {
            this.goToPage("previous");
        }
    }

    goToNext() {
        if (this.isListView()) {
            this.goToNextListItem();
        } else {
            this.goToPage("next");
        }
    }

    goToPreviousListItem() {
        let items = DomHelper.getVisibleElements(".items .item");
        if (items.length === 0) {
            return;
        }

        if (document.querySelector(".current-item") === null) {
            items[0].classList.add("current-item");
            items[0].querySelector('.item-header a').focus();
            return;
        }

        for (let i = 0; i < items.length; i++) {
            if (items[i].classList.contains("current-item")) {
                items[i].classList.remove("current-item");

                if (i - 1 >= 0) {
                    items[i - 1].classList.add("current-item");
                    DomHelper.scrollPageTo(items[i - 1]);
                    items[i - 1].querySelector('.item-header a').focus();
                }

                break;
            }
        }
    }

    goToNextListItem() {
        let currentItem = document.querySelector(".current-item");
        let items = DomHelper.getVisibleElements(".items .item");
        if (items.length === 0) {
            return;
        }

        if (currentItem === null) {
            items[0].classList.add("current-item");
            items[0].querySelector('.item-header a').focus();
            return;
        }

        for (let i = 0; i < items.length; i++) {
            if (items[i].classList.contains("current-item")) {
                items[i].classList.remove("current-item");

                if (i + 1 < items.length) {
                    items[i + 1].classList.add("current-item");
                    DomHelper.scrollPageTo(items[i + 1]);
                    items[i + 1].querySelector('.item-header a').focus();
                }

                break;
            }
        }
    }

    isListView() {
        return document.querySelector(".items") !== null;
    }
}
