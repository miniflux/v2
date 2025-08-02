// Sentinel values for specific list navigation
const TOP = 9999;
const BOTTOM = -9999;

/**
 * Get the CSRF token from the HTML document.
 *
 * @returns {string} The CSRF token.
 */
function getCsrfToken() {
    return document.body.dataset.csrfToken || "";
}

/**
 * Open a new tab with the given URL.
 *
 * @param {string} url
 */
function openNewTab(url) {
    const win = window.open("");
    win.opener = null;
    win.location = url;
    win.focus();
}

/**
 * Scroll the page to the given element.
 *
 * @param {Element} element
 * @param {boolean} evenIfOnScreen
 */
function scrollPageTo(element, evenIfOnScreen) {
    const windowScrollPosition = window.scrollY;
    const windowHeight = document.documentElement.clientHeight;
    const viewportPosition = windowScrollPosition + windowHeight;
    const itemBottomPosition = element.offsetTop + element.offsetHeight;

    if (evenIfOnScreen || viewportPosition - itemBottomPosition < 0 || viewportPosition - element.offsetTop > windowHeight) {
        window.scrollTo(0, element.offsetTop - 10);
    }
}

/**
 * Attach a click event listener to elements matching the selector.
 *
 * @param {string} selector
 * @param {function} callback
 * @param {boolean} noPreventDefault
 */
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

/**
 * Attach an auxiliary click event listener to elements matching the selector.
 *
 * @param {string} selector
 * @param {function} callback
 * @param {boolean} noPreventDefault
 */
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

/**
 * Filter visible elements based on the selector.
 *
 * @param {string} selector
 * @returns {Array<Element>}
 */
function getVisibleElements(selector) {
    const elements = document.querySelectorAll(selector);
    return [...elements].filter((element) => element.offsetParent !== null);
}

/**
 * Get all visible entries on the current page.
 *
 * @return {Array<Element>}
 */
function getVisibleEntries() {
    return getVisibleElements(".items .item");
}

/**
 * Check if the current view is a list view.
 *
 * @returns {boolean}
 */
function isListView() {
    return document.querySelector(".items") !== null;
}

/**
 * Check if the current view is an entry view.
 *
 * @return {boolean}
 */
function isEntryView() {
    return document.querySelector("section.entry") !== null;
}

/**
 * Find the entry element for the given element.
 *
 * @returns {Element|null}
 */
function findEntry(element) {
    if (isListView()) {
        if (element) {
            return element.closest(".item");
        }
        return document.querySelector(".current-item");
    }
    return document.querySelector(".entry");
}

/**
 * Insert an icon label element into the parent element.
 *
 * @param {Element} parentElement The parent element to insert the icon label into.
 * @param {string} iconLabelText The text to display in the icon label.
 * @returns {void}
 */
function insertIconLabelElement(parentElement, iconLabelText) {
    const span = document.createElement('span');
    span.classList.add('icon-label');
    span.textContent = iconLabelText;
    parentElement.appendChild(span);
}

/**
 * Navigate to a specific page.
 *
 * @param {string} page - The page to navigate to.
 * @param {boolean} reloadOnFail - If true, reload the current page if the target page is not found.
 */
function goToPage(page, reloadOnFail = false) {
    const element = document.querySelector(":is(a, button)[data-page=" + page + "]");

    if (element) {
        document.location.href = element.href;
    } else if (reloadOnFail) {
        window.location.reload();
    }
}

/**
 * Navigate to the previous page.
 *
 * If the offset is a KeyboardEvent, it will navigate to the previous item in the list.
 * If the offset is a number, it will jump that many items in the list.
 * If the offset is TOP, it will jump to the first item in the list.
 * If the offset is BOTTOM, it will jump to the last item in the list.
 * If the current view is an entry view, it will redirect to the previous page.
 *
 * @param {number|KeyboardEvent} offset - How many items to jump for focus.
 */
function goToPreviousPage(offset) {
    if (offset instanceof KeyboardEvent) offset = -1;
    if (isListView()) {
        goToListItem(offset);
    } else {
        goToPage("previous");
    }
}

