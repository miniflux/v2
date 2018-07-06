class EntryHandler {
    static updateEntriesStatus(entryIDs, status, callback) {
        let url = document.body.dataset.entriesStatusUrl;
        let request = new RequestBuilder(url);
        request.withBody({entry_ids: entryIDs, status: status});
        request.withCallback(callback);
        request.execute();

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
