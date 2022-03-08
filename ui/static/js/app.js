// OnClick attaches a listener to the elements that match the selector.
function onClick(selector, callback, noPreventDefault) {
    let elements = document.querySelectorAll(selector);
    elements.forEach((element) => {
        element.onclick = (event) => {
            if (!noPreventDefault) {
                event.preventDefault();
            }

            callback(event);
        };
    });
}

function onAuxClick(selector, callback, noPreventDefault) {
    let elements = document.querySelectorAll(selector);
    elements.forEach((element) => {
        element.onauxclick = (event) => {
            if (!noPreventDefault) {
                event.preventDefault();
            }

            callback(event);
        };
    });
}

// Show and hide the main menu on mobile devices.
function toggleMainMenu() {
    let menu = document.querySelector(".header nav ul");
    if (DomHelper.isVisible(menu)) {
        menu.style.display = "none";
    } else {
        menu.style.display = "block";
    }

    let searchElement = document.querySelector(".header .search");
    if (DomHelper.isVisible(searchElement)) {
        searchElement.style.display = "none";
    } else {
        searchElement.style.display = "block";
    }
}

// Handle click events for the main menu (<li> and <a>).
function onClickMainMenuListItem(event) {
    let element = event.target;

    if (element.tagName === "A") {
        window.location.href = element.getAttribute("href");
    } else {
        window.location.href = element.querySelector("a").getAttribute("href");
    }
}

// Change the button label when the page is loading.
function handleSubmitButtons() {
    let elements = document.querySelectorAll("form");
    elements.forEach((element) => {
        element.onsubmit = () => {
            let button = element.querySelector("button");

            if (button) {
                button.innerHTML = button.dataset.labelLoading;
                button.disabled = true;
            }
        };
    });
}

