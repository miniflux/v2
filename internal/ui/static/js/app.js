// OnClick attaches a listener to the elements that match the selector.
function onClick(selector, callback, noPreventDefault) {
    document.querySelectorAll(selector).forEach((element) => {
        element.onclick = (event) => {
            if (!noPreventDefault) {
                event.preventDefault();
            }
            callback(event);
        };
    });
}

function onAuxClick(selector, callback, noPreventDefault) {
    document.querySelectorAll(selector).forEach((element) => {
        element.onauxclick = (event) => {
            if (!noPreventDefault) {
                event.preventDefault();
            }
            callback(event);
        };
    });
}

// make logo element as button on mobile layout
function checkMenuToggleModeByLayout() {
    const logoElement = document.querySelector(".logo");
    if (!logoElement) return;

    const homePageLinkElement = document.querySelector(".logo > a");

    if (document.documentElement.clientWidth < 620) {
        const navMenuElement = document.getElementById("header-menu");
        const navMenuElementIsExpanded = navMenuElement.classList.contains("js-menu-show");
        const logoToggleButtonLabel = logoElement.getAttribute("data-toggle-button-label");
        logoElement.setAttribute("role", "button");
        logoElement.setAttribute("tabindex", "0");
        logoElement.setAttribute("aria-label", logoToggleButtonLabel);
        logoElement.setAttribute("aria-expanded", navMenuElementIsExpanded?"true":"false");
        homePageLinkElement.setAttribute("tabindex", "-1");
    } else {
        logoElement.removeAttribute("role");
        logoElement.removeAttribute("tabindex");
        logoElement.removeAttribute("aria-expanded");
        logoElement.removeAttribute("aria-label");
        homePageLinkElement.removeAttribute("tabindex");
    }
}

function fixVoiceOverDetailsSummaryBug() {
    document.querySelectorAll("details").forEach((details) => {
        const summaryElement = details.querySelector("summary");
        summaryElement.setAttribute("role", "button");
        summaryElement.setAttribute("aria-expanded", details.open? "true": "false");

        details.addEventListener("toggle", () => {
            summaryElement.setAttribute("aria-expanded", details.open? "true": "false");
        });
    });
}

// Show and hide the main menu on mobile devices.
function toggleMainMenu(event) {
    if (event.type === "keydown" && !(event.key === "Enter" || event.key === " ")) {
        return;
    }

    if (event.currentTarget.getAttribute("role")) {
        event.preventDefault();
    }

    const menu = document.querySelector(".header nav ul");
    const menuToggleButton = document.querySelector(".logo");
    if (menu.classList.contains("js-menu-show")) {
        menu.classList.remove("js-menu-show");
        menuToggleButton.setAttribute("aria-expanded", false);
    } else {
        menu.classList.add("js-menu-show");
        menuToggleButton.setAttribute("aria-expanded", true);
    }
}

// Handle click events for the main menu (<li> and <a>).
function onClickMainMenuListItem(event) {
    const element = event.target;

    if (element.tagName === "A") {
        window.location.href = element.getAttribute("href");
    } else {
        window.location.href = element.querySelector("a").getAttribute("href");
    }
}

// Change the button label when the page is loading.
function handleSubmitButtons() {
    document.querySelectorAll("form").forEach((element) => {
        element.onsubmit = () => {
            const button = element.querySelector("button");
            if (button) {
                button.textContent = button.dataset.labelLoading;
                button.disabled = true;
            }
        };
    });
}

// Show modal dialog with the list of keyboard shortcuts.
function showKeyboardShortcuts() {
    const template = document.getElementById("keyboard-shortcuts");
    if (template !== null) {
        ModalHandler.open(template.content, "dialog-title");
    }
}

