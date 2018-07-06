class UnreadCounterHandler {
    static decrement(n) {
        this.updateValue((current) => {
            return current - n;
        });
    }

    static increment(n) {
        this.updateValue((current) => {
            return current + n;
        });
    }

    static updateValue(callback) {
        let counterElements = document.querySelectorAll("span.unread-counter");
        counterElements.forEach((element) => {
            let oldValue = parseInt(element.textContent, 10);
            element.innerHTML = callback(oldValue);
        });

        if (window.location.href.endsWith('/unread')) {
            let oldValue = parseInt(document.title.split('(')[1], 10);
            let newValue = callback(oldValue);

            document.title = document.title.replace(
                /(.*?)\(\d+\)(.*?)/,
                function (match, prefix, suffix, offset, string) {
                    return prefix + '(' + newValue + ')' + suffix;
                }
            );
        }
    }
}
