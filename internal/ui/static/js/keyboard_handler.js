class KeyboardHandler {
    constructor() {
        this.queue = [];
        this.shortcuts = {};
        this.triggers = new Set();
    }

    on(combination, callback) {
        this.shortcuts[combination] = callback;
        this.triggers.add(combination.split(" ")[0]);
    }

    listen() {
        document.onkeydown = (event) => {
            let key = this.getKey(event);
            if (this.isEventIgnored(event, key) || this.isModifierKeyDown(event)) {
                return;
            }

            if (key != "Enter") {
                event.preventDefault();
            }

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

    isEventIgnored(event, key) {
        return event.target.tagName === "INPUT" ||
            event.target.tagName === "TEXTAREA" ||
            (this.queue.length < 1 && !this.triggers.has(key));
    }

    isModifierKeyDown(event) {
        return event.getModifierState("Control") || event.getModifierState("Alt") || event.getModifierState("Meta");
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
