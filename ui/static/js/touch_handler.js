class TouchHandler {
    constructor() {
        this.reset();
    }

    reset() {
        this.touch = {
            start: { x: -1, y: -1 },
            move: { x: -1, y: -1 },
            moved: false,
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
        if (element.classList.contains("touch-item")) {
            return element;
        }

        return DomHelper.findParent(element, "touch-item");
    }

    onTouchStart(event) {
        if (event.touches === undefined || event.touches.length !== 1) {
            return;
        }

        this.reset();
        this.touch.start.x = event.touches[0].clientX;
        this.touch.start.y = event.touches[0].clientY;
        this.touch.element = this.findElement(event.touches[0].target);
        this.touch.element.style.transitionDuration = "0s";
    }

    onTouchMove(event) {
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

    onTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        if (this.touch.element !== null) {
            let distance = Math.abs(this.calculateDistance());

            if (distance > 75) {
                toggleEntryStatus(this.touch.element);
            }

            if (this.touch.moved) {
                this.touch.element.style.transitionDuration = "0.15s";
                this.touch.element.style.transform = "none";
            }
        }

        this.reset();
    }

    listen() {
        let elements = document.querySelectorAll(".touch-item");
        let hasPassiveOption = DomHelper.hasPassiveEventListenerOption();

        elements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onTouchStart(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchmove", (e) => this.onTouchMove(e), hasPassiveOption ? { passive: false } : false);
            element.addEventListener("touchend", (e) => this.onTouchEnd(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
        });

        let entryContentElement = document.querySelector(".entry-content");
        if (entryContentElement) {
            let doubleTapTimers = {
                previous: null,
                next: null
            };

            const detectDoubleTap = (doubleTapTimer, event) => {
                const timer = doubleTapTimers[doubleTapTimer];
                if (timer === null) {
                    doubleTapTimers[doubleTapTimer] = setTimeout(() => {
                        doubleTapTimers[doubleTapTimer] = null;
                    }, 200);
                } else {
                    event.preventDefault();
                    goToPage(doubleTapTimer);
                }
            };

            entryContentElement.addEventListener("touchend", (e) => {
                if (e.changedTouches[0].clientX >= (entryContentElement.offsetWidth / 2)) {
                    detectDoubleTap("next", e);
                } else {
                    detectDoubleTap("previous", e);
                }
            }, hasPassiveOption ? { passive: false } : false);

            entryContentElement.addEventListener("touchmove", (e) => {
                Object.keys(doubleTapTimers).forEach(timer => doubleTapTimers[timer] = null);
            });
        }
    }
}
