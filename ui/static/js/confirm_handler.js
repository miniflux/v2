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
