// Sentinel values for specific list navigation.
const TOP = 9999;
const BOTTOM = -9999;

// Simple Polyfill for browsers that don't support Trusted Types
// See https://caniuse.com/?search=trusted%20types
if (!window.trustedTypes || !trustedTypes.createPolicy) {
    window.trustedTypes = {
        createPolicy: (name, policy) => ({
            createScriptURL: src => src,
            createHTML: html => html,
        })
    };
}

/**
 * Send a POST request to the specified URL with the given body.
 *
 * @param {string} url - The URL to send the request to.
 * @param {Object} [body] - The body of the request (optional).
 * @returns {Promise<Response>} The response from the fetch request.
 */
function sendPOSTRequest(url, body = null) {
    const options = {
        method: "POST",
        headers: {
            "X-Csrf-Token": document.body.dataset.csrfToken || ""
        }
    };

    if (body !== null) {
        options.headers["Content-Type"] = "application/json";
        options.body = JSON.stringify(body);
    }

    return fetch(url, options);
}

/**
 * Open a new tab with the given URL.
 *
 * @param {string} url
 */
function openNewTab(url) {
    const win = window.open(url, "_blank", "noreferrer");
    if (win) win.focus();
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
 * Create an icon label element with the given text.
 *
 * @param {string} labelText - The text to display in the icon label.
 * @returns {Element} The created icon label element.
 */
function createIconLabelElement(labelText) {
    const labelElement = document.createElement("span");
    labelElement.classList.add("icon-label");
    labelElement.textContent = labelText;
    return labelElement;
}

/**
 * Set the icon and label element in the parent element.
 *
 * @param {Element} parentElement - The parent element to insert the icon and label into.
 * @param {string} iconName - The name of the icon to display.
 * @param {string} labelText - The text to display in the label.
 */
function setIconAndLabelElement(parentElement, iconName, labelText) {
    const iconElement = document.querySelector(`template#icon-${iconName}`);
    if (iconElement) {
        const iconClone = iconElement.content.cloneNode(true);
        parentElement.textContent = ""; // Clear existing content
        parentElement.appendChild(iconClone);
    }

    if (labelText) {
        const labelElement = createIconLabelElement(labelText);
        parentElement.appendChild(labelElement);
    }
}

/**
 * Set the button to a loading state and return a clone of the original button element.
 *
 * @param {Element} buttonElement - The button element to set to loading state.
 * @return {Element} The original button element cloned before modification.
 */
function setButtonToLoadingState(buttonElement) {
    const originalButtonElement = buttonElement.cloneNode(true);

    buttonElement.textContent = "";
    buttonElement.appendChild(createIconLabelElement(buttonElement.dataset.labelLoading));

    return originalButtonElement;
}

/**
 * Restore the button to its original state.
 *
 * @param {Element} buttonElement The button element to restore.
 * @param {Element} originalButtonElement The original button element to restore from.
 * @returns {void}
 */
function restoreButtonState(buttonElement, originalButtonElement) {
    buttonElement.textContent = "";
    buttonElement.appendChild(originalButtonElement);
}

/**
 * Set the button to a saved state.
 *
 * @param {Element} buttonElement The button element to set to saved state.
 */
function setButtonToSavedState(buttonElement) {
    buttonElement.dataset.completed = "true";
    setIconAndLabelElement(buttonElement, "save", buttonElement.dataset.labelDone);
}

/**
 * Set the star button state.
 *
 * @param {Element} buttonElement - The button element to update.
 * @param {string} newState - The new state to set ("star" or "unstar").
 */
function setStarredButtonState(buttonElement, newState) {
    buttonElement.dataset.value = newState;
    const iconType = newState === "star" ? "unstar" : "star";
    setIconAndLabelElement(buttonElement, iconType, buttonElement.dataset[newState === "star" ? "labelUnstar" : "labelStar"]);
}

/**
 * Set the read status button state.
 *
 * @param {Element} buttonElement - The button element to update.
 * @param {string} newState - The new state to set ("read" or "unread").
 */
function setReadStatusButtonState(buttonElement, newState) {
    buttonElement.dataset.value = newState;
    const iconType = newState === "read" ? "unread" : "read";
    setIconAndLabelElement(buttonElement, iconType, buttonElement.dataset[newState === "read" ? "labelUnread" : "labelRead"]);
}

/**
 * Show a toast notification.
 *
 * @param {string} iconType - The type of icon to display.
 * @param {string} notificationMessage - The message to display in the toast.
 * @returns {void}
 */
function showToastNotification(iconType, notificationMessage) {
    const toastMsgElement = document.createElement("span");
    toastMsgElement.id = "toast-msg";

    setIconAndLabelElement(toastMsgElement, iconType, notificationMessage);

    const toastElementWrapper = document.createElement("div");
    toastElementWrapper.id = "toast-wrapper";
    toastElementWrapper.setAttribute("role", "alert");
    toastElementWrapper.setAttribute("aria-live", "assertive");
    toastElementWrapper.setAttribute("aria-atomic", "true");
    toastElementWrapper.appendChild(toastMsgElement);
    toastElementWrapper.addEventListener("animationend", () => {
        toastElementWrapper.remove();
    });

    document.body.appendChild(toastElementWrapper);

    setTimeout(() => toastElementWrapper.classList.add("toast-animate"), 100);
}

/**
 * Navigate to a specific page.
 *
 * @param {string} page - The page to navigate to.
 * @param {boolean} reloadOnFail - If true, reload the current page if the target page is not found.
 */
function goToPage(page, reloadOnFail = false) {
    const element = document.querySelector(`:is(a, button)[data-page=${page}]`);

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

    const currentItem = document.querySelector(".current-item");

    // If no current item exists, select the first item
    if (!currentItem) {
        items[0].classList.add("current-item");
        items[0].focus();
        scrollPageTo(items[0]);
        return;
    }

    // Find the index of the current item
    const currentIndex = items.indexOf(currentItem);
    if (currentIndex === -1) {
        // Current item not found in visible items, select first item
        currentItem.classList.remove("current-item");
        items[0].classList.add("current-item");
        items[0].focus();
        scrollPageTo(items[0]);
        return;
    }

    // Calculate the new item index
    let newIndex;
    if (offset === TOP) {
        newIndex = 0;
    } else if (offset === BOTTOM) {
        newIndex = items.length - 1;
    } else {
        newIndex = (currentIndex + offset + items.length) % items.length;
    }

    // Update selection if moving to a different item
    if (newIndex !== currentIndex) {
        const newItem = items[newIndex];

        currentItem.classList.remove("current-item");
        newItem.classList.add("current-item");
        newItem.focus();
        scrollPageTo(newItem);
    }
}

/**
 * Handle the share action for the entry.
 *
 * If the share status is "shared", it will trigger the Web Share API.
 * If the share status is "share", it will send an Ajax request to fetch the share URL and then trigger the Web Share API.
 * If the Web Share API is not supported, it will redirect to the entry URL.
 */
async function handleEntryShareAction() {
    const link = document.querySelector(':is(a, button)[data-share-status]');
    if (link.dataset.shareStatus === "shared") {
        const title = document.querySelector(".entry-header > h1 > a");
        const url = link.href;

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
    }
}

/**
 * Toggle the ARIA attributes on the main menu based on the viewport width.
 */
function toggleAriaAttributesOnMainMenu() {
    const logoElement = document.querySelector(".logo");
    const homePageLinkElement = document.querySelector(".logo > a");

    if (!logoElement || !homePageLinkElement) return;

    const isMobile = document.documentElement.clientWidth < 650;

    if (isMobile) {
        const navMenuElement = document.getElementById("header-menu");
        const isExpanded = navMenuElement?.classList.contains("js-menu-show") ?? false;
        const toggleButtonLabel = logoElement.getAttribute("data-toggle-button-label");

        // Set mobile menu button attributes
        Object.assign(logoElement, {
            role: "button",
            tabIndex: 0,
            ariaLabel: toggleButtonLabel,
            ariaExpanded: isExpanded.toString()
        });
        homePageLinkElement.tabIndex = -1;
    } else {
        // Remove mobile menu button attributes
        ["role", "tabindex", "aria-expanded", "aria-label"].forEach(attr => {
            logoElement.removeAttribute(attr);
        });
        homePageLinkElement.removeAttribute("tabindex");
    }
}

/**
 * Toggle the main menu dropdown.
 *
 * @param {Event} event - The event object.
 */
function toggleMainMenuDropdown(event) {
    // Only handle Enter, Space, or click events
    if (event.type === "keydown" && !["Enter", " "].includes(event.key)) {
        return;
    }

    // Prevent default only if element has role attribute (mobile menu button)
    if (event.currentTarget.getAttribute("role")) {
        event.preventDefault();
    }

    const navigationMenu = document.querySelector(".header nav ul");
    const menuToggleButton = document.querySelector(".logo");

    if (!navigationMenu || !menuToggleButton) {
        return;
    }

    const isShowing = navigationMenu.classList.toggle("js-menu-show");
    menuToggleButton.setAttribute("aria-expanded", isShowing.toString());
}

/**
 * Initialize the main menu handlers.
 */
function initializeMainMenuHandlers() {
    toggleAriaAttributesOnMainMenu();
    window.addEventListener("resize", toggleAriaAttributesOnMainMenu, { passive: true });

    const logoElement = document.querySelector(".logo");
    if (logoElement) {
        logoElement.addEventListener("click", toggleMainMenuDropdown);
        logoElement.addEventListener("keydown", toggleMainMenuDropdown);
    }

    onClick(".header nav li", (event) => {
        const linkElement = event.target.closest("a") || event.target.querySelector("a");
        if (linkElement && !event.ctrlKey && !event.shiftKey && !event.metaKey) {
            event.preventDefault();
            window.location.href = linkElement.getAttribute("href");
        }
    }, true);
}

/**
 * This function changes the button label to the loading state and disables the button.
 *
 * @returns {void}
 */
function initializeFormHandlers() {
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

/**
 * Show the keyboard shortcuts modal.
 */
function showKeyboardShortcutsAction() {
    document.getElementById("keyboard-shortcuts-modal").showModal();
}

/**
 * Mark all visible entries on the current page as read.
 */
function markPageAsReadAction() {
    const items = getVisibleEntries();
    if (items.length === 0) return;

    const entryIDs = items.map((element) => {
        element.classList.add("item-status-read");
        return parseInt(element.dataset.id, 10);
    });

    updateEntriesStatus(entryIDs, "read", () => {
        const element = document.querySelector(":is(a, button)[data-action=markPageAsRead]");
        const showOnlyUnread = element?.dataset.showOnlyUnread || false;

        if (showOnlyUnread) {
            window.location.reload();
        } else {
            goToPage("next", true);
        }
    });
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
    const currentEntry = findEntry(element);

    if (currentEntry) {
        if (!setToRead || currentEntry.querySelector(":is(a, button)[data-toggle-status]").dataset.value === "unread") {
            toggleEntryStatus(currentEntry, isEntryView());
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
    const buttonElement = element.querySelector(":is(a, button)[data-toggle-status]");
    if (!buttonElement) return;

    const currentStatus = buttonElement.dataset.value;
    const newStatus = currentStatus === "read" ? "unread" : "read";

    setButtonToLoadingState(buttonElement);

    updateEntriesStatus([entryID], newStatus, () => {
        setReadStatusButtonState(buttonElement, newStatus);

        if (toasting) {
            showToastNotification(newStatus, currentStatus === "read" ? buttonElement.dataset.toastUnread : buttonElement.dataset.toastRead);
        }

        element.classList.replace(`item-status-${currentStatus}`, `item-status-${newStatus}`);

        if (isListView() && getVisibleEntries().length === 0) {
            window.location.reload();
        }
    });
}

/**
 * Handle the refresh of all feeds.
 *
 * This function redirects the user to the URL specified in the data-refresh-all-feeds-url attribute of the body element.
 */
function handleRefreshAllFeedsAction() {
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
    sendPOSTRequest(url, { entry_ids: entryIDs, status: status }).then((resp) => {
        resp.json().then(count => {
            if (callback) {
                callback(resp);
            }
            updateUnreadCounterValue(status === "read" ? -count : count);
        });
    });
}

/**
 * Handle save entry from list view and entry view.
 *
 * @param {Element|null} element - The element that triggered the save action (optional).
 */
function handleSaveEntryAction(element = null) {
    const currentEntry = findEntry(element);
    if (!currentEntry) return;

    const buttonElement = currentEntry.querySelector(":is(a, button)[data-save-entry]");
    if (!buttonElement || buttonElement.dataset.completed) return;

    setButtonToLoadingState(buttonElement);

    sendPOSTRequest(buttonElement.dataset.saveUrl).then(() => {
        setButtonToSavedState(buttonElement);
        if (isEntryView()) {
            showToastNotification("save", buttonElement.dataset.toastDone);
        }
    });
}

/**
 * Handle starring an entry.
 *
 * @param {Element} element - The element that triggered the star action.
 */
function handleStarAction(element) {
    const currentEntry = findEntry(element);
    if (!currentEntry) return;

    const buttonElement = currentEntry.querySelector(":is(a, button)[data-toggle-starred]");
    if (!buttonElement) return;

    setButtonToLoadingState(buttonElement);

    sendPOSTRequest(buttonElement.dataset.starUrl).then(() => {
        const currentState = buttonElement.dataset.value;
        const isStarred = currentState === "star";
        const newStarStatus = isStarred ? "unstar" : "star";

        setStarredButtonState(buttonElement, newStarStatus);

        if (isEntryView()) {
            showToastNotification(currentState, buttonElement.dataset[isStarred ? "toastUnstar" : "toastStar"]);
        }
    });
}

/**
 * Handle fetching the original content of an entry.
 *
 * @returns {void}
 */
function handleFetchOriginalContentAction() {
    if (isListView()) return;

    const buttonElement = document.querySelector(":is(a, button)[data-fetch-content-entry]");
    if (!buttonElement) return;

    const originalButtonElement = setButtonToLoadingState(buttonElement);

    sendPOSTRequest(buttonElement.dataset.fetchContentUrl).then((response) => {
        restoreButtonState(buttonElement, originalButtonElement);

        response.json().then((data) => {
            if (data.content && data.reading_time) {
                const ttpolicy = trustedTypes.createPolicy('html', {createHTML: html => html});
                document.querySelector(".entry-content").innerHTML = ttpolicy.createHTML(data.content);
                const entryReadingtimeElement = document.querySelector(".entry-reading-time");
                if (entryReadingtimeElement) {
                    entryReadingtimeElement.textContent = data.reading_time;
                }
            }
        });
    });
}

/**
 * Open the original link of an entry.
 *
 * @param {boolean} openLinkInCurrentTab - Whether to open the link in the current tab.
 * @returns {void}
 */
function openOriginalLinkAction(openLinkInCurrentTab) {
    if (isEntryView()) {
        openOriginalLinkFromEntryView(openLinkInCurrentTab);
    } else if (isListView()) {
        openOriginalLinkFromListView();
    }
}

/**
 * Open the original link from entry view.
 *
 * @param {boolean} openLinkInCurrentTab - Whether to open the link in the current tab.
 * @returns {void}
 */
function openOriginalLinkFromEntryView(openLinkInCurrentTab) {
    const entryLink = document.querySelector(".entry h1 a");
    if (!entryLink) return;

    const url = entryLink.getAttribute("href");
    if (openLinkInCurrentTab) {
        window.location.href = url;
    } else {
        openNewTab(url);
    }
}

/**
 * Open the original link from list view.
 *
 * @returns {void}
 */
function openOriginalLinkFromListView() {
    const currentItem = document.querySelector(".current-item");
    const originalLink = currentItem?.querySelector(":is(a, button)[data-original-link]");

    if (!currentItem || !originalLink) return;

    // Open the link
    openNewTab(originalLink.getAttribute("href"));

    // Don't navigate or mark as read on starred page
    const isStarredPage = document.location.href === document.querySelector(':is(a, button)[data-page=starred]').href;
    if (isStarredPage) return;

    // Navigate to next item
    goToListItem(1);

    // Mark as read if currently unread
    if (currentItem.classList.replace("item-status-unread", "item-status-read")) {
        const entryID = parseInt(currentItem.dataset.id, 10);
        updateEntriesStatus([entryID], "read");
    }
}

/**
 * Open the comments link of an entry.
 *
 * @param {boolean} openLinkInCurrentTab - Whether to open the link in the current tab.
 * @returns {void}
 */
function openCommentLinkAction(openLinkInCurrentTab) {
    const entryLink = document.querySelector(isListView() ? ".current-item :is(a, button)[data-comments-link]" : ":is(a, button)[data-comments-link]");

    if (entryLink) {
        if (openLinkInCurrentTab) {
            window.location.href = entryLink.getAttribute("href");
        } else {
            openNewTab(entryLink.getAttribute("href"));
        }
    }
}

/**
 * Open the selected item in the current view.
 *
 * If the current view is a list view, it will navigate to the link of the currently selected item.
 * If the current view is an entry view, it will navigate to the link of the entry.
 */
function openSelectedItemAction() {
    const currentItemLink = document.querySelector(".current-item .item-title a");
    if (currentItemLink) {
        window.location.href = currentItemLink.getAttribute("href");
    }
}

/**
 * Unsubscribe from the feed of the currently selected item.
 */
function handleRemoveFeedAction() {
    const unsubscribeLink = document.querySelector("[data-action=remove-feed]");
    if (unsubscribeLink) {
        sendPOSTRequest(unsubscribeLink.dataset.url).then(() => {
            window.location.href = unsubscribeLink.dataset.redirectUrl || window.location.href;
        });
    }
}

/**
 * Scroll the page to the currently selected item.
 */
function scrollToCurrentItemAction() {
    const currentItem = document.querySelector(".current-item");
    if (currentItem) {
        scrollPageTo(currentItem, true);
    }
}

/**
 * Update the unread counter value.
 *
 * @param {number} delta - The amount to change the counter by.
 */
function updateUnreadCounterValue(delta) {
    document.querySelectorAll("span.unread-counter").forEach((element) => {
        const oldValue = parseInt(element.textContent, 10);
        element.textContent = oldValue + delta;
    });

    if (window.location.href.endsWith('/unread')) {
        const oldValue = parseInt(document.title.split('(')[1], 10);
        const newValue = oldValue + delta;
        document.title = document.title.replace(/\(\d+\)/, `(${newValue})`);
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
    questionElement.appendChild(document.createTextNode(`${linkElement.dataset.labelQuestion} `));
    questionElement.appendChild(yesElement);
    questionElement.appendChild(document.createTextNode(", "));
    questionElement.appendChild(noElement);

    containerElement.appendChild(questionElement);
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

        sendPOSTRequest(playerElement.dataset.saveUrl, { progression: currentPositionInSeconds });

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

/**
 * Initialize the service worker and PWA installation prompt.
 */
function initializeServiceWorker() {
    // Register service worker if supported
    if ("serviceWorker" in navigator) {
        const serviceWorkerURL = document.body.dataset.serviceWorkerUrl;
        if (serviceWorkerURL) {
            const ttpolicy = trustedTypes.createPolicy('url', {createScriptURL: src => src});
            navigator.serviceWorker.register(ttpolicy.createScriptURL(serviceWorkerURL), {
                type: "module"
            }).catch((error) => {
                console.error("Service Worker registration failed:", error);
            });
        }
    }

    // PWA installation prompt handling
    window.addEventListener("beforeinstallprompt", (event) => {
        let deferredPrompt = event;
        const promptHomeScreen = document.getElementById("prompt-home-screen");
        const btnAddToHomeScreen = document.getElementById("btn-add-to-home-screen");

        if (!promptHomeScreen || !btnAddToHomeScreen) return;

        promptHomeScreen.style.display = "block";

        btnAddToHomeScreen.addEventListener("click", (event) => {
            event.preventDefault();
            deferredPrompt.prompt();
            deferredPrompt.userChoice.then(() => {
                deferredPrompt = null;
                promptHomeScreen.style.display = "none";
            });
        });
    });
}

/**
 * Initialize WebAuthn handlers if supported.
 */
function initializeWebAuthn() {
    if (typeof WebAuthnHandler !== 'function') return;
    if (!WebAuthnHandler.isWebAuthnSupported()) return;

    const webauthnHandler = new WebAuthnHandler();

    // Setup delete credentials handler
    onClick("#webauthn-delete", () => { webauthnHandler.removeAllCredentials(); });

    // Setup registration
    const registerButton = document.getElementById("webauthn-register");
    if (registerButton) {
        registerButton.disabled = false;
        onClick("#webauthn-register", () => {
            webauthnHandler.register().catch((err) => WebAuthnHandler.showErrorMessage(err));
        });
    }

    // Setup login
    const loginButton = document.getElementById("webauthn-login");
    const usernameField = document.getElementById("form-username");

    if (loginButton && usernameField) {
        const abortController = new AbortController();
        loginButton.disabled = false;

        onClick("#webauthn-login", () => {
            abortController.abort();
            webauthnHandler.login(usernameField.value).catch(err => WebAuthnHandler.showErrorMessage(err));
        });

        webauthnHandler.conditionalLogin(abortController).catch(err => WebAuthnHandler.showErrorMessage(err));
    }
}

/**
 * Initialize keyboard shortcuts for navigation and actions.
 */
function initializeKeyboardShortcuts() {
    if (document.querySelector("body[data-disable-keyboard-shortcuts=true]")) return;

    const keyboardHandler = new KeyboardHandler();

    // Navigation shortcuts
    keyboardHandler.on("g u", () => goToPage("unread"));
    keyboardHandler.on("g b", () => goToPage("starred"));
    keyboardHandler.on("g h", () => goToPage("history"));
    keyboardHandler.on("g f", goToFeedOrFeedsPage);
    keyboardHandler.on("g c", () => goToPage("categories"));
    keyboardHandler.on("g s", () => goToPage("settings"));
    keyboardHandler.on("g g", () => goToPreviousPage(TOP));
    keyboardHandler.on("G", () => goToNextPage(BOTTOM));
    keyboardHandler.on("/", () => goToPage("search"));

    // Item navigation
    keyboardHandler.on("ArrowLeft", goToPreviousPage);
    keyboardHandler.on("ArrowRight", goToNextPage);
    keyboardHandler.on("k", goToPreviousPage);
    keyboardHandler.on("p", goToPreviousPage);
    keyboardHandler.on("j", goToNextPage);
    keyboardHandler.on("n", goToNextPage);
    keyboardHandler.on("h", () => goToPage("previous"));
    keyboardHandler.on("l", () => goToPage("next"));
    keyboardHandler.on("z t", scrollToCurrentItemAction);

    // Item actions
    keyboardHandler.on("o", openSelectedItemAction);
    keyboardHandler.on("Enter", () => openSelectedItemAction());
    keyboardHandler.on("v", () => openOriginalLinkAction(false));
    keyboardHandler.on("V", () => openOriginalLinkAction(true));
    keyboardHandler.on("c", () => openCommentLinkAction(false));
    keyboardHandler.on("C", () => openCommentLinkAction(true));

    // Entry management
    keyboardHandler.on("m", () => handleEntryStatus("next"));
    keyboardHandler.on("M", () => handleEntryStatus("previous"));
    keyboardHandler.on("A", markPageAsReadAction);
    keyboardHandler.on("s", () => handleSaveEntryAction());
    keyboardHandler.on("d", handleFetchOriginalContentAction);
    keyboardHandler.on("f", () => handleStarAction());

    // Feed actions
    keyboardHandler.on("F", goToFeedPage);
    keyboardHandler.on("R", handleRefreshAllFeedsAction);
    keyboardHandler.on("+", goToAddSubscriptionPage);
    keyboardHandler.on("#", handleRemoveFeedAction);

    // UI actions
    keyboardHandler.on("?", showKeyboardShortcutsAction);
    keyboardHandler.on("a", () => {
        const enclosureElement = document.querySelector('.entry-enclosures');
        if (enclosureElement) {
            enclosureElement.toggleAttribute('open');
        }
    });

    keyboardHandler.listen();
}

/**
 * Initialize touch handler for mobile devices.
 */
function initializeTouchHandler() {
    if ( "ontouchstart" in window || navigator.maxTouchPoints > 0) {
        const touchHandler = new TouchHandler();
        touchHandler.listen();
    }
}

/**
 * Initialize click handlers for various UI elements.
 */
function initializeClickHandlers() {
    // Entry actions
    onClick(":is(a, button)[data-save-entry]", (event) => handleSaveEntryAction(event.target));
    onClick(":is(a, button)[data-toggle-starred]", (event) => handleStarAction(event.target));
    onClick(":is(a, button)[data-toggle-status]", (event) => handleEntryStatus("next", event.target));
    onClick(":is(a, button)[data-fetch-content-entry]", handleFetchOriginalContentAction);
    onClick(":is(a, button)[data-share-status]", handleEntryShareAction);

    // Page actions with confirmation
    onClick(":is(a, button)[data-action=markPageAsRead]", (event) => handleConfirmationMessage(event.target, markPageAsReadAction));

    // Generic confirmation handler
    onClick(":is(a, button)[data-confirm]", (event) => {
        handleConfirmationMessage(event.target, (url, redirectURL) => {
            sendPOSTRequest(url).then((response) => {
                if (redirectURL) {
                    window.location.href = redirectURL;
                } else if (response?.redirected && response.url) {
                    window.location.href = response.url;
                } else {
                    window.location.reload();
                }
            });
        });
    });

    // Original link handlers (both click and middle-click)
    const handleOriginalLink = (event) => handleEntryStatus("next", event.target, true);

    onClick("a[data-original-link='true']", handleOriginalLink, true);
    onAuxClick("a[data-original-link='true']", (event) => {
        if (event.button === 1) {
            handleOriginalLink(event);
        }
    }, true);
}

// Initialize application handlers
initializeMainMenuHandlers();
initializeFormHandlers();
initializeMediaPlayerHandlers();
initializeWebAuthn();
initializeKeyboardShortcuts();
initializeTouchHandler();
initializeClickHandlers();
initializeServiceWorker();

// Reload the page if it was restored from the back-forward cache and mark entries as read is enabled.
window.addEventListener("pageshow", (event) => {
    if (event.persisted && document.body.dataset.markAsReadOnView === "true") {
        location.reload();
    }
});
