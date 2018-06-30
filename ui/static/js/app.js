/*jshint esversion: 6 */
(function() {
'use strict';

class DomHelper {
    static isVisible(element) {
        return element.offsetParent !== null;
    }

    static openNewTab(url) {
        let win = window.open("");
        win.opener = null;
        win.location = url;
        win.focus();
    }

    static scrollPageTo(element) {
        let windowScrollPosition = window.pageYOffset;
        let windowHeight = document.documentElement.clientHeight;
        let viewportPosition = windowScrollPosition + windowHeight;
        let itemBottomPosition = element.offsetTop + element.offsetHeight;

        if (viewportPosition - itemBottomPosition < 0 || viewportPosition - element.offsetTop > windowHeight) {
            window.scrollTo(0, element.offsetTop - 10);
        }
    }

    static getVisibleElements(selector) {
        let elements = document.querySelectorAll(selector);
        let result = [];

        for (let i = 0; i < elements.length; i++) {
            if (this.isVisible(elements[i])) {
                result.push(elements[i]);
            }
        }

        return result;
    }

    static findParent(element, selector) {
        for (; element && element !== document; element = element.parentNode) {
            if (element.classList.contains(selector)) {
                return element;
            }
        }

        return null;
    }
}

class TouchHandler {
    constructor() {
        this.reset();
    }

    reset() {
        this.touch = {
            start: {x: -1, y: -1},
            move: {x: -1, y: -1},
            element: null
        };
    }

    calculateDistance() {
        if (this.touch.start.x >= -1 && this.touch.move.x >= -1) {
            let horizontalDistance = Math.abs(this.touch.move.x - this.touch.start.x);
            let verticalDistance = Math.abs(this.touch.move.y - this.touch.start.y);

            if (horizontalDistance > 30 && verticalDistance < 70) {
                return this.touch.move.x - this.touch.start.x;
            }
        }

        return 0;
    }

    findElement(element) {
        if (element.classList.contains("touch-item")) {
            return element;
        }

        return DomHelper.findParent(element, "touch-item");
    }

    onTouchStart(event) {
        if (event.touches === undefined || event.touches.length !== 1) {
            return;
        }

        this.reset();
        this.touch.start.x = event.touches[0].clientX;
        this.touch.start.y = event.touches[0].clientY;
        this.touch.element = this.findElement(event.touches[0].target);
    }

    onTouchMove(event) {
        if (event.touches === undefined || event.touches.length !== 1 || this.element === null) {
            return;
        }

        this.touch.move.x = event.touches[0].clientX;
        this.touch.move.y = event.touches[0].clientY;

        let distance = this.calculateDistance();
        let absDistance = Math.abs(distance);

        if (absDistance > 0) {
            let opacity = 1 - (absDistance > 75 ? 0.9 : absDistance / 75 * 0.9);
            let tx = distance > 75 ? 75 : (distance < -75 ? -75 : distance);

            this.touch.element.style.opacity = opacity;
            this.touch.element.style.transform = "translateX(" + tx + "px)";
        }
    }

    onTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        if (this.touch.element !== null) {
            let distance = Math.abs(this.calculateDistance());

            if (distance > 75) {
                EntryHandler.toggleEntryStatus(this.touch.element);
            }
            this.touch.element.style.opacity = 1;
            this.touch.element.style.transform = "none";
        }

        this.reset();
    }

    listen() {
        let elements = document.querySelectorAll(".touch-item");

        elements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onTouchStart(e), false);
            element.addEventListener("touchmove", (e) => this.onTouchMove(e), false);
            element.addEventListener("touchend", (e) => this.onTouchEnd(e), false);
            element.addEventListener("touchcancel", () => this.reset(), false);
        });
    }
}

class KeyboardHandler {
    constructor() {
        this.queue = [];
        this.shortcuts = {};
    }

    on(combination, callback) {
        this.shortcuts[combination] = callback;
    }

