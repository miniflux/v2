class ActionMenuHandler {
    constructor(element) {
        this.element = element;
        document.querySelectorAll(".current-item")
            .forEach(e => e.classList.remove("current-item"));
        this.element.classList.add("current-item");
    }
    show() {
        let template = document.getElementById("action-menus");
        if (template === null) return;

        if (this.element.classList.contains("item")) {
            // menu for entries
            initMenu(this.element.querySelectorAll(".item-meta a"), "entries");
            document.querySelector("#menu-mark-above-read").addEventListener("click", () => {
                EntryHandler.setEntriesAboveStatusRead(this.element);
                ModalHandler.close();
            });
        } else if (this.element.classList.contains("entry")) {
            // menu for entry
            initMenu(this.element.querySelectorAll(".entry-actions a"), "entry");
        }
        // cancel menu
        document.querySelector("#menu-action-cancel").addEventListener("click", () => {
            ModalHandler.close();
        });

        // initMenu creates menu for given links in action modal.
        // dataForValue specifies the part of predefined menu to keep, 
        // which have the given value for "data-for" attribute.
        function initMenu(links, dataForValue) {
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

            document.querySelectorAll(".action-menus li[data-for]")
                .forEach(e => {
                    if (e.dataset.for !== dataForValue)
                        e.style.display = 'none';
                });
        }

        function clickElement(element) {
            let e = document.createEvent("MouseEvents");
            e.initEvent("click", true, true);
            element.dispatchEvent(e);
        }
    }
}