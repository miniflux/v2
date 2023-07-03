class ModalHandler {
    static exists() {
        return document.getElementById("modal-container") !== null;
    }

    static getModalContainer() {
        let container = document.getElementById("modal-container");

        if (container === undefined) {
            return;
        }

        return container;
    }

    static getFocusableElements() {
        let container = this.getModalContainer();

        if (container === undefined) {
            return;
        }

        return container.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    }

    static setupFocusTrap() {
        let focusableElements = this.getFocusableElements();

        if (focusableElements === undefined) {
            return;
        }

        let firstFocusableElement = focusableElements[0];
        let lastFocusableElement = focusableElements[focusableElements.length - 1];

        this.getModalContainer().onkeydown = (e) => {
            if (e.key !== 'Tab') {
                return;
            }

            // If there is only one focusable element in the dialog we always want to focus that one with the tab key.
            // This handles the special case of having just one focusable element in a dialog where keyboard focus is placed on an element that is not in the tab order.
            if (focusableElements.length === 1) {
                firstFocusableElement.focus();
                e.preventDefault();
                return;
            }

            if (e.shiftKey && document.activeElement === firstFocusableElement) {
                lastFocusableElement.focus();
                e.preventDefault();
            } else if (!e.shiftKey && document.activeElement === lastFocusableElement) {
                firstFocusableElement.focus();
                e.preventDefault();
            }
        };
    }

    static open(fragment, initialFocusElementId) {
        if (ModalHandler.exists()) {
            return;
        }

        this.activeElement = document.activeElement;

        let container = document.createElement("div");
        container.id = "modal-container";
        container.setAttribute("role", "dialog");
        container.appendChild(document.importNode(fragment, true));
        document.body.appendChild(container);

        let closeButton = document.querySelector("button.btn-close-modal");
        if (closeButton !== null) {
            closeButton.onclick = (event) => {
                event.preventDefault();
                ModalHandler.close();
            };
        }

        let initialFocusElement;
        if (initialFocusElementId !== undefined) {
            initialFocusElement = document.getElementById(initialFocusElementId);
        } else {
            initialFocusElement = this.getFocusableElements()[0];
        }

        initialFocusElement.focus();

        this.setupFocusTrap();
    }

    static close() {
        let container = this.getModalContainer();
        if (container !== null) {
            container.parentNode.removeChild(container);
        }

        if (this.activeElement !== undefined) {
            this.activeElement.focus();
        }
    }
}
