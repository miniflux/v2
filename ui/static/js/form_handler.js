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
