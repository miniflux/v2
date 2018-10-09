class MouseHandler {
    onClick(selector, callback, noPreventDefault) {
        let elements = document.querySelectorAll(selector);
        elements.forEach((element) => {
            element.onclick = (event) => {
                if (! noPreventDefault) {
                    event.preventDefault();
                }

                callback(event);
            };
        });
    }
}