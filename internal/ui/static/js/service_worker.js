// Incrementing OFFLINE_VERSION will kick off the install event and force
// previously cached resources to be updated from the network.
const OFFLINE_VERSION = 2;
const CACHE_NAME = "offline";

self.addEventListener("install", (event) => {
  event.waitUntil(
    (async () => {
      const cache = await caches.open(CACHE_NAME);

      if (USE_CACHE) {
        await cache.addAll(["/", "/unread", OFFLINE_URL]);
      } else {
        // Setting {cache: 'reload'} in the new request will ensure that the
        // response isn't fulfilled from the HTTP cache; i.e., it will be from
        // the network.
        await cache.add(new Request(OFFLINE_URL, { cache: "reload" }));
      }
    })(),
  );

  // Force the waiting service worker to become the active service worker.
  self.skipWaiting();
});

async function cacheFirstWithRefresh(request) {
  const cache = await caches.open(CACHE_NAME);

  const fetchResponsePromise = fetch(request).then(async (networkResponse) => {
    if (!networkResponse.ok) return networkResponse;

    const contentType = networkResponse.headers.get("Content-Type");
    if (!contentType || !contentType.includes("text/html")) {
      cache.put(request, networkResponse.clone());
      return networkResponse;
    }

    const text = await networkResponse.clone().text();

    const modifiedHtml = text.replace(/offline-hidden/g, "offline-visibe");

    const clonedResponse = new Response(modifiedHtml, {
      status: networkResponse.status,
      statusText: networkResponse.statusText,
      headers: networkResponse.headers,
    });

    cache.put(request, clonedResponse.clone());
    return networkResponse;
  });

  return (await cache.match(request)) || (await fetchResponsePromise);
}

self.addEventListener("fetch", (event) => {
  if (USE_CACHE) {
    return event.respondWith(cacheFirstWithRefresh(event.request));
  }

  // We proxify requests through fetch() only if we are offline because it's slower.
  if (navigator.onLine === false && event.request.mode === "navigate") {
    event.respondWith(
      (async () => {
        try {
          // Always try the network first.
          return await fetch(event.request);
        } catch (error) {
          // catch is only triggered if an exception is thrown, which is likely
          // due to a network error.
          // If fetch() returns a valid HTTP response with a response code in
          // the 4xx or 5xx range, the catch() will NOT be called.
          const cache = await caches.open(CACHE_NAME);
          return await cache.match(OFFLINE_URL);
        }
      })(),
    );
  }
});

self.addEventListener("load", async (event) => {
  if (
    navigator.onLine === true &&
    event.target.location.pathname === "/unread" &&
    USE_CACHE
  ) {
    const cache = await caches.open(CACHE_NAME);

    for (let article of document.getElementsByTagName("article")) {
      const as = article.getElementsByTagName("a");
      if (as.length > 0) {
        const a = as[0];
        const href = a.href;
        cache
          .add(
            new Request(href, {
              headers: new Headers({
                "Client-Type": "service-worker",
              }),
            }),
          )
          .then(() => {
            article;
          });
      }
    }
  }
});
