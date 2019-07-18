class ConfirmHandler {
    executeRequest(url, redirectURL) {
        let request = new RequestBuilder(url);

        request.withCallback(() => {
            if (redirectURL) {
                window.location.href = redirectURL;
            } else {
                window.location.reload();
            }
        });

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

            if (linkElement.dataset.markPageAsRead) {
                (new NavHandler()).markPageAsRead(event.target.dataset.showOnlyUnread || false);
            } else {
                this.executeRequest(linkElement.dataset.url, linkElement.dataset.redirectUrl);
            }
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
