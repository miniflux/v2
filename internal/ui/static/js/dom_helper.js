class DomHelper {
    static isVisible(element) {
        return element.offsetParent !== null;
    }

    static openNewTab(url) {
        let win = window.open("");
        win.opener = null;
        win.location = url;
        win.focus();
    }

    static scrollPageTo(element, evenIfOnScreen) {
        let windowScrollPosition = window.pageYOffset;
        let windowHeight = document.documentElement.clientHeight;
        let viewportPosition = windowScrollPosition + windowHeight;
        let itemBottomPosition = element.offsetTop + element.offsetHeight;

        if (evenIfOnScreen || viewportPosition - itemBottomPosition < 0 || viewportPosition - element.offsetTop > windowHeight) {
            window.scrollTo(0, element.offsetTop - 10);
        }
    }

    static getVisibleElements(selector) {
        let elements = document.querySelectorAll(selector);
        let result = [];

        for (let i = 0; i < elements.length; i++) {
            if (this.isVisible(elements[i])) {
                result.push(elements[i]);
            }
        }

        return result;
    }

    static findParent(element, selector) {
        for (; element && element !== document; element = element.parentNode) {
            if (element.classList.contains(selector)) {
                return element;
            }
        }

        return null;
    }

    static hasPassiveEventListenerOption() {
        var passiveSupported = false;

        try {
            var options = Object.defineProperty({}, "passive", {
                get: function() {
                    passiveSupported = true;
                }
            });

            window.addEventListener("test", options, options);
            window.removeEventListener("test", options, options);
        } catch(err) {
            passiveSupported = false;
        }

        return passiveSupported;
    }
}
