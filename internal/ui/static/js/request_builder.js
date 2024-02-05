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
                "X-Csrf-Token": getCsrfToken()
            })
        };
    }

    withHttpMethod(method) {
        this.options.method = method;
        return this;
    }

    withBody(body) {
        this.options.body = JSON.stringify(body);
        return this;
    }

    withCallback(callback) {
        this.callback = callback;
        return this;
    }

    execute() {
        fetch(new Request(this.url, this.options)).then((response) => {
            if (this.callback) {
                this.callback(response);
            }
        });
    }
}
