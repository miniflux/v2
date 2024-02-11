class KeyboardHandler {
    constructor() {
        this.queue = [];
        this.shortcuts = {};
        this.triggers = [];
    }

    on(combination, callback) {
        this.shortcuts[combination] = callback;
        this.triggers.push(combination.split(" ")[0]);
    }

    listen() {
        document.onkeydown = (event) => {
            let key = this.getKey(event);
            if (this.isEventIgnored(event, key) || this.isModifierKeyDown(event) || this.isConfirmAlertModalCloseEvent(key)) {
                return;
            }

            event.preventDefault();
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
            (this.queue.length < 1 && !this.triggers.includes(key));
    }

    isModifierKeyDown(event) {
        return event.getModifierState("Control") || event.getModifierState("Alt") || event.getModifierState("Meta");
    }

    isConfirmAlertModalCloseEvent(key) {
        const dialogElement = document.getElementById("confirm-alert-dialog")
        return dialogElement.getAttribute("open") !== null && key == "Escape"
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
