class LinkStateHandler {
    static flip(element) {
        let labelElement = document.createElement("span")
        labelElement.className = "link-flipped-state";
        labelElement.appendChild(document.createTextNode(element.dataset.labelNewState));

        element.parentNode.appendChild(labelElement);
        element.parentNode.removeChild(element);
    }
}
