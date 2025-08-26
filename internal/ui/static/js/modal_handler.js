class KeyboardModalHandler {
    static setupFocusTrap() {
        const container = document.getElementById("modal-container");
        if (container !== null) {
            container.onkeydown = (e) => {
                if (e.key === 'Tab') {
                    // Since there is only one focusable button in the keyboard modal we always want to focus it with the tab key. This handles
                    // the special case of having just one focusable element in a dialog/ where keyboard focus is placed on an element that is not in the
                    // tab order.
                    container.querySelectorAll('button')[0].focus();
                    e.preventDefault();
                }
            };
        }
    }

    static open(fragment, initialFocusElementId) {
        if (document.getElementById("modal-container") !== null){
            return;
        }

        this.activeElement = document.activeElement;

        const container = document.createElement("div");
        container.id = "modal-container";
        container.setAttribute("role", "dialog");
        container.appendChild(document.importNode(fragment, true));
        document.body.appendChild(container);

        const closeButton = document.querySelector("button.btn-close-modal");
        if (closeButton !== null) {
            closeButton.onclick = (event) => {
                event.preventDefault();
                KeyboardModalHandler.close();
            };
        }

        const initialFocusElement = document.getElementById(initialFocusElementId);
        if (initialFocusElement !== undefined) {
            initialFocusElement.focus();
        }

        this.setupFocusTrap();
    }

    static close() {
        const container = document.getElementById("modal-container");
        if (container !== null) {
            container.parentNode.removeChild(container);
            if (this.activeElement !== undefined && this.activeElement !== null) {
                this.activeElement.focus();
            }
        }
    }
}
