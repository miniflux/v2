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