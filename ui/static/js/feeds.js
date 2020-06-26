var sort_ascending = {
    name:1,
    lastcheck:1,
    errors:-1,
    read:-1,
    unread:-1,
};
function compareName(a,b) {
    if (a.dataset.name<b.dataset.name)return -sort_ascending.name;
    if (a.dataset.name>b.dataset.name)return sort_ascending.name;
    return 0;
}
function compareLastCheck(a,b) {
    if (a.dataset.lastcheck>b.dataset.lastcheck)return -sort_ascending.lastcheck;
    if (a.dataset.lastcheck<b.dataset.lastcheck)return sort_ascending.lastcheck;
    return 0;
}
function compareErrors(a,b) {
    let a_value = parseInt(a.dataset.errors);
    let b_value = parseInt(b.dataset.errors);
    if (a_value<b_value)return -sort_ascending.errors;
    if (a_value>b_value)return sort_ascending.errors;
    return 0;
}
function compareRead(a,b) {
    let a_value = parseInt(a.dataset.read);
    let b_value = parseInt(b.dataset.read);
    if (a_value<b_value)return -sort_ascending.read;
    if (a_value>b_value)return sort_ascending.read;
    return 0;
}
function compareUnread(a,b) {
    let a_value = parseInt(a.dataset.unread);
    let b_value = parseInt(b.dataset.unread);
    if (a_value<b_value)return -sort_ascending.unread;
    if (a_value>b_value)return sort_ascending.unread;
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
function createSortElement(containerElement, sortText, compareFunction, ascendingKey, isLastElement) {
    let element = document.createElement("a");
    element.href = "#";
    element.appendChild(document.createTextNode(sortText));
    element.onclick = (event) => {
        event.preventDefault();

        let loadingElement = document.createElement("a");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(containerElement.dataset.labelLoading));
        containerElement.appendChild(loadingElement);

        setTimeout(function(){
            sortFeedsBy(compareFunction);
            sort_ascending[ascendingKey] = -sort_ascending[ascendingKey];
            loadingElement.remove();
        },20);
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
        sort_ascending.name = 1;
        sort_ascending.lastcheck = 1;
        sort_ascending.errors = -1;
        sort_ascending.read = -1;
        sort_ascending.unread = -1;

        let questionElement = document.createElement("span");
        questionElement.setAttribute("id", id);
        questionElement.dataset.labelLoading = linkElement.dataset.labelLoading;
        questionElement.className = "confirm";

        createSortElement(questionElement, linkElement.dataset.sortByName, compareName, "name", false);
        createSortElement(questionElement, linkElement.dataset.sortByLastcheck, compareLastCheck, "lastcheck", false);
        createSortElement(questionElement, linkElement.dataset.sortByErrors, compareErrors, "errors", false);
        createSortElement(questionElement, linkElement.dataset.sortByRead, compareRead, "read", false);
        createSortElement(questionElement, linkElement.dataset.sortByUnread, compareUnread, "unread", true);

        containerElement.appendChild(questionElement);
    }
}