    listen() {
        document.onkeydown = (event) => {
            if (this.isEventIgnored(event)) {
                return;
            }

            let key = this.getKey(event);
            this.queue.push(key);

            for (let combination in this.shortcuts) {
                let keys = combination.split(" ");

                if (keys.every((value, index) => value === this.queue[index])) {
                    this.queue = [];
                    this.shortcuts[combination]();
                    return;
                }

                if (keys.length === 1 && key === keys[0]) {
                    this.queue = [];
                    this.shortcuts[combination]();
                    return;
                }
            }

            if (this.queue.length >= 2) {
                this.queue = [];
            }
        };
    }

    isEventIgnored(event) {
        return event.target.tagName === "INPUT" || event.target.tagName === "TEXTAREA";
    }

    getKey(event) {
        const mapping = {
            'Esc': 'Escape',
            'Up': 'ArrowUp',
            'Down': 'ArrowDown',
            'Left': 'ArrowLeft',
            'Right': 'ArrowRight'
        };

        for (let key in mapping) {
            if (mapping.hasOwnProperty(key) && key === event.key) {
                return mapping[key];
            }
        }

        return event.key;
    }
}

class FormHandler {
    static handleSubmitButtons() {
        let elements = document.querySelectorAll("form");
        elements.forEach((element) => {
            element.onsubmit = () => {
                let button = document.querySelector("button");

                if (button) {
                    button.innerHTML = button.dataset.labelLoading;
                    button.disabled = true;
                }
            };
        });
    }
}

class MouseHandler {
    onClick(selector, callback) {
        let elements = document.querySelectorAll(selector);
        elements.forEach((element) => {
            element.onclick = (event) => {
                event.preventDefault();
                callback(event);
            };
        });
    }
}

class RequestBuilder {
    constructor(url) {
        this.callback = null;
        this.url = url;
        this.options = {
            method: "POST",
            cache: "no-cache",
            credentials: "include",
            body: null,
            headers: new Headers({
                "Content-Type": "application/json",
                "X-Csrf-Token": this.getCsrfToken()
            })
        };
    }

    withBody(body) {
        this.options.body = JSON.stringify(body);
        return this;
    }

    withCallback(callback) {
        this.callback = callback;
        return this;
    }

    getCsrfToken() {
        let element = document.querySelector("meta[name=X-CSRF-Token]");
        if (element !== null) {
            return element.getAttribute("value");
        }

        return "";
    }

    execute() {
        fetch(new Request(this.url, this.options)).then((response) => {
            if (this.callback) {
                this.callback(response);
            }
        });
    }
}

class UnreadCounterHandler {
    static decrement(n) {
        this.updateValue((current) => {
            return current - n;
        });
    }

    static increment(n) {
        this.updateValue((current) => {
            return current + n;
        });
    }

    static updateValue(callback) {
        let counterElements = document.querySelectorAll("span.unread-counter");
        counterElements.forEach((element) => {
            let oldValue = parseInt(element.textContent, 10);
            element.innerHTML = callback(oldValue);
        });
        // The titlebar must be updated only on the "Unread" page.
        if (window.location.href.endsWith('/unread')) {
            // The following 3 lines ensure that the unread count in the titlebar
            // is updated correctly when users presses "v".
            let oldValue = parseInt(document.title.split('(')[1], 10);
            let newValue = callback(oldValue);
            // Notes:
            // - This will only be executed in the /unread page. Therefore, it
            //   will not affect titles on other pages.
            // - When there are no unread items, user cannot press "v".
            //   Therefore, we need not handle the case where title is
            //   "Unread Items - Miniflux". This applies to other cases as well.
            //   i.e.: if there are no unread items, user cannot decrement or
            //   increment anything.
            document.title = document.title.replace(
                /(.*?)\(\d+\)(.*?)/,
                function (match, prefix, suffix, offset, string) {
                    return prefix + '(' + newValue + ')' + suffix;
                }
            );
        }
    }
}

class EntryHandler {
    static updateEntriesStatus(entryIDs, status, callback) {
        let url = document.body.dataset.entriesStatusUrl;
        let request = new RequestBuilder(url);
        request.withBody({entry_ids: entryIDs, status: status});
        request.withCallback(callback);
        request.execute();
        // The following 5 lines ensure that the unread count in the menu is
        // updated correctly when users presses "v".
        if (status === "read") {
            UnreadCounterHandler.decrement(1);
        } else {
            UnreadCounterHandler.increment(1);
        }
    }

