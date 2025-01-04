class DomHelper {
    static isVisible(element) {
        return element.offsetParent !== null;
    }

    static openNewTab(url) {
        const win = window.open("");
        win.opener = null;
        win.location = url;
        win.focus();
    }

    static scrollPageTo(element, evenIfOnScreen) {
        const windowScrollPosition = window.pageYOffset;
        const windowHeight = document.documentElement.clientHeight;
        const viewportPosition = windowScrollPosition + windowHeight;
        const itemBottomPosition = element.offsetTop + element.offsetHeight;

        if (evenIfOnScreen || viewportPosition - itemBottomPosition < 0 || viewportPosition - element.offsetTop > windowHeight) {
            window.scrollTo(0, element.offsetTop - 10);
        }
    }

    static getVisibleElements(selector) {
        const elements = document.querySelectorAll(selector);
        return [...elements].filter((element) => this.isVisible(element));
    }
}
