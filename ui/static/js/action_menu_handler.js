class ActionMenuHandler {
    constructor(entry) {
        this.entry = DomHelper.findParent(entry, "item");
        this.nav = new NavHandler();

        document.querySelectorAll(".current-item")
            .forEach(e => e.classList.remove("current-item"));
        this.entry.classList.add("current-item");
    }
    show() {
        let template = document.getElementById("action-menus");
        if (template === null) return;

        ModalHandler.open(template.content, true);
        let list = document.querySelector(".action-menus #item-meta-links");
        while (list.hasChildNodes()) {
            list.removeChild(list.firstChild);
        }

        this.entry.querySelectorAll(".item-meta a").forEach(
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

        document.querySelector("#menu-mark-above-read").addEventListener("click", () => {
            EntryHandler.setEntriesAboveStatusRead(this.entry);
            ModalHandler.close();
        });

        document.querySelector("#menu-action-cancel").addEventListener("click", () => {
            ModalHandler.close();
        });

        function clickElement(element) {
            let e = document.createEvent("MouseEvents");
            e.initEvent("click", true, true);
            element.dispatchEvent(e);
        }
    }
}