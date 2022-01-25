class TouchHandler {
    constructor() {
        this.reset();
    }

    reset() {
        this.touch = {
            start: { x: -1, y: -1 },
            move: { x: -1, y: -1 },
            element: null
        };
    }

    calculateDistance() {
        if (this.touch.start.x >= -1 && this.touch.move.x >= -1) {
            let horizontalDistance = Math.abs(this.touch.move.x - this.touch.start.x);
            let verticalDistance = Math.abs(this.touch.move.y - this.touch.start.y);

            if (horizontalDistance > 30 && verticalDistance < 70) {
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
            let opacity = 1 - (absDistance > 75 ? 0.9 : absDistance / 75 * 0.9);
            let tx = distance > 75 ? 75 : (distance < -75 ? -75 : distance);

            this.touch.element.style.opacity = opacity;
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
	    } else {
                this.touch.element.style.opacity = 1;
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
