class TouchHandler {
    constructor() {
        this.reset();
    }

    reset() {
        this.touch = {
            start: { x: -1, y: -1 },
            move: { x: -1, y: -1 },
            moved: false,
            time: 0,
            element: null
        };
    }

    calculateDistance() {
        if (this.touch.start.x >= -1 && this.touch.move.x >= -1) {
            let horizontalDistance = Math.abs(this.touch.move.x - this.touch.start.x);
            let verticalDistance = Math.abs(this.touch.move.y - this.touch.start.y);

            if (horizontalDistance > 30 && verticalDistance < 70 || this.touch.moved) {
                return this.touch.move.x - this.touch.start.x;
            }
        }

        return 0;
    }

    findElement(element) {
        if (element.classList.contains("entry-swipe")) {
            return element;
        }

        return DomHelper.findParent(element, "entry-swipe");
    }

    onItemTouchStart(event) {
        if (event.touches === undefined || event.touches.length !== 1) {
            return;
        }

        this.reset();
        this.touch.start.x = event.touches[0].clientX;
        this.touch.start.y = event.touches[0].clientY;
        this.touch.element = this.findElement(event.touches[0].target);
        this.touch.element.style.transitionDuration = "0s";
    }

    onItemTouchMove(event) {
        if (event.touches === undefined || event.touches.length !== 1 || this.element === null) {
            return;
        }

        this.touch.move.x = event.touches[0].clientX;
        this.touch.move.y = event.touches[0].clientY;

        let distance = this.calculateDistance();
        let absDistance = Math.abs(distance);

        if (absDistance > 0) {
            this.touch.moved = true;

            let tx = absDistance > 75 ? Math.pow(absDistance - 75, 0.5) + 75 : absDistance;

            if (distance < 0) {
                tx = -tx;
            }

            this.touch.element.style.transform = "translateX(" + tx + "px)";

            event.preventDefault();
        }
    }

    onItemTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        if (this.touch.element !== null) {
            let absDistance = Math.abs(this.calculateDistance());

            if (absDistance > 75) {
                toggleEntryStatus(this.touch.element);
            }

            if (this.touch.moved) {
                this.touch.element.style.transitionDuration = "0.15s";
                this.touch.element.style.transform = "none";
            }
        }

        this.reset();
    }

    onContentTouchStart(event) {
        if (event.touches === undefined || event.touches.length !== 1) {
            return;
        }

        this.reset();
        this.touch.start.x = event.touches[0].clientX;
        this.touch.start.y = event.touches[0].clientY;
        this.touch.time = Date.now();
    }

    onContentTouchMove(event) {
        if (event.touches === undefined || event.touches.length !== 1 || this.element === null) {
            return;
        }

        this.touch.move.x = event.touches[0].clientX;
        this.touch.move.y = event.touches[0].clientY;
    }

    onContentTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        let distance = this.calculateDistance();
        let absDistance = Math.abs(distance);
        let now = Date.now();

        if (now - this.touch.time <= 1000 && absDistance > 75) {
            if (distance > 0) {
                goToPage("previous");
            } else {
                goToPage("next");
            }
        }

        this.reset();
    }

    onTapEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        let now = Date.now();

        if (this.touch.start.x !== -1 && now - this.touch.time <= 200) {
            let innerWidthHalf = window.innerWidth / 2;

            if (this.touch.start.x >= innerWidthHalf && event.changedTouches[0].clientX >= innerWidthHalf) {
                goToPage("next");
            } else if (this.touch.start.x < innerWidthHalf && event.changedTouches[0].clientX < innerWidthHalf) {
                goToPage("previous");
            }

            this.reset();
        } else {
            this.reset();
            this.touch.start.x = event.changedTouches[0].clientX;
            this.touch.time = now;
        }
    }

    listen() {
        let hasPassiveOption = DomHelper.hasPassiveEventListenerOption();

        let elements = document.querySelectorAll(".entry-swipe");

        elements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onItemTouchStart(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchmove", (e) => this.onItemTouchMove(e), hasPassiveOption ? { passive: false } : false);
            element.addEventListener("touchend", (e) => this.onItemTouchEnd(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
        });

        let element = document.querySelector(".entry-content");

        if (element) {
            if (element.classList.contains("gesture-nav-tap")) {
                element.addEventListener("touchend", (e) => this.onTapEnd(e), hasPassiveOption ? { passive: true } : false);
                element.addEventListener("touchmove", () => this.reset(), hasPassiveOption ? { passive: true } : false);
                element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
            } else if (element.classList.contains("gesture-nav-swipe")) {
                element.addEventListener("touchstart", (e) => this.onContentTouchStart(e), hasPassiveOption ? { passive: true } : false);
                element.addEventListener("touchmove", (e) => this.onContentTouchMove(e), hasPassiveOption ? { passive: true } : false);
                element.addEventListener("touchend", (e) => this.onContentTouchEnd(e), hasPassiveOption ? { passive: true } : false);
                element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
            }
        }
    }
}
