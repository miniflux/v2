class WebpushHandler {
    askPermissionToSubscribe() {
        return Notification.requestPermission().then((permissionResult) => {
            return 'granted' === permissionResult;
        }).catch(() => false)
    }

    subscribeUserToPush(registration) {
        const urlBase64ToUint8Array = (base64String) => {
            const padding = '='.repeat((4 - base64String.length % 4) % 4);
            const base64 = (base64String + padding)
                .replace(/\-/g, '+')
                .replace(/_/g, '/');

            const rawData = window.atob(base64);
            let outputArray = new Uint8Array(rawData.length);

            for (let i = 0; i < rawData.length; ++i) {
                outputArray[i] = rawData.charCodeAt(i);
            }

            return outputArray;
        }


        const subscribeOptions = {
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array (
                'BIcdVkENsKKa17UpS9MmblMGQv1lix2IOGgp8Ngw9ZY1Bz9yiTiWjAsetABASSgtnvRJpzIUjcSlv6PhYEvidzk'
            )
        };

        return registration.pushManager.subscribe(subscribeOptions);
    }

    registerSubscription(pushSubscription) {
        let scriptElement = document.getElementById("service-worker-script");
        let request = new RequestBuilder(scriptElement.dataset.addSubscriptionUrl)
            .withBody({
                subscription: JSON.stringify(pushSubscription)
            });

        request.execute()

        return pushSubscription;
    }
}
