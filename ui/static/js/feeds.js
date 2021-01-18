var sortAscending = {
    disabled:1,
    parsing_error_count:-1,
    title:1,
    total_count:-1,
    unread_count:-1,
};
var checkElements = {
    disabled:null,
    parsing_error_count:null,
    title:null,
    total_count:null,
    unread_count:null,
};
var currentChecked;
var currentAscending;
const arrowElementId = "arrowElement";
function compareDisabled(a,b) {
    if (a.dataset.disabled>b.dataset.disabled)return -sortAscending.disabled;
    if (a.dataset.disabled<b.dataset.disabled)return sortAscending.disabled;
    return 0;
}
function compareErrors(a,b) {
    let a_value = parseInt(a.dataset.errors);
    let b_value = parseInt(b.dataset.errors);
    if (a_value<b_value)return -sortAscending.parsing_error_count;
    if (a_value>b_value)return sortAscending.parsing_error_count;
    return 0;
}
function compareTitle(a,b) {
    if (a.dataset.title.toLowerCase()<b.dataset.title.toLowerCase())return -sortAscending.title;
    if (a.dataset.title.toLowerCase()>b.dataset.title.toLowerCase())return sortAscending.title;
    return 0;
}
function compareTotal(a,b) {
    let a_value = parseInt(a.dataset.total);
    let b_value = parseInt(b.dataset.total);
    if (a_value<b_value)return -sortAscending.total_count;
    if (a_value>b_value)return sortAscending.total_count;
    return 0;
}
function compareUnread(a,b) {
    let a_value = parseInt(a.dataset.unread);
    let b_value = parseInt(b.dataset.unread);
    if (a_value<b_value)return -sortAscending.unread_count;
    if (a_value>b_value)return sortAscending.unread_count;
    return 0;
}
function sortFeedsBy(compareFunction) {
    let items = DomHelper.getVisibleElements(".items .item");
    if (items.length === 0) {
        return;
    }
    items.sort(compareFunction);
    let feed_list = document.getElementsByClassName("items")[0];
    feed_list.innerHTML="";
    items.forEach((element) => {
        feed_list.appendChild(element);
    });
    return;
}
function createSortElement(containerElement, sortText, compareFunction, compareKey, isLastElement) {
    let checkElement = document.createElement("a");
    checkElement.innerHTML = "";
    if (compareKey === currentChecked) {
        checkElement.innerHTML = "&#x2713;";
    }
    containerElement.appendChild(checkElement);
    checkElements[compareKey] = checkElement;

    let element = document.createElement("a");
    element.href = "#";
    element.appendChild(document.createTextNode(sortText));
    element.onclick = (event) => {
        event.preventDefault();

        currentAscending = sortAscending[compareKey];
        let arrow = "&uarr; ";
        if (currentAscending === 1) {
            arrow = "&darr; ";
        }
        let arrowElement = document.getElementById(arrowElementId);
        arrowElement.innerHTML = arrow;

        for (var key in checkElements) {
            checkElements[key].innerHTML = "";
        }
        checkElements[compareKey].innerHTML = "&#x2713;";
        currentChecked = compareKey;

        let loadingElement = document.createElement("a");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(containerElement.dataset.labelLoading));
        containerElement.appendChild(loadingElement);

        setTimeout(function(){
            sortFeedsBy(compareFunction);
            sortAscending[compareKey] = -sortAscending[compareKey];
            loadingElement.remove();
        }, 20);
    };   

    containerElement.appendChild(element);
    var separator = ", ";
    if (isLastElement) {
        separator = " ";
    }
    containerElement.appendChild(document.createTextNode(separator));
    return;
}
function triggerSortFeeds(linkElement) {
    let containerElement = linkElement.parentNode;
    let id = "sort-feeds-question";
    let existingQuestionElement = document.getElementById(id);
    if (existingQuestionElement) {
        existingQuestionElement.remove();
    } else {
        sortAscending.disabled = 1;
        sortAscending.parsing_error_count = -1;
        sortAscending.title = 1;
        sortAscending.total_count = -1;
        sortAscending.unread_count = -1;

        if ((typeof currentChecked === "undefined") || (typeof currentAscending === "undefined")) {
            currentChecked =  linkElement.dataset.defaultSortedBy;
            currentAscending = -1;
            if (linkElement.dataset.defaultSortDirection === "asc") {
                currentAscending = 1;
            }
        }

        sortAscending[currentChecked] = -currentAscending;
        let arrow = "&uarr; ";
        if (currentAscending === 1) {
            arrow = "&darr; ";
        }

        let questionElement = document.createElement("span");
        questionElement.setAttribute("id", id);
        questionElement.dataset.labelLoading = linkElement.dataset.labelLoading;
        questionElement.className = "confirm";

        let arrowElement = document.createElement("a");
        arrowElement.setAttribute("id", arrowElementId);
        arrowElement.innerHTML = arrow;
        questionElement.appendChild(arrowElement);
        createSortElement(questionElement, linkElement.dataset.sortedByDisabled, compareDisabled, "disabled", false);
        createSortElement(questionElement, linkElement.dataset.sortedByErrors, compareErrors, "parsing_error_count", false);
        createSortElement(questionElement, linkElement.dataset.sortedByTitle, compareTitle, "title", false);
        createSortElement(questionElement, linkElement.dataset.sortedByTotal, compareTotal, "total_count", false);
        createSortElement(questionElement, linkElement.dataset.sortedByUnread, compareUnread, "unread_count", true);

        containerElement.appendChild(questionElement);
    }
}