    static toggleEntryStatus(element) {
        let entryID = parseInt(element.dataset.id, 10);
        let statuses = {read: "unread", unread: "read"};

        for (let currentStatus in statuses) {
            let newStatus = statuses[currentStatus];

            if (element.classList.contains("item-status-" + currentStatus)) {
                element.classList.remove("item-status-" + currentStatus);
                element.classList.add("item-status-" + newStatus);

                this.updateEntriesStatus([entryID], newStatus);

                let link = element.querySelector("a[data-toggle-status]");
                if (link) {
                    this.toggleLinkStatus(link);
                }

                break;
            }
        }
    }

    static toggleLinkStatus(link) {
        if (link.dataset.value === "read") {
            link.innerHTML = link.dataset.labelRead;
            link.dataset.value = "unread";
        } else {
            link.innerHTML = link.dataset.labelUnread;
            link.dataset.value = "read";
        }
    }

    static toggleBookmark(element) {
        element.innerHTML = element.dataset.labelLoading;

        let request = new RequestBuilder(element.dataset.bookmarkUrl);
        request.withCallback(() => {
            if (element.dataset.value === "star") {
                element.innerHTML = element.dataset.labelStar;
                element.dataset.value = "unstar";
            } else {
                element.innerHTML = element.dataset.labelUnstar;
                element.dataset.value = "star";
            }
        });
        request.execute();
    }

    static markEntryAsRead(element) {
        if (element.classList.contains("item-status-unread")) {
            element.classList.remove("item-status-unread");
            element.classList.add("item-status-read");

            let entryID = parseInt(element.dataset.id, 10);
            this.updateEntriesStatus([entryID], "read");
        }
    }

    static saveEntry(element) {
        if (element.dataset.completed) {
            return;
        }

        element.innerHTML = element.dataset.labelLoading;

        let request = new RequestBuilder(element.dataset.saveUrl);
        request.withCallback(() => {
            element.innerHTML = element.dataset.labelDone;
            element.dataset.completed = true;
        });
        request.execute();
    }

    static fetchOriginalContent(element) {
        if (element.dataset.completed) {
            return;
        }

        element.innerHTML = element.dataset.labelLoading;

        let request = new RequestBuilder(element.dataset.fetchContentUrl);
        request.withCallback((response) => {
            element.innerHTML = element.dataset.labelDone;
            element.dataset.completed = true;

            response.json().then((data) => {
                if (data.hasOwnProperty("content")) {
                    document.querySelector(".entry-content").innerHTML = data.content;
                }
            });
        });
        request.execute();
    }
}

class ConfirmHandler {
    remove(url) {
        let request = new RequestBuilder(url);
        request.withCallback(() => window.location.reload());
        request.execute();
    }

    handle(event) {
        let questionElement = document.createElement("span");
        let linkElement = event.target;
        let containerElement = linkElement.parentNode;
        linkElement.style.display = "none";

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

            this.remove(linkElement.dataset.url);
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
}

class MenuHandler {
    clickMenuListItem(event) {
        let element = event.target;

        if (element.tagName === "A") {
            window.location.href = element.getAttribute("href");
        } else {
            window.location.href = element.querySelector("a").getAttribute("href");
        }
    }

    toggleMainMenu() {
        let menu = document.querySelector(".header nav ul");
        if (DomHelper.isVisible(menu)) {
            menu.style.display = "none";
        } else {
            menu.style.display = "block";
        }
    }
}

class ModalHandler {
    static exists() {
        return document.getElementById("modal-container") !== null;
    }

    static open(fragment) {
        if (ModalHandler.exists()) {
            return;
        }

        let container = document.createElement("div");
        container.id = "modal-container";
        container.appendChild(document.importNode(fragment, true));
        document.body.appendChild(container);

        let closeButton = document.querySelector("a.btn-close-modal");
        if (closeButton !== null) {
            closeButton.onclick = (event) => {
                event.preventDefault();
                ModalHandler.close();
            };
        }
    }

    static close() {
        let container = document.getElementById("modal-container");
        if (container !== null) {
            container.parentNode.removeChild(container);
        }
    }
}

class NavHandler {
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
            return;
        }