/**
 * Navigate to the next page.
 *
 * If the offset is a KeyboardEvent, it will navigate to the next item in the list.
 * If the offset is a number, it will jump that many items in the list.
 * If the offset is TOP, it will jump to the first item in the list.
 * If the offset is BOTTOM, it will jump to the last item in the list.
 * If the current view is an entry view, it will redirect to the next page.
 *
 * @param {number|KeyboardEvent} offset - How many items to jump for focus.
 */
function goToNextPage(offset) {
    if (offset instanceof KeyboardEvent) offset = 1;
    if (isListView()) {
        goToListItem(offset);
    } else {
        goToPage("next");
    }
}

/**
 * Navigate to the individual feed or feeds page.
 *
 * If the current view is an entry view, it will redirect to the feed link of the entry.
 * If the current view is a list view, it will redirect to the feeds page.
 */
function goToFeedOrFeedsPage() {
    if (isEntryView()) {
        goToFeedPage();
    } else {
        goToPage("feeds");
    }
}

/**
 * Navigate to the feed page of the current entry.
 *
 * If the current view is an entry view, it will redirect to the feed link of the entry.
 * If the current view is a list view, it will redirect to the feed link of the currently selected item.
 * If no feed link is available, it will do nothing.
 */
function goToFeedPage() {
    if (isEntryView()) {
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
 * Navigate to the add subscription page.
 *
 * @returns {void}
 */
function goToAddSubscriptionPage() {
    window.location.href = document.body.dataset.addSubscriptionUrl;
}

/**
 * Navigate to the next or previous item in the list.
 *
 * If the offset is TOP, it will jump to the first item in the list.
 * If the offset is BOTTOM, it will jump to the last item in the list.
 * If the offset is a number, it will jump that many items in the list.
 * If the current view is an entry view, it will redirect to the next or previous page.
 *
 * @param {number} offset - How many items to jump for focus.
 * @return {void}
 */
function goToListItem(offset) {
    const items = getVisibleEntries();
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

            // By default adjust selection by offset
            let itemOffset = (i + offset + items.length) % items.length;
            // Allow jumping to top or bottom
            if (offset === TOP) {
                itemOffset = 0;
            } else if (offset === BOTTOM) {
                itemOffset = items.length - 1;
            }
            const item = items[itemOffset];

            item.classList.add("current-item");
            scrollPageTo(item);
            item.focus();

            break;
        }
    }
}

/**
 * Handle the share action for the entry.
 *
 * If the share status is "shared", it will trigger the Web Share API.
 * If the share status is "share", it will send an Ajax request to fetch the share URL and then trigger the Web Share API.
 * If the Web Share API is not supported, it will redirect to the entry URL.
 */
async function handleShare() {
    const link = document.querySelector(':is(a, button)[data-share-status]');
    const title = document.querySelector(".entry-header > h1 > a");
    if (link.dataset.shareStatus === "shared") {
        await triggerWebShare(title, link.href);
    }
    else if (link.dataset.shareStatus === "share") {
        const request = new RequestBuilder(link.href);
        request.withCallback((r) => {
            // Ensure title is not null before passing to triggerWebShare
            triggerWebShare(title, r.url);
        });
        request.withHttpMethod("GET");
        request.execute();
    }
}

/**
 * Trigger the Web Share API to share the entry.
 *
 * If the Web Share API is not supported, it will redirect to the entry URL.
 *
 * @param {Element} title - The title element of the entry.
 * @param {string} url - The URL of the entry to share.
 */
async function triggerWebShare(title, url) {
    if (!navigator.canShare) {
        console.error("Your browser doesn't support the Web Share API.");
        window.location = url;
        return;
    }
    try {
        await navigator.share({
            title: title ? title.textContent : url,
            url: url
        });
    } catch (err) {
        console.error(err);
    }
    window.location.reload();
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
        menuToggleButton.setAttribute("aria-expanded", "false");
    } else {
        menuToggleButton.setAttribute("aria-expanded", "true");
    }
    menu.classList.toggle("js-menu-show");
}

// Handle click events for the main menu (<li> and <a>).
function onClickMainMenuListItem(event) {
    const element = event.target;

    if (element.tagName === "A") {
        window.location.href = element.getAttribute("href");
    } else {
        const linkElement = element.querySelector("a") || element.closest("a");
        window.location.href = linkElement.getAttribute("href");
    }
}

/**
 * This function changes the button label to the loading state and disables the button.
 *
 * @returns {void}
 */
