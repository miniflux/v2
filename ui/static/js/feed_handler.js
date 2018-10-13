class FeedHandler {
    static unsubscribe(feedUrl, callback) {
        let request = new RequestBuilder(feedUrl);
        request.withCallback(callback);
        request.execute();
    }
}