        for (let i = 0; i < items.length; i++) {
            if (items[i].classList.contains("current-item")) {
                items[i].classList.remove("current-item");

                if (i - 1 >= 0) {
                    items[i - 1].classList.add("current-item");
                    DomHelper.scrollPageTo(items[i - 1]);
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
            return;
        }

        for (let i = 0; i < items.length; i++) {
            if (items[i].classList.contains("current-item")) {
                items[i].classList.remove("current-item");

                if (i + 1 < items.length) {
                    items[i + 1].classList.add("current-item");
                    DomHelper.scrollPageTo(items[i + 1]);
                }

                break;
            }
        }
    }

    isListView() {
        return document.querySelector(".items") !== null;
    }
}

document.addEventListener("DOMContentLoaded", function() {
    FormHandler.handleSubmitButtons();

    let touchHandler = new TouchHandler();
    touchHandler.listen();

    let navHandler = new NavHandler();
    let keyboardHandler = new KeyboardHandler();
    keyboardHandler.on("g u", () => navHandler.goToPage("unread"));
    keyboardHandler.on("g b", () => navHandler.goToPage("starred"));
    keyboardHandler.on("g h", () => navHandler.goToPage("history"));
    keyboardHandler.on("g f", () => navHandler.goToPage("feeds"));
    keyboardHandler.on("g c", () => navHandler.goToPage("categories"));
    keyboardHandler.on("g s", () => navHandler.goToPage("settings"));
    keyboardHandler.on("ArrowLeft", () => navHandler.goToPrevious());
    keyboardHandler.on("ArrowRight", () => navHandler.goToNext());
    keyboardHandler.on("j", () => navHandler.goToPrevious());
    keyboardHandler.on("p", () => navHandler.goToPrevious());
    keyboardHandler.on("k", () => navHandler.goToNext());
    keyboardHandler.on("n", () => navHandler.goToNext());
    keyboardHandler.on("h", () => navHandler.goToPage("previous"));
    keyboardHandler.on("l", () => navHandler.goToPage("next"));
    keyboardHandler.on("o", () => navHandler.openSelectedItem());
    keyboardHandler.on("v", () => navHandler.openOriginalLink());
    keyboardHandler.on("m", () => navHandler.toggleEntryStatus());
    keyboardHandler.on("A", () => navHandler.markPageAsRead());
    keyboardHandler.on("s", () => navHandler.saveEntry());
    keyboardHandler.on("d", () => navHandler.fetchOriginalContent());
    keyboardHandler.on("f", () => navHandler.toggleBookmark());
    keyboardHandler.on("?", () => navHandler.showKeyboardShortcuts());
    keyboardHandler.on("Escape", () => ModalHandler.close());
    keyboardHandler.listen();

    let mouseHandler = new MouseHandler();
    mouseHandler.onClick("a[data-save-entry]", (event) => {
        event.preventDefault();
        EntryHandler.saveEntry(event.target);
    });

    mouseHandler.onClick("a[data-toggle-bookmark]", (event) => {
        event.preventDefault();
        EntryHandler.toggleBookmark(event.target);
    });

    mouseHandler.onClick("a[data-toggle-status]", (event) => {
        event.preventDefault();

        let currentItem = DomHelper.findParent(event.target, "item");
        if (currentItem) {
            EntryHandler.toggleEntryStatus(currentItem);
        }
    });

    mouseHandler.onClick("a[data-fetch-content-entry]", (event) => {
        event.preventDefault();
        EntryHandler.fetchOriginalContent(event.target);
    });

    mouseHandler.onClick("a[data-on-click=markPageAsRead]", () => navHandler.markPageAsRead());
    mouseHandler.onClick("a[data-confirm]", (event) => {
        (new ConfirmHandler()).handle(event);
    });

    if (document.documentElement.clientWidth < 600) {
        let menuHandler = new MenuHandler();
        mouseHandler.onClick(".logo", () => menuHandler.toggleMainMenu());
        mouseHandler.onClick(".header nav li", (event) => menuHandler.clickMenuListItem(event));
    }
});

})();
