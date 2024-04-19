class ModalHandler {
    static exists() {
        return document.getElementById("modal-container") !== null;
    }

    static getModalContainer() {
        return document.getElementById("modal-container");
    }

    static getFocusableElements() {
        const container = this.getModalContainer();

        if (container === null) {
            return null;
        }

        return container.querySelectorAll('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
    }

    static setupFocusTrap() {
        const focusableElements = this.getFocusableElements();

        if (focusableElements === null) {
            return;
        }

        const firstFocusableElement = focusableElements[0];
        const lastFocusableElement = focusableElements[focusableElements.length - 1];

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

        const container = document.createElement("div");
        container.id = "modal-container";
        container.setAttribute("role", "dialog");
        container.appendChild(document.importNode(fragment, true));
        document.body.appendChild(container);

        const closeButton = document.querySelector("button.btn-close-modal");
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
            let focusableElements = this.getFocusableElements();
            if (focusableElements !== null) {
                initialFocusElement = focusableElements[0];
            }
        }

        if (initialFocusElement !== undefined) {
            initialFocusElement.focus();
        }

        this.setupFocusTrap();
    }

    static close() {
        const container = this.getModalContainer();
        if (container !== null) {
            container.parentNode.removeChild(container);
        }

        if (this.activeElement !== undefined && this.activeElement !== null) {
            this.activeElement.focus();
        }
    }
}
