var sort_ascending = 1;
function compareName(a,b) {
    if (a.dataset.name<b.dataset.name)return -sort_ascending;
    if (a.dataset.name>b.dataset.name)return sort_ascending;
    return 0;
}
function comparePublished(a,b) {
    if (a.dataset.published>b.dataset.published)return -sort_ascending;
    if (a.dataset.published<b.dataset.published)return sort_ascending;
}
function compareLastCheck(a,b) {
    if (a.dataset.lastcheck>b.dataset.lastcheck)return -sort_ascending;
    if (a.dataset.lastcheck<b.dataset.lastcheck)return sort_ascending;
    return 0;
}
function compareRead(a,b) {
    let a_value = parseInt(a.dataset.read);
    let b_value = parseInt(b.dataset.read);
    if (a_value<b_value)return sort_ascending;
    if (a_value>b_value)return -sort_ascending;
    return 0;
}
function compareUnread(a,b) {
    let a_value = parseInt(a.dataset.unread);
    let b_value = parseInt(b.dataset.unread);
    if (a_value<b_value)return sort_ascending;
    if (a_value>b_value)return -sort_ascending;
    return 0;
}
function compareTotal(a,b) {
    let a_value = parseInt(a.dataset.total);
    let b_value = parseInt(b.dataset.total);
    if (a_value<b_value)return sort_ascending;
    if (a_value>b_value)return -sort_ascending;
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
function createSortElement(containerElement, sortText, compareFunction) {
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
            sort_ascending = -sort_ascending;
            loadingElement.remove();
        },20);
    };

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
        sort_ascending = 1;

        let questionElement = document.createElement("span");
        questionElement.setAttribute("id", id);
        questionElement.dataset.labelLoading = linkElement.dataset.labelLoading;
        questionElement.className = "confirm";

        createSortElement(questionElement, linkElement.dataset.sortByName, compareName);
        questionElement.appendChild(document.createTextNode(", "));
        createSortElement(questionElement, linkElement.dataset.sortByPublished, comparePublished);
        questionElement.appendChild(document.createTextNode(", "));
        createSortElement(questionElement, linkElement.dataset.sortByLastcheck, compareLastCheck);
        questionElement.appendChild(document.createTextNode(", "));
        createSortElement(questionElement, linkElement.dataset.sortByRead, compareRead);
        questionElement.appendChild(document.createTextNode(", "));
        createSortElement(questionElement, linkElement.dataset.sortByUnread, compareUnread);
        questionElement.appendChild(document.createTextNode(", "));
        createSortElement(questionElement, linkElement.dataset.sortByTotal, compareTotal);
        questionElement.appendChild(document.createTextNode(" "));

        containerElement.appendChild(questionElement);
    }
}