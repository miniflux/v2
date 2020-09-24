self.addEventListener("fetch", (event) => {
    if (event.request.url.includes("/feed/icon/")) {
        event.respondWith(
            caches.open("feed_icons").then((cache) => {
                return cache.match(event.request).then((response) => {
                    return response || fetch(event.request).then((response) => {
                        cache.put(event.request, response.clone());
                        return response;
                    });
                });
            })
        );
    }
});
