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