// Set cursor focus to the search input.
function setFocusToSearchInput(event) {
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

// Show modal dialog with the list of keyboard shortcuts.
function showKeyboardShortcuts() {
    let template = document.getElementById("keyboard-shortcuts");
    if (template !== null) {
        ModalHandler.open(template.content);
    }
}

// Mark as read visible items of the current page.
function markPageAsRead() {
    let items = DomHelper.getVisibleElements(".items .item");
    let entryIDs = [];

    items.forEach((element) => {
        element.classList.add("item-status-read");
        entryIDs.push(parseInt(element.dataset.id, 10));
    });

    if (entryIDs.length > 0) {
        updateEntriesStatus(entryIDs, "read", () => {
            // Make sure the Ajax request reach the server before we reload the page.

            let element = document.querySelector("a[data-action=markPageAsRead]");
            let showOnlyUnread = false;
            if (element) {
                showOnlyUnread = element.dataset.showOnlyUnread || false;
            }

            if (showOnlyUnread) {
                window.location.href = window.location.href;
            } else {
                goToPage("next", true);
            }
        });
    }
}

/**
 * Handle entry status changes from the list view and entry view.
 * Focus the next or the previous entry if it exists.
 * @param {string} item Item to focus: "previous" or "next".
 * @param {Element} element
 * @param {boolean} setToRead
 */
function handleEntryStatus(item, element, setToRead) {
    let toasting = !element;
    let currentEntry = findEntry(element);
    if (currentEntry) {
        if (!setToRead || currentEntry.querySelector("a[data-toggle-status]").dataset.value == "unread") {
            toggleEntryStatus(currentEntry, toasting);
        }
        if (isListView() && currentEntry.classList.contains('current-item')) {
            switch (item) {
                case "previous":
                    goToListItem(-1);
                    break;
                case "next":
                    goToListItem(1);
                    break;
            }
        }
    }
}

// Change the entry status to the opposite value.
function toggleEntryStatus(element, toasting) {
    let entryID = parseInt(element.dataset.id, 10);
    let link = element.querySelector("a[data-toggle-status]");

    let currentStatus = link.dataset.value;
    let newStatus = currentStatus === "read" ? "unread" : "read";

    link.querySelector("span").innerHTML = link.dataset.labelLoading;
    updateEntriesStatus([entryID], newStatus, () => {
        let iconElement, label;

        if (currentStatus === "read") {
            iconElement = document.querySelector("template#icon-read");
            label = link.dataset.labelRead;
            if (toasting) {
                showToast(link.dataset.toastUnread, iconElement);
            }
        } else {
            iconElement = document.querySelector("template#icon-unread");
            label = link.dataset.labelUnread;
            if (toasting) {
                showToast(link.dataset.toastRead, iconElement);
            }
        }

        link.innerHTML = iconElement.innerHTML + '<span class="icon-label">' + label + '</span>';
        link.dataset.value = newStatus;

        if (element.classList.contains("item-status-" + currentStatus)) {
            element.classList.remove("item-status-" + currentStatus);
            element.classList.add("item-status-" + newStatus);
        }
    });
}

// Mark a single entry as read.
function markEntryAsRead(element) {
    if (element.classList.contains("item-status-unread")) {
        element.classList.remove("item-status-unread");
        element.classList.add("item-status-read");

        let entryID = parseInt(element.dataset.id, 10);
        updateEntriesStatus([entryID], "read");
    }
}

// Send the Ajax request to refresh all feeds in the background
function handleRefreshAllFeeds() {
    let url = document.body.dataset.refreshAllFeedsUrl;
    let request = new RequestBuilder(url);

    request.withCallback(() => {
        window.location.reload();
    });

    request.withHttpMethod("GET");
    request.execute();
}

// Send the Ajax request to change entries statuses.
function updateEntriesStatus(entryIDs, status, callback) {
    let url = document.body.dataset.entriesStatusUrl;
    let request = new RequestBuilder(url);
    request.withBody({entry_ids: entryIDs, status: status});
    request.withCallback((resp) => {
        resp.json().then(count => {
        if (callback) {
            callback(resp);
        }

            if (status === "read") {
                decrementUnreadCounter(count);
            } else {
                incrementUnreadCounter(count);
            }
        });
    });
    request.execute();
}

// Handle save entry from list view and entry view.
function handleSaveEntry(element) {
    let toasting = !element;
    let currentEntry = findEntry(element);
    if (currentEntry) {
        saveEntry(currentEntry.querySelector("a[data-save-entry]"), toasting);
    }
}

// Send the Ajax request to save an entry.
function saveEntry(element, toasting) {
    if (!element) {
        return;
    }

    if (element.dataset.completed) {
        return;
    }

    let previousInnerHTML = element.innerHTML;
    element.innerHTML = '<span class="icon-label">' + element.dataset.labelLoading + '</span>';

    let request = new RequestBuilder(element.dataset.saveUrl);
    request.withCallback(() => {
        element.innerHTML = previousInnerHTML;
        element.dataset.completed = true;
        if (toasting) {
            let iconElement = document.querySelector("template#icon-save");
            showToast(element.dataset.toastDone, iconElement);
        }
    });
    request.execute();
}

// Handle bookmark from the list view and entry view.
function handleBookmark(element) {
    let toasting = !element;
    let currentEntry = findEntry(element);
    if (currentEntry) {
        toggleBookmark(currentEntry, toasting);
    }
}

// Send the Ajax request and change the icon when bookmarking an entry.
function toggleBookmark(parentElement, toasting) {
    let element = parentElement.querySelector("a[data-toggle-bookmark]");
    if (!element) {
        return;
    }

    element.innerHTML = '<span class="icon-label">' + element.dataset.labelLoading + '</span>';

    let request = new RequestBuilder(element.dataset.bookmarkUrl);
    request.withCallback(() => {

        let currentStarStatus = element.dataset.value;
        let newStarStatus = currentStarStatus === "star" ? "unstar" : "star";

        let iconElement, label;

        if (currentStarStatus === "star") {
            iconElement = document.querySelector("template#icon-star");
            label = element.dataset.labelStar;
            if (toasting) {
                showToast(element.dataset.toastUnstar, iconElement);
            }
        } else {
            iconElement = document.querySelector("template#icon-unstar");
            label = element.dataset.labelUnstar;
            if (toasting) {
                showToast(element.dataset.toastStar, iconElement);
            }
        }

        element.innerHTML = iconElement.innerHTML + '<span class="icon-label">' + label + '</span>';
        element.dataset.value = newStarStatus;
    });
    request.execute();
}

// Send the Ajax request to download the original web page.
function handleFetchOriginalContent() {
    if (isListView()) {
        return;
    }

    let element = document.querySelector("a[data-fetch-content-entry]");
    if (!element) {
        return;
    }

    let previousInnerHTML = element.innerHTML;
    element.innerHTML = '<span class="icon-label">' + element.dataset.labelLoading + '</span>';

    let request = new RequestBuilder(element.dataset.fetchContentUrl);
    request.withCallback((response) => {
        element.innerHTML = previousInnerHTML;

        response.json().then((data) => {
            if (data.hasOwnProperty("content")) {
                document.querySelector(".entry-content").innerHTML = data.content;
            }
        });
    });
    request.execute();
}

function openOriginalLink(openLinkInCurrentTab) {
    let entryLink = document.querySelector(".entry h1 a");
    if (entryLink !== null) {
        if (openLinkInCurrentTab) {
            window.location.href = entryLink.getAttribute("href");
        } else {
            DomHelper.openNewTab(entryLink.getAttribute("href"));
        }
        return;
    }

    let currentItemOriginalLink = document.querySelector(".current-item a[data-original-link]");
    if (currentItemOriginalLink !== null) {
        DomHelper.openNewTab(currentItemOriginalLink.getAttribute("href"));

        let currentItem = document.querySelector(".current-item");
        // If we are not on the list of starred items, move to the next item
        if (document.location.href != document.querySelector('a[data-page=starred]').href) {
            goToListItem(1);
        }
        markEntryAsRead(currentItem);
    }
}

function openCommentLink(openLinkInCurrentTab) {
    if (!isListView()) {
        let entryLink = document.querySelector("a[data-comments-link]");
        if (entryLink !== null) {
            if (openLinkInCurrentTab) {
                window.location.href = entryLink.getAttribute("href");
            } else {
                DomHelper.openNewTab(entryLink.getAttribute("href"));
            }
            return;
        }
    } else {
        let currentItemCommentsLink = document.querySelector(".current-item a[data-comments-link]");
        if (currentItemCommentsLink !== null) {
            DomHelper.openNewTab(currentItemCommentsLink.getAttribute("href"));
        }
    }
}

function openSelectedItem() {
    let currentItemLink = document.querySelector(".current-item .item-title a");
    if (currentItemLink !== null) {
        window.location.href = currentItemLink.getAttribute("href");
    }
}

function unsubscribeFromFeed() {
    let unsubscribeLinks = document.querySelectorAll("[data-action=remove-feed]");
    if (unsubscribeLinks.length === 1) {
        let unsubscribeLink = unsubscribeLinks[0];

        let request = new RequestBuilder(unsubscribeLink.dataset.url);
        request.withCallback(() => {
            if (unsubscribeLink.dataset.redirectUrl) {
                window.location.href = unsubscribeLink.dataset.redirectUrl;
            } else {
                window.location.reload();
            }
        });
        request.execute();
    }
}

/**
 * @param {string} page Page to redirect to.
 * @param {boolean} fallbackSelf Refresh actual page if the page is not found.
 */
function goToPage(page, fallbackSelf) {
    let element = document.querySelector("a[data-page=" + page + "]");

    if (element) {
        document.location.href = element.href;
    } else if (fallbackSelf) {
        window.location.reload();
    }
}

function goToPrevious() {
    if (isListView()) {
        goToListItem(-1);
    } else {
        goToPage("previous");
    }
}

function goToNext() {
    if (isListView()) {
        goToListItem(1);
    } else {
        goToPage("next");
    }
}

function goToFeedOrFeeds() {
    if (isEntry()) {
        goToFeed();
    } else {
        goToPage('feeds');
    }
}

function goToFeed() {
    if (isEntry()) {
        let feedAnchor = document.querySelector("span.entry-website a");
        if (feedAnchor !== null) {
            window.location.href = feedAnchor.href;
        }
    } else {
        let currentItemFeed = document.querySelector(".current-item a[data-feed-link]");
        if (currentItemFeed !== null) {
            window.location.href = currentItemFeed.getAttribute("href");
        }
    }
}

/**
 * @param {number} offset How many items to jump for focus.
 */
function goToListItem(offset) {
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

            let item = items[(i + offset + items.length) % items.length];

            item.classList.add("current-item");
            DomHelper.scrollPageTo(item);
            item.querySelector('.item-header a').focus();

            break;
        }
    }
}