function disableSubmitButtonsOnFormSubmit() {
    document.querySelectorAll("form").forEach((element) => {
        element.onsubmit = () => {
            const buttons = element.querySelectorAll("button[type=submit]");
            buttons.forEach((button) => {
                if (button.dataset.labelLoading) {
                    button.textContent = button.dataset.labelLoading;
                }
                button.disabled = true;
            });
        };
    });
}

// Show modal dialog with the list of keyboard shortcuts.
function showKeyboardShortcuts() {
    const template = document.getElementById("keyboard-shortcuts");
    ModalHandler.open(template.content, "dialog-title");
}

// Mark as read visible items of the current page.
function markPageAsRead() {
    const items = getVisibleEntries();
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
 *
 * @param {string} navigationDirection Navigation direction: "previous" or "next".
 * @param {Element} element Element that triggered the action.
 * @param {boolean} setToRead If true, set the entry to read instead of toggling the status.
 * @returns {void}
 */
function handleEntryStatus(navigationDirection, element, setToRead) {
    const toasting = !element;
    const currentEntry = findEntry(element);

    if (currentEntry) {
        if (!setToRead || currentEntry.querySelector(":is(a, button)[data-toggle-status]").dataset.value === "unread") {
            toggleEntryStatus(currentEntry, toasting);
        }
        if (isListView() && currentEntry.classList.contains('current-item')) {
            switch (navigationDirection) {
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

/**
 * Toggle the entry status between "read" and "unread".
 *
 * @param {Element} element The entry element to toggle the status for.
 * @param {boolean} toasting If true, show a toast notification after toggling the status.
 */
function toggleEntryStatus(element, toasting) {
    const entryID = parseInt(element.dataset.id, 10);
    const link = element.querySelector(":is(a, button)[data-toggle-status]");
    if (!link) {
        return;
    }

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
        insertIconLabelElement(link, label);
        link.dataset.value = newStatus;

        if (element.classList.contains("item-status-" + currentStatus)) {
            element.classList.remove("item-status-" + currentStatus);
            element.classList.add("item-status-" + newStatus);
        }

        if (isListView() && getVisibleEntries().length === 0) {
            window.location.reload();
        }
    });
}

/**
 * Mark the entry as read if it is currently unread.
 *
 * @param {Element} element The entry element to mark as read.
 */
function markEntryAsRead(element) {
    if (element.classList.contains("item-status-unread")) {
        element.classList.remove("item-status-unread");
        element.classList.add("item-status-read");

        const entryID = parseInt(element.dataset.id, 10);
        updateEntriesStatus([entryID], "read");
    }
}

/**
 * Handle the refresh of all feeds.
 *
 * This function redirects the user to the URL specified in the data-refresh-all-feeds-url attribute of the body element.
 */
function handleRefreshAllFeeds() {
    const refreshAllFeedsUrl = document.body.dataset.refreshAllFeedsUrl;
    if (refreshAllFeedsUrl) {
        window.location.href = refreshAllFeedsUrl;
    }
}

/**
 * Update the status of multiple entries.
 *
 * @param {Array<number>} entryIDs - The IDs of the entries to update.
 * @param {string} status - The new status to set for the entries (e.g., "read", "unread").
 */
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

/**
 * Handle save entry from list view and entry view.
 *
 * @param {Element} element
 */
function handleSaveEntry(element) {
    const toasting = !element;
    const currentEntry = findEntry(element);
    if (currentEntry) {
        saveEntry(currentEntry.querySelector(":is(a, button)[data-save-entry]"), toasting);
    }
}

/**
 * Save the entry by sending an Ajax request to the server.
 *
 * @param {Element} element The element that triggered the save action.
 * @param {boolean} toasting If true, show a toast notification after saving the entry.
 * @return {void}
 */
function saveEntry(element, toasting) {
    if (!element || element.dataset.completed) {
        return;
    }

    element.textContent = "";
    insertIconLabelElement(element, element.dataset.labelLoading);

    const request = new RequestBuilder(element.dataset.saveUrl);
    request.withCallback(() => {
        element.textContent = "";
        insertIconLabelElement(element, element.dataset.labelDone);
        element.dataset.completed = "true";
        if (toasting) {
            const iconElement = document.querySelector("template#icon-save");
            showToast(element.dataset.toastDone, iconElement);
        }
    });
    request.execute();
}

/**
 * Handle bookmarking an entry.
 *
 * @param {Element} element - The element that triggered the bookmark action.
 */
function handleBookmark(element) {
    const toasting = !element;
    const currentEntry = findEntry(element);
    if (currentEntry) {
        toggleBookmark(currentEntry, toasting);
    }
}

/**
 * Toggle the bookmark status of an entry.
 *
 * @param {Element} parentElement - The parent element containing the bookmark button.
 * @param {boolean} toasting - Whether to show a toast notification.
 * @returns {void}
 */
function toggleBookmark(parentElement, toasting) {
    const buttonElement = parentElement.querySelector(":is(a, button)[data-toggle-bookmark]");
    if (!buttonElement) {
        return;
    }

    buttonElement.textContent = "";
    insertIconLabelElement(buttonElement, buttonElement.dataset.labelLoading);

    const request = new RequestBuilder(buttonElement.dataset.bookmarkUrl);
    request.withCallback(() => {
        const currentStarStatus = buttonElement.dataset.value;
        const newStarStatus = currentStarStatus === "star" ? "unstar" : "star";

        let iconElement, label;
        if (currentStarStatus === "star") {
            iconElement = document.querySelector("template#icon-star");
            label = buttonElement.dataset.labelStar;
            if (toasting) {
                showToast(buttonElement.dataset.toastUnstar, iconElement);
            }
        } else {
            iconElement = document.querySelector("template#icon-unstar");
            label = buttonElement.dataset.labelUnstar;
            if (toasting) {
                showToast(buttonElement.dataset.toastStar, iconElement);
            }
        }

        buttonElement.replaceChildren(iconElement.content.cloneNode(true));
        insertIconLabelElement(buttonElement, label);
        buttonElement.dataset.value = newStarStatus;
    });
    request.execute();
}

/**
 * Handle fetching the original content of an entry.
 *
 * @returns {void}
 */
function handleFetchOriginalContent() {
    if (isListView()) {
        return;
    }

    const buttonElement = document.querySelector(":is(a, button)[data-fetch-content-entry]");
    if (!buttonElement) {
        return;
    }

    const previousElement = buttonElement.cloneNode(true);

    buttonElement.textContent = "";
    insertIconLabelElement(buttonElement, buttonElement.dataset.labelLoading);

    const request = new RequestBuilder(buttonElement.dataset.fetchContentUrl);
    request.withCallback((response) => {
        buttonElement.textContent = '';
        buttonElement.appendChild(previousElement);

        response.json().then((data) => {
            if (data.hasOwnProperty("content") && data.hasOwnProperty("reading_time")) {
                document.querySelector(".entry-content").innerHTML = ttpolicy.createHTML(data.content);
                const entryReadingtimeElement = document.querySelector(".entry-reading-time");
                if (entryReadingtimeElement) {
                    entryReadingtimeElement.textContent = data.reading_time;
                }
            }
        });
    });
    request.execute();
}

/**
 * Open the original link of an entry.
 *
 * @param {boolean} openLinkInCurrentTab - Whether to open the link in the current tab.
 * @returns {void}
 */
function openOriginalLink(openLinkInCurrentTab) {
    const entryLink = document.querySelector(".entry h1 a");
    if (entryLink !== null) {
        if (openLinkInCurrentTab) {
            window.location.href = entryLink.getAttribute("href");
        } else {
            openNewTab(entryLink.getAttribute("href"));
        }
        return;
    }

    const currentItemOriginalLink = document.querySelector(".current-item :is(a, button)[data-original-link]");
    if (currentItemOriginalLink !== null) {
        openNewTab(currentItemOriginalLink.getAttribute("href"));

        const currentItem = document.querySelector(".current-item");
        // If we are not on the list of starred items, move to the next item
        if (document.location.href !== document.querySelector(':is(a, button)[data-page=starred]').href) {
            goToListItem(1);
        }
        markEntryAsRead(currentItem);
    }
}

/**
 * Open the comments link of an entry.
 *
 * @param {boolean} openLinkInCurrentTab - Whether to open the link in the current tab.
 * @returns {void}
 */
function openCommentLink(openLinkInCurrentTab) {
    if (!isListView()) {
        const entryLink = document.querySelector(":is(a, button)[data-comments-link]");
        if (entryLink !== null) {
            if (openLinkInCurrentTab) {
                window.location.href = entryLink.getAttribute("href");
            } else {
                openNewTab(entryLink.getAttribute("href"));
            }
        }
    } else {
        const currentItemCommentsLink = document.querySelector(".current-item :is(a, button)[data-comments-link]");
        if (currentItemCommentsLink !== null) {
            openNewTab(currentItemCommentsLink.getAttribute("href"));
        }
    }
}

/**
 * Open the selected item in the current view.
 *
 * If the current view is a list view, it will navigate to the link of the currently selected item.
 * If the current view is an entry view, it will navigate to the link of the entry.
 */
function openSelectedItem() {
    const currentItemLink = document.querySelector(".current-item .item-title a");
    if (currentItemLink !== null) {
        window.location.href = currentItemLink.getAttribute("href");
    }
}

/**
 * Unsubscribe from the feed of the currently selected item.
 */
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
 * Scroll the page to the currently selected item.
 */
function scrollToCurrentItem() {
    const currentItem = document.querySelector(".current-item");
    if (currentItem !== null) {
        scrollPageTo(currentItem, true);
    }
}

/**
 * Decrement the unread counter by a specified amount.
 *
 * @param {number} n - The amount to decrement the counter by.
 */
function decrementUnreadCounter(n) {
    updateUnreadCounterValue((current) => {
        return current - n;
    });
}

/**
 * Increment the unread counter by a specified amount.
 *
 * @param {number} n - The amount to increment the counter by.
 */
function incrementUnreadCounter(n) {
    updateUnreadCounterValue((current) => {
        return current + n;
    });
}

/**
 * Update the unread counter value.
 *
 * @param {function} callback - The function to call with the old value.
 */
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

/**
 * Handle confirmation messages for actions that require user confirmation.
 *
 * This function modifies the link element to show a confirmation question with "Yes" and "No" buttons.
 * If the user clicks "Yes", it calls the provided callback with the URL and redirect URL.
 * If the user clicks "No", it either redirects to a no-action URL or restores the link element.
 *
 * @param {Element} linkElement - The link or button element that triggered the confirmation.
 * @param {function} callback - The callback function to execute if the user confirms the action.
 * @returns {void}
 */
function handleConfirmationMessage(linkElement, callback) {
    if (linkElement.tagName !== 'A' && linkElement.tagName !== "BUTTON") {
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

/**
 * Show a toast notification.
 *
 * @param {string} toastMessage - The label to display in the toast.
 * @param {Element} iconElement - The icon element to display in the toast.
 * @returns {void}
 */
function showToast(toastMessage, iconElement) {
    if (!toastMessage || !iconElement) {
        return;
    }

    const toastMsgElement = document.getElementById("toast-msg");
    toastMsgElement.replaceChildren(iconElement.content.cloneNode(true));
    insertIconLabelElement(toastMsgElement, toastMessage);

    const toastElementWrapper = document.getElementById("toast-wrapper");
    toastElementWrapper.classList.remove('toast-animate');
    setTimeout(() => {
        toastElementWrapper.classList.add('toast-animate');
    }, 100);
}

/**
 * Check if the player is actually playing a media
 *
 * @param mediaElement the player element itself
 * @returns {boolean}
 */
function isPlayerPlaying(mediaElement) {
    return mediaElement &&
        mediaElement.currentTime > 0 &&
        !mediaElement.paused &&
        !mediaElement.ended &&
        mediaElement.readyState > 2; // https://developer.mozilla.org/en-US/docs/Web/API/HTMLMediaElement/readyState
}

/**
 * Handle player progression save and mark as read on completion.
 *
 * This function is triggered on the `timeupdate` event of the media player.
 * It saves the current playback position and marks the entry as read if the completion percentage is reached.
 *
 * @param {Element} playerElement The media player element (audio or video).
 */
function handlePlayerProgressionSaveAndMarkAsReadOnCompletion(playerElement) {
    if (!isPlayerPlaying(playerElement)) {
        return;
    }

    const currentPositionInSeconds = Math.floor(playerElement.currentTime);
    const lastKnownPositionInSeconds = parseInt(playerElement.dataset.lastPosition, 10);
    const markAsReadOnCompletion = parseFloat(playerElement.dataset.markReadOnCompletion);
    const recordInterval = 10;

    // We limit the number of update to only one by interval. Otherwise, we would have multiple update per seconds
    if (currentPositionInSeconds >= (lastKnownPositionInSeconds + recordInterval) ||
        currentPositionInSeconds <= (lastKnownPositionInSeconds - recordInterval)
    ) {
        playerElement.dataset.lastPosition = currentPositionInSeconds.toString();

        const request = new RequestBuilder(playerElement.dataset.saveUrl);
        request.withBody({ progression: currentPositionInSeconds });
        request.execute();

        // Handle the mark as read on completion
        if (markAsReadOnCompletion >= 0 && playerElement.duration > 0) {
            const completion =  currentPositionInSeconds / playerElement.duration;
            if (completion >= markAsReadOnCompletion) {
                handleEntryStatus("none", document.querySelector(":is(a, button)[data-toggle-status]"), true);
            }
        }
    }
}

/**
 * Handle media control actions like seeking and changing playback speed.
 *
 * This function is triggered by clicking on media control buttons.
 * It adjusts the playback position or speed of media elements with the same enclosure ID.
 *
 * @param {Element} mediaPlayerButtonElement
 */
function handleMediaControlButtonClick(mediaPlayerButtonElement) {
    const actionType = mediaPlayerButtonElement.dataset.enclosureAction;
    const actionValue = parseFloat(mediaPlayerButtonElement.dataset.actionValue);
    const enclosureID = mediaPlayerButtonElement.dataset.enclosureId;
    const mediaElements = document.querySelectorAll(`audio[data-enclosure-id="${enclosureID}"],video[data-enclosure-id="${enclosureID}"]`);
    const speedIndicatorElements = document.querySelectorAll(`span.speed-indicator[data-enclosure-id="${enclosureID}"]`);
    mediaElements.forEach((mediaElement) => {
        switch (actionType) {
        case "seek":
            mediaElement.currentTime = Math.max(mediaElement.currentTime + actionValue, 0);
            break;
        case "speed":
            // 0.25 was chosen because it will allow to get back to 1x in two "faster" clicks.
            // A lower value would result in a playback rate of 0, effectively pausing playback.
            mediaElement.playbackRate = Math.max(0.25, mediaElement.playbackRate + actionValue);
            speedIndicatorElements.forEach((speedIndicatorElement) => {
                speedIndicatorElement.innerText = `${mediaElement.playbackRate.toFixed(2)}x`;
            });
            break;
        case "speed-reset":
            mediaElement.playbackRate = actionValue ;
            speedIndicatorElements.forEach((speedIndicatorElement) => {
                // Two digit precision to ensure we always have the same number of characters (4) to avoid controls moving when clicking buttons because of more or less characters.
                // The trick only works on rates less than 10, but it feels an acceptable trade-off considering the feature
                speedIndicatorElement.innerText = `${mediaElement.playbackRate.toFixed(2)}x`;
            });
            break;
        }
    });
}

/**
 * Initialize media player event handlers.
 */
function initializeMediaPlayerHandlers() {
    document.querySelectorAll("button[data-enclosure-action]").forEach((element) => {
        element.addEventListener("click", () => handleMediaControlButtonClick(element));
    });

    // Set playback from the last position if available
    document.querySelectorAll("audio[data-last-position],video[data-last-position]").forEach((element) => {
        if (element.dataset.lastPosition) {
            element.currentTime = element.dataset.lastPosition;
        }
        element.ontimeupdate = () => handlePlayerProgressionSaveAndMarkAsReadOnCompletion(element);
    });

    // Set playback speed from the data attribute if available
    document.querySelectorAll("audio[data-playback-rate],video[data-playback-rate]").forEach((element) => {
        if (element.dataset.playbackRate) {
            element.playbackRate = element.dataset.playbackRate;
            if (element.dataset.enclosureId) {
                document.querySelectorAll(`span.speed-indicator[data-enclosure-id="${element.dataset.enclosureId}"]`).forEach((speedIndicatorElement) => {
                    speedIndicatorElement.innerText = `${parseFloat(element.dataset.playbackRate).toFixed(2)}x`;
                });
            }
        }
    });
}