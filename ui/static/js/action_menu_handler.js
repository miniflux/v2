class ActionMenuHandler {
    constructor(element) {
        this.element = element;
        this.nav = new NavHandler();

        document.querySelectorAll(".current-item")
            .forEach(e => e.classList.remove("current-item"));
        this.element.classList.add("current-item");
    }
    show() {
        let template = document.getElementById("action-menus");
        if (template === null) return;

        if (this.element.classList.contains("item")) {
            // menu for entries
            hideMenuExcept("li[data-for='entries']");
            initMenu(this.element.querySelectorAll(".item-meta a"));
            document.querySelector("#menu-mark-above-read").addEventListener("click", () => {
                EntryHandler.setEntriesAboveStatusRead(this.element);
                ModalHandler.close();
            });
        } else if (this.element.classList.contains("entry")) {
            // menu for entry
            hideMenuExcept("li[data-for='entry']");
            initMenu(this.element.querySelectorAll(".entry-actions a"));
        }
        // cancel menu
        document.querySelector("#menu-action-cancel").addEventListener("click", () => {
            ModalHandler.close();
        });

        function hideMenuExcept(selector) {
            document.querySelectorAll(".action-menus li[data-for]").forEach(e => e.style.display = 'none');
            if (selector) document.querySelectorAll(".action-menus " + selector).forEach(e => e.style.display = '');
        }

        function initMenu(links) {
            ModalHandler.open(template.content, true);
            let list = document.querySelector(".action-menus #element-links");
            while (list.hasChildNodes()) {
                list.removeChild(list.firstChild);
            }

            links.forEach(
                link => {
                    let menu = document.createElement("li");
                    menu.innerText = link.innerText;
                    menu.addEventListener("click", () => {
                        clickElement(link);
                        ModalHandler.close();
                    });
                    list.appendChild(menu);
                }
            );
        }

        function clickElement(element) {
            let e = document.createEvent("MouseEvents");
            e.initEvent("click", true, true);
            element.dispatchEvent(e);
        }
    }
}