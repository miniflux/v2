class ModalHandler {
    static exists() {
        return document.getElementById("modal-container") !== null;
    }

    static open(fragment, preventScroll) {
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

        if (preventScroll) ModalHelper.afterOpen()
    }

    static close() {
        let container = document.getElementById("modal-container");
        if (container !== null) {
            if (document.body.classList.contains("modal-open"))
                ModalHelper.beforeClose();
            container.parentNode.removeChild(container);
        }
    }
}

var ModalHelper = (function (bodyCls) {
    var scrollTop;
    return {
        afterOpen: function () {
            scrollTop = document.scrollingElement.scrollTop;
            document.body.classList.add(bodyCls);
            document.body.style.top = -scrollTop + 'px';
        },
        beforeClose: function () {
            document.body.classList.remove(bodyCls);
            // scrollTop lost after set position:fixed, restore it back.
            document.scrollingElement.scrollTop = scrollTop;
        }
    };
})('modal-open');