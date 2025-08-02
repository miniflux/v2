initializeMainMenuHandlers();
initializeFormHandlers();
initializeMediaPlayerHandlers();

// Initialize the keyboard shortcuts if enabled.
if (!document.querySelector("body[data-disable-keyboard-shortcuts=true]")) {
    const keyboardHandler = new KeyboardHandler();
    keyboardHandler.on("g u", () => goToPage("unread"));
    keyboardHandler.on("g b", () => goToPage("starred"));
    keyboardHandler.on("g h", () => goToPage("history"));
    keyboardHandler.on("g f", goToFeedOrFeedsPage);
    keyboardHandler.on("g c", () => goToPage("categories"));
    keyboardHandler.on("g s", () => goToPage("settings"));
    keyboardHandler.on("g g", () => goToPreviousPage(TOP));
    keyboardHandler.on("G", () => goToNextPage(BOTTOM));
    keyboardHandler.on("ArrowLeft", goToPreviousPage);
    keyboardHandler.on("ArrowRight", goToNextPage);
    keyboardHandler.on("k", goToPreviousPage);
    keyboardHandler.on("p", goToPreviousPage);
    keyboardHandler.on("j", goToNextPage);
    keyboardHandler.on("n", goToNextPage);
    keyboardHandler.on("h", () => goToPage("previous"));
    keyboardHandler.on("l", () => goToPage("next"));
    keyboardHandler.on("z t", scrollToCurrentItem);
    keyboardHandler.on("o", openSelectedItem);
    keyboardHandler.on("Enter", () => openSelectedItem());
    keyboardHandler.on("v", () => openOriginalLink(false));
    keyboardHandler.on("V", () => openOriginalLink(true));
    keyboardHandler.on("c", () => openCommentLink(false));
    keyboardHandler.on("C", () => openCommentLink(true));
    keyboardHandler.on("m", () => handleEntryStatus("next"));
    keyboardHandler.on("M", () => handleEntryStatus("previous"));
    keyboardHandler.on("A", markPageAsRead);
    keyboardHandler.on("s", () => handleSaveEntry());
    keyboardHandler.on("d", handleFetchOriginalContent);
    keyboardHandler.on("f", () => handleBookmark());
    keyboardHandler.on("F", goToFeedPage);
    keyboardHandler.on("R", handleRefreshAllFeeds);
    keyboardHandler.on("?", showKeyboardShortcuts);
    keyboardHandler.on("+", goToAddSubscriptionPage);
    keyboardHandler.on("#", unsubscribeFromFeed);
    keyboardHandler.on("/", () => goToPage("search"));
    keyboardHandler.on("a", () => {
        const enclosureElement = document.querySelector('.entry-enclosures');
        if (enclosureElement) {
            enclosureElement.toggleAttribute('open');
        }
    });
    keyboardHandler.on("Escape", () => ModalHandler.close());
    keyboardHandler.listen();
}

// Initialize the touch handler for mobile devices.
const touchHandler = new TouchHandler();
touchHandler.listen();

// Initialize click handlers.
onClick(":is(a, button)[data-save-entry]", (event) => handleSaveEntry(event.target));
onClick(":is(a, button)[data-toggle-bookmark]", (event) => handleBookmark(event.target));
onClick(":is(a, button)[data-fetch-content-entry]", handleFetchOriginalContent);
onClick(":is(a, button)[data-share-status]", handleShare);
onClick(":is(a, button)[data-action=markPageAsRead]", (event) => handleConfirmationMessage(event.target, markPageAsRead));
onClick(":is(a, button)[data-toggle-status]", (event) => handleEntryStatus("next", event.target));
onClick(":is(a, button)[data-confirm]", (event) => handleConfirmationMessage(event.target, (url, redirectURL) => {
    const request = new RequestBuilder(url);

    request.withCallback((response) => {
        if (redirectURL) {
            window.location.href = redirectURL;
        } else if (response && response.redirected && response.url) {
            window.location.href = response.url;
        } else {
            window.location.reload();
        }
    });

    request.execute();
}));

onClick("a[data-original-link='true']", (event) => {
    handleEntryStatus("next", event.target, true);
}, true);
onAuxClick("a[data-original-link='true']", (event) => {
    if (event.button === 1) {
        handleEntryStatus("next", event.target, true);
    }
}, true);

// Register the service worker if supported.
if ("serviceWorker" in navigator) {
    const serviceWorkerURL = document.body.dataset.serviceWorkerUrl;
    if (serviceWorkerURL) {
        navigator.serviceWorker.register(ttpolicy.createScriptURL(serviceWorkerURL), {
            type: "module"
        }).catch((error) => {
            console.error("Service Worker registration failed:", error);
        });
    }
}

// PWA install prompt handling.
window.addEventListener('beforeinstallprompt', (e) => {
    let deferredPrompt = e;
    const promptHomeScreen = document.getElementById('prompt-home-screen');
    if (promptHomeScreen) {
        promptHomeScreen.style.display = "block";

        const btnAddToHomeScreen = document.getElementById('btn-add-to-home-screen');
        if (btnAddToHomeScreen) {
            btnAddToHomeScreen.addEventListener('click', (e) => {
                e.preventDefault();
                deferredPrompt.prompt();
                deferredPrompt.userChoice.then(() => {
                    deferredPrompt = null;
                    promptHomeScreen.style.display = "none";
                });
            });
        }
    }
});

// PassKey handling.
if (WebAuthnHandler.isWebAuthnSupported()) {
    const webauthnHandler = new WebAuthnHandler();

    onClick("#webauthn-delete", () => { webauthnHandler.removeAllCredentials(); });

    const registerButton = document.getElementById("webauthn-register");
    if (registerButton !== null) {
        registerButton.disabled = false;

        onClick("#webauthn-register", () => {
            webauthnHandler.register().catch((err) => WebAuthnHandler.showErrorMessage(err));
        });
    }

    const loginButton = document.getElementById("webauthn-login");
    if (loginButton !== null) {
        const abortController = new AbortController();
        loginButton.disabled = false;

        onClick("#webauthn-login", () => {
            const usernameField = document.getElementById("form-username");
            if (usernameField !== null) {
                abortController.abort();
                webauthnHandler.login(usernameField.value).catch(err => WebAuthnHandler.showErrorMessage(err));
            }
        });

        webauthnHandler.conditionalLogin(abortController).catch(err => WebAuthnHandler.showErrorMessage(err));
    }
}