// Mark as read visible items of the current page.
function markPageAsRead() {
    const items = DomHelper.getVisibleElements(".items .item");
    const entryIDs = [];

    items.forEach((element) => {
        element.classList.add("item-status-read");
        entryIDs.push(parseInt(element.dataset.id, 10));
    });

    if (entryIDs.length > 0) {
        updateEntriesStatus(entryIDs, "read", () => {
            // Make sure the Ajax request reach the server before we reload the page.

            const element = document.querySelector(":is(a, button)[data-action=markPageAsRead]");
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
    const toasting = !element;
    const currentEntry = findEntry(element);
    if (currentEntry) {
        if (!setToRead || currentEntry.querySelector(":is(a, button)[data-toggle-status]").dataset.value == "unread") {
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

// Add a span-icon with a `label` to `element` as a child
function addIcon(element, label) {
    const span = document.createElement('span');
    span.classList.add('icon-label');
    span.textContent = label;
    element.appendChild(span);
}

// Change the entry status to the opposite value.
function toggleEntryStatus(element, toasting) {
    const entryID = parseInt(element.dataset.id, 10);
    const link = element.querySelector(":is(a, button)[data-toggle-status]");

    const currentStatus = link.dataset.value;
    const newStatus = currentStatus === "read" ? "unread" : "read";

    link.querySelector("span").textContent = link.dataset.labelLoading;
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

        link.replaceChildren(iconElement.content.cloneNode(true));
        addIcon(link, label);
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

        const entryID = parseInt(element.dataset.id, 10);
        updateEntriesStatus([entryID], "read");
    }
}

// Send the Ajax request to refresh all feeds in the background
function handleRefreshAllFeeds() {
    const url = document.body.dataset.refreshAllFeedsUrl;
    if (url) {
        window.location.href = url;
    }
}

// Send the Ajax request to change entries statuses.
function updateEntriesStatus(entryIDs, status, callback) {
    const url = document.body.dataset.entriesStatusUrl;
    const request = new RequestBuilder(url);
    request.withBody({ entry_ids: entryIDs, status: status });
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
    const toasting = !element;
    const currentEntry = findEntry(element);
    if (currentEntry) {
        saveEntry(currentEntry.querySelector(":is(a, button)[data-save-entry]"), toasting);
    }
}

// Send the Ajax request to save an entry.
function saveEntry(element, toasting) {
    if (!element || element.dataset.completed) {
        return;
    }

    element.textContent = "";
    addIcon(element, element.dataset.labelLoading);

    const request = new RequestBuilder(element.dataset.saveUrl);
    request.withCallback(() => {
        element.textContent = "";
        addIcon(element, element.dataset.labelDone);
        element.dataset.completed = true;
        if (toasting) {
            const iconElement = document.querySelector("template#icon-save");
            showToast(element.dataset.toastDone, iconElement);
        }
    });
    request.execute();
}

// Handle bookmark from the list view and entry view.
function handleBookmark(element) {
    const toasting = !element;
    const currentEntry = findEntry(element);
    if (currentEntry) {
        toggleBookmark(currentEntry, toasting);
    }
}

// Send the Ajax request and change the icon when bookmarking an entry.
function toggleBookmark(parentElement, toasting) {
    const element = parentElement.querySelector(":is(a, button)[data-toggle-bookmark]");
    if (!element) {
        return;
    }

    element.textContent = "";
    addIcon(element, element.dataset.labelLoading);

    const request = new RequestBuilder(element.dataset.bookmarkUrl);
    request.withCallback(() => {
        const currentStarStatus = element.dataset.value;
        const newStarStatus = currentStarStatus === "star" ? "unstar" : "star";

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

        element.replaceChildren(iconElement.content.cloneNode(true));
        addIcon(element, label);
        element.dataset.value = newStarStatus;
    });
    request.execute();
}

// Send the Ajax request to download the original web page.
function handleFetchOriginalContent() {
    if (isListView()) {
        return;
    }

    const element = document.querySelector(":is(a, button)[data-fetch-content-entry]");
    if (!element) {
        return;
    }

    const previousElement = element.cloneNode(true);
    addIcon(element, element.dataset.labelLoading);

    const request = new RequestBuilder(element.dataset.fetchContentUrl);
    request.withCallback((response) => {
        element.textContent = '';
        element.appendChild(previousElement);

        response.json().then((data) => {
            if (data.hasOwnProperty("content") && data.hasOwnProperty("reading_time")) {
                document.querySelector(".entry-content").innerHTML = data.content;
                const entryReadingtimeElement = document.querySelector(".entry-reading-time");
                if (entryReadingtimeElement) {
                    entryReadingtimeElement.textContent = data.reading_time;
                }
            }
        });
    });
    request.execute();
}

function openOriginalLink(openLinkInCurrentTab) {
    const entryLink = document.querySelector(".entry h1 a");
    if (entryLink !== null) {
        if (openLinkInCurrentTab) {
            window.location.href = entryLink.getAttribute("href");
        } else {
            DomHelper.openNewTab(entryLink.getAttribute("href"));
        }
        return;
    }

    const currentItemOriginalLink = document.querySelector(".current-item :is(a, button)[data-original-link]");
    if (currentItemOriginalLink !== null) {
        DomHelper.openNewTab(currentItemOriginalLink.getAttribute("href"));

        const currentItem = document.querySelector(".current-item");
        // If we are not on the list of starred items, move to the next item
        if (document.location.href != document.querySelector(':is(a, button)[data-page=starred]').href) {
            goToListItem(1);
        }
        markEntryAsRead(currentItem);
    }
}

function openCommentLink(openLinkInCurrentTab) {
    if (!isListView()) {
        const entryLink = document.querySelector(":is(a, button)[data-comments-link]");
        if (entryLink !== null) {
            if (openLinkInCurrentTab) {
                window.location.href = entryLink.getAttribute("href");
            } else {
                DomHelper.openNewTab(entryLink.getAttribute("href"));
            }
            return;
        }
    } else {
        const currentItemCommentsLink = document.querySelector(".current-item :is(a, button)[data-comments-link]");
        if (currentItemCommentsLink !== null) {
            DomHelper.openNewTab(currentItemCommentsLink.getAttribute("href"));
        }
    }
}

function openSelectedItem() {
    const currentItemLink = document.querySelector(".current-item .item-title a");
    if (currentItemLink !== null) {
        window.location.href = currentItemLink.getAttribute("href");
    }
}

function unsubscribeFromFeed() {
    const unsubscribeLinks = document.querySelectorAll("[data-action=remove-feed]");
    if (unsubscribeLinks.length === 1) {
        const unsubscribeLink = unsubscribeLinks[0];

        const request = new RequestBuilder(unsubscribeLink.dataset.url);
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
    const element = document.querySelector(":is(a, button)[data-page=" + page + "]");

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
        const feedAnchor = document.querySelector("span.entry-website a");
        if (feedAnchor !== null) {
            window.location.href = feedAnchor.href;
        }
    } else {
        const currentItemFeed = document.querySelector(".current-item :is(a, button)[data-feed-link]");
        if (currentItemFeed !== null) {
            window.location.href = currentItemFeed.getAttribute("href");
        }
    }
}

/**
 * @param {number} offset How many items to jump for focus.
 */
function goToListItem(offset) {
    const items = DomHelper.getVisibleElements(".items .item");
    if (items.length === 0) {
        return;
    }

    if (document.querySelector(".current-item") === null) {
        items[0].classList.add("current-item");
        items[0].focus();
        return;
    }

    for (let i = 0; i < items.length; i++) {
        if (items[i].classList.contains("current-item")) {
            items[i].classList.remove("current-item");

            const index = (i + offset + items.length) % items.length;
            const item = items[index];

            item.classList.add("current-item");
            DomHelper.scrollPageTo(item);
            item.focus();

            break;
        }
    }
}

function scrollToCurrentItem() {
    const currentItem = document.querySelector(".current-item");
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
    document.querySelectorAll("span.unread-counter").forEach((element) => {
        const oldValue = parseInt(element.textContent, 10);
        element.textContent = callback(oldValue);
    });

    if (window.location.href.endsWith('/unread')) {
        const oldValue = parseInt(document.title.split('(')[1], 10);
        const newValue = callback(oldValue);

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
            return element.closest(".item");
        }
        return document.querySelector(".current-item");
    }
    return document.querySelector(".entry");
}

function handleConfirmationMessage(linkElement, callback) {
    if (linkElement.tagName != 'A' && linkElement.tagName != "BUTTON") {
        linkElement = linkElement.parentNode;
    }

    linkElement.style.display = "none";

    const containerElement = linkElement.parentNode;
    const questionElement = document.createElement("span");

    function createLoadingElement() {
        const loadingElement = document.createElement("span");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(linkElement.dataset.labelLoading));

        questionElement.remove();
        containerElement.appendChild(loadingElement);
    }

    const yesElement = document.createElement("button");
    yesElement.appendChild(document.createTextNode(linkElement.dataset.labelYes));
    yesElement.onclick = (event) => {
        event.preventDefault();

        createLoadingElement();

        callback(linkElement.dataset.url, linkElement.dataset.redirectUrl);
    };

    const noElement = document.createElement("button");
    noElement.appendChild(document.createTextNode(linkElement.dataset.labelNo));
    noElement.onclick = (event) => {
        event.preventDefault();

        const noActionUrl = linkElement.dataset.noActionUrl;
        if (noActionUrl) {
            createLoadingElement();

            callback(noActionUrl, linkElement.dataset.redirectUrl);
        } else {
            linkElement.style.display = "inline";
            questionElement.remove();
        }
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
        toastMsgElement.replaceChildren(iconElement.content.cloneNode(true));
        addIcon(toastMsgElement, label);

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

/**
 * save player position to allow to resume playback later
 * @param {Element} playerElement
 */
function handlePlayerProgressionSave(playerElement) {
    const currentPositionInSeconds = Math.floor(playerElement.currentTime); // we do not need a precise value
    const lastKnownPositionInSeconds = parseInt(playerElement.dataset.lastPosition, 10);
    const recordInterval = 10;

    // we limit the number of update to only one by interval. Otherwise, we would have multiple update per seconds
    if (currentPositionInSeconds >= (lastKnownPositionInSeconds + recordInterval) ||
        currentPositionInSeconds <= (lastKnownPositionInSeconds - recordInterval)
    ) {
        playerElement.dataset.lastPosition = currentPositionInSeconds.toString();
        const request = new RequestBuilder(playerElement.dataset.saveUrl);
        request.withBody({ progression: currentPositionInSeconds });
        request.execute();
    }
}

/**
 * handle new share entires and already shared entries
 */
function handleShare() {
    const link = document.querySelector(':is(a, button)[data-share-status]');
    const title = document.querySelector("body > main > section > header > h1 > a");
    if (link.dataset.shareStatus === "shared") {
        checkShareAPI(title, link.href);
    }
    if (link.dataset.shareStatus === "share") {
        const request = new RequestBuilder(link.href);
        request.withCallback((r) => {
            checkShareAPI(title, r.url);
        });
        request.withHttpMethod("GET");
        request.execute();
    }
}

/**
* wrapper for Web Share API
*/
function checkShareAPI(title, url) {
    if (!navigator.canShare) {
        console.error("Your browser doesn't support the Web Share API.");
        window.location = url;
        return;
    }
    try {
        navigator.share({
            title: title,
            url: url
        });
        window.location.reload();
    } catch (err) {
        console.error(err);
        window.location.reload();
    }
}

function getCsrfToken() {
    const element = document.querySelector("body[data-csrf-token]");
    if (element !== null) {
        return element.dataset.csrfToken;
    }

    return "";
}
