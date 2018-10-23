class KeyboardHandler {
    constructor() {
        this.queue = [];
        this.shortcuts = {};
    }

    on(combination, callback) {
        this.shortcuts[combination] = callback;
    }

    listen() {
        document.onkeydown = (event) => {
            if (this.isEventIgnored(event)) {
                return;
            }

            let key = this.getKey(event);
            this.queue.push(key);

            for (let combination in this.shortcuts) {
                let keys = combination.split(" ");

                if (keys.every((value, index) => value === this.queue[index])) {
                    this.queue = [];
                    this.shortcuts[combination](event);
                    return;
                }

                if (keys.length === 1 && key === keys[0]) {
                    this.queue = [];
                    this.shortcuts[combination](event);
                    return;
                }
            }

            if (this.queue.length >= 2) {
                this.queue = [];
            }
        };
    }

    isEventIgnored(event) {
        return event.target.tagName === "INPUT" || event.target.tagName === "TEXTAREA";
    }

    getKey(event) {
        const mapping = {
            'Esc': 'Escape',
            'Up': 'ArrowUp',
            'Down': 'ArrowDown',
            'Left': 'ArrowLeft',
            'Right': 'ArrowRight'
        };

        for (let key in mapping) {
            if (mapping.hasOwnProperty(key) && key === event.key) {
                return mapping[key];
            }
        }

        return event.key;
    }
}
