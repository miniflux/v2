class MenuHandler {
    clickMenuListItem(event) {
        let element = event.target;

        if (element.tagName === "A") {
            window.location.href = element.getAttribute("href");
        } else {
            window.location.href = element.querySelector("a").getAttribute("href");
        }
    }

    toggleMainMenu() {
        let menu = document.querySelector(".header nav ul");
        if (DomHelper.isVisible(menu)) {
            menu.style.display = "none";
        } else {
            menu.style.display = "block";
        }

        let searchElement = document.querySelector(".header .search");
        if (DomHelper.isVisible(searchElement)) {
            searchElement.style.display = "none";
        } else {
            searchElement.style.display = "block";
        }
    }
}