function scrollToCurrentItem() {
    let currentItem = document.querySelector(".current-item");
    if (currentItem !== null) {
        DomHelper.scrollPageTo(currentItem, true);
    }
}

function decrementUnreadCounter(n) {
    updateUnreadCounterValue((current) => {
        return current - n;
    });
}

function incrementUnreadCounter(n) {
    updateUnreadCounterValue((current) => {
        return current + n;
    });
}

function updateUnreadCounterValue(callback) {
    let counterElements = document.querySelectorAll("span.unread-counter");
    counterElements.forEach((element) => {
        let oldValue = parseInt(element.textContent, 10);
        element.innerHTML = callback(oldValue);
    });

    if (window.location.href.endsWith('/unread')) {
        let oldValue = parseInt(document.title.split('(')[1], 10);
        let newValue = callback(oldValue);

        document.title = document.title.replace(
            /(.*?)\(\d+\)(.*?)/,
            function (match, prefix, suffix, offset, string) {
                return prefix + '(' + newValue + ')' + suffix;
            }
        );
    }
}

function isEntry() {
    return document.querySelector("section.entry") !== null;
}

function isListView() {
    return document.querySelector(".items") !== null;
}

function findEntry(element) {
    if (isListView()) {
        if (element) {
            return DomHelper.findParent(element, "item");
        } else {
            return document.querySelector(".current-item");
        }
    } else {
        return document.querySelector(".entry");
    }
}

function handleConfirmationMessage(linkElement, callback) {
    if (linkElement.tagName != 'A') {
        linkElement = linkElement.parentNode;
    }

    linkElement.style.display = "none";
    
    let containerElement = linkElement.parentNode;
    let questionElement = document.createElement("span");

    let yesElement = document.createElement("a");
    yesElement.href = "#";
    yesElement.appendChild(document.createTextNode(linkElement.dataset.labelYes));
    yesElement.onclick = (event) => {
        event.preventDefault();

        let loadingElement = document.createElement("span");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(linkElement.dataset.labelLoading));

        questionElement.remove();
        containerElement.appendChild(loadingElement);

        callback(linkElement.dataset.url, linkElement.dataset.redirectUrl);
    };

    let noElement = document.createElement("a");
    noElement.href = "#";
    noElement.appendChild(document.createTextNode(linkElement.dataset.labelNo));
    noElement.onclick = (event) => {
        event.preventDefault();
        linkElement.style.display = "inline";
        questionElement.remove();
    };

    questionElement.className = "confirm";
    questionElement.appendChild(document.createTextNode(linkElement.dataset.labelQuestion + " "));
    questionElement.appendChild(yesElement);
    questionElement.appendChild(document.createTextNode(", "));
    questionElement.appendChild(noElement);

    containerElement.appendChild(questionElement);
}

function showToast(label, iconElement) {
    if (!label || !iconElement) {
        return;
    }

    const toastMsgElement = document.getElementById("toast-msg");
    if (toastMsgElement) {
        toastMsgElement.innerHTML = iconElement.innerHTML + '<span class="icon-label">' + label + '</span>';

        const toastElementWrapper = document.getElementById("toast-wrapper");
        if (toastElementWrapper) {
            toastElementWrapper.classList.remove('toast-animate');
            setTimeout(function () {
                toastElementWrapper.classList.add('toast-animate');
            }, 100);
        }
    }
}

/** Navigate to the new subscription page. */
function goToAddSubscription() {
    window.location.href = document.body.dataset.addSubscriptionUrl;
}
