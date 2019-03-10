class TouchHandler {
    constructor() {
        this.reset();
   this.toaster=document.getElementById("toaster");
    }

    reset() {
        this.touch = {
            start: {x: -1, y: -1},
            move: {x: -1, y: -1},
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
        if (element.classList.contains("touch-article")) {
            return element;
        }
        if (element.classList.contains("touch-item")) {
            return element;
        }
       let parentElement = DomHelper.findParent(element, "touch-article");
       if( parentElement == null) 
       {
        return DomHelper.findParent(element, "touch-item");
       }
       else
       { 
          return parentElement;
       }
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

    onItemTouchMove(event) {
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
    onItemTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        if (this.touch.element !== null) {
            let distance = Math.abs(this.calculateDistance());

            if (distance > 75) {
                EntryHandler.toggleEntryStatus(this.touch.element);
            }
            this.touch.element.style.opacity = 1;
            this.touch.element.style.transform = "none";
        }

        this.reset();
    }

    onArticleTouchMove(event) {
        if (event.touches === undefined || event.touches.length !== 1 || this.element === null) {
            return;
        }

        this.touch.move.x = event.touches[0].clientX;
        this.touch.move.y = event.touches[0].clientY;

        let distance = this.calculateDistance();
        if (distance > 0) {
            this.toaster.innerHTML = "Previous";
   }else{
            this.toaster.innerHTML = "Next";
   }
        let absDistance = Math.abs(distance);

        if (absDistance > 0) {
//            let opacity = 1 - (absDistance > 125 ? 0.9 : absDistance / 125 * 0.9);
//            this.touch.element.style.opacity = opacity;
       let tx = 40 - (absDistance > 90 ? 40 : absDistance / 90 * 40);
       this.toaster.style.transform = "translate(-50%," + tx + "px)";
            event.preventDefault();
        }
        if (absDistance > 100) {
            this.toaster.style.opacity = 1.0;
        } else {
            this.toaster.style.opacity = 0.3;
        }
    }
    onArticleTouchEnd(event) {
        if (event.touches === undefined) {
            return;
        }

        if (this.touch.element !== null) {
            let distance = this.calculateDistance();
            if (Math.abs(distance) > 150) {
          if (distance < 100) {
             let element = document.querySelector("a[data-page=next]");
             if (element) {
                document.location.href = element.href;
             }
          }
          if (distance > 100) {
             let element = document.querySelector("a[data-page=previous]");

             if (element) {
                document.location.href = element.href;
             } 
          }
       }
//            this.touch.element.style.opacity = 1;
//            this.touch.element.style.transform = "none";
            this.toaster.style.opacity = 0.3;
       this.toaster.style.transform = "translate(-50%,40px)";
            this.toaster.innerHTML = "";
        }
        this.reset();
    }

    listen() {
        let itemElements = document.querySelectorAll(".touch-item");
        let articleElements = document.querySelectorAll(".touch-article");
        let hasPassiveOption = DomHelper.hasPassiveEventListenerOption();

        itemElements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onTouchStart(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchmove", (e) => this.onItemTouchMove(e), hasPassiveOption ? { passive: false } : false);
            element.addEventListener("touchend", (e) => this.onItemTouchEnd(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
        });
        articleElements.forEach((element) => {
            element.addEventListener("touchstart", (e) => this.onTouchStart(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchmove", (e) => this.onArticleTouchMove(e), hasPassiveOption ? { passive: false } : false);
            element.addEventListener("touchend", (e) => this.onArticleTouchEnd(e), hasPassiveOption ? { passive: true } : false);
            element.addEventListener("touchcancel", () => this.reset(), hasPassiveOption ? { passive: true } : false);
        });
    }
}
