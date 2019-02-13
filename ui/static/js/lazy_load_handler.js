class LazyloadHandler {
    static add(selector, onevent, callback) {
        var loadImages = lazyload(selector);
        if (!loadImages) return;
        loadImages();
        window.addEventListener('scroll', throttle(loadImages, 500, 1000), false);

        function throttle(fn, delay, atleast) {
            var timeout = null,
                startTime = new Date();
            return function () {
                var curTime = new Date();
                clearTimeout(timeout);
                if (curTime - startTime >= atleast) {
                    fn();
                    startTime = curTime;
                } else {
                    timeout = setTimeout(fn, delay);
                }
            }
        }

        function lazyload(selector) {
            let lastVisbileIndex = 0;
            return function () {
                let items = Array.from(document.querySelectorAll(selector));
                if (!items || !items.length) return null;
                let len = items.length;
                if (len <= lastVisbileIndex) return;

                let seeHeight = document.documentElement.clientHeight;
                let scrollTop = document.documentElement.scrollTop || document.body.scrollTop;

                let i = 0;
                for (i = lastVisbileIndex; i < len; i++) {
                    if (items[i].offsetTop > seeHeight + scrollTop) {
                        break;
                    }
                }
                if (i <= lastVisbileIndex) return;
                lastVisbileIndex = i;
                // pre-load images for the next 10 entries.
                items.slice(0, lastVisbileIndex + 10)
                    .reduce((imgs, item) => {
                        imgs.push(...Array.from(item.querySelectorAll('img.lazy')))
                        return imgs;
                    }, [])
                    .filter(img => img && img.dataset.src !== "")
                    .forEach(img => {
                        img.src = img.dataset.src;
                        img.dataset.src = "";
                        imagesLoaded(img).on(onevent, callback);
                    });
            }
        }
    }
}