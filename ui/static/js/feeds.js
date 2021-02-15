function createSortElement(containerElement, baseUrl, sortText, targetSortedBy, currentSortedBy, currentDirection, isFirstElement) {
    let checkElement = document.createElement("a");
    checkElement.innerHTML = "";
    let targetSortDirection = currentDirection;
    if (targetSortedBy === currentSortedBy) {
        checkElement.innerHTML = "&#x2713;";
        targetSortDirection = (currentDirection==="asc") ? "desc" : "asc";
    }
    
    let element = document.createElement("a");
    element.href = "#";
    element.appendChild(document.createTextNode(sortText));
    element.onclick = (event) => {
        event.preventDefault();

        let loadingElement = document.createElement("a");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(containerElement.dataset.labelLoading));
        containerElement.appendChild(loadingElement);

        let request = new RequestBuilder(baseUrl);
        request.withHttpMethod("GET");
        request.withCallback(() => {
            window.location.href = baseUrl + "?feed_sorted_by=" + targetSortedBy + "&feed_direction=" + targetSortDirection;
        });

        request.execute();
    };

    containerElement.appendChild(document.createTextNode(isFirstElement ? " " : ", "));
    containerElement.appendChild(checkElement);
    containerElement.appendChild(element);
    return;
}
function triggerSortFeeds(linkElement) {
    let containerElement = linkElement.parentNode;
    let id = "sort-feeds-question";
    let existingQuestionElement = document.getElementById(id);
    if (existingQuestionElement) {
        existingQuestionElement.remove();
    } else {
        let currentSortedBy = linkElement.dataset.currentSortedBy;
        let currentDirection = linkElement.dataset.currentSortDirection;
        let baseUrl = linkElement.dataset.sortedBaseUrl;

        let arrow = "&uarr; ";
        if (currentDirection === "asc") {
            arrow = "&darr; ";
        }

        let questionElement = document.createElement("span");
        questionElement.setAttribute("id", id);
        questionElement.dataset.labelLoading = linkElement.dataset.labelLoading;
        questionElement.className = "confirm";

        let arrowElement = document.createElement("a");
        arrowElement.setAttribute("id", "arrow-element");
        arrowElement.innerHTML = arrow;
        questionElement.appendChild(arrowElement);

        for (let i = 0; true; i++) {
            let label = linkElement.getAttribute("data-sorted-text-" + i.toString());
            let targetSortedBy = linkElement.getAttribute("data-sorted-by-" + i.toString());
            if (!label || !targetSortedBy) {
                break;
            }
            createSortElement(questionElement, baseUrl, label, targetSortedBy, currentSortedBy, currentDirection, i===0);
        }
        questionElement.appendChild(document.createTextNode(" "));

        containerElement.appendChild(questionElement);
    }
} 