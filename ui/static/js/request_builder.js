class RequestBuilder {
    constructor(url) {
        this.callback = null;
        this.url = url;
        this.options = {
            method: "POST",
            cache: "no-cache",
            credentials: "include",
            body: null,
            headers: new Headers({
                "Content-Type": "application/json",
                "X-Csrf-Token": this.getCsrfToken()
            })
        };
    }

    withBody(body) {
        this.options.body = JSON.stringify(body);
        return this;
    }

    withCallback(callback) {
        this.callback = callback;
        return this;
    }

    getCsrfToken() {
        let element = document.querySelector("meta[name=X-CSRF-Token]");
        if (element !== null) {
            return element.getAttribute("value");
        }

        return "";
    }

    execute() {
        fetch(new Request(this.url, this.options)).then((response) => {
            if (this.callback) {
                this.callback(response);
            }
        });
    }
}
