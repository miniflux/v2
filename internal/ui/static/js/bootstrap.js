document.addEventListener("DOMContentLoaded", () => {
    handleSubmitButtons();

    if (!document.querySelector("body[data-disable-keyboard-shortcuts=true]")) {
        const keyboardHandler = new KeyboardHandler();
        keyboardHandler.on("g u", () => goToPage("unread"));
        keyboardHandler.on("g b", () => goToPage("starred"));
        keyboardHandler.on("g h", () => goToPage("history"));
        keyboardHandler.on("g f", goToFeedOrFeeds);
        keyboardHandler.on("g c", () => goToPage("categories"));
        keyboardHandler.on("g s", () => goToPage("settings"));
        keyboardHandler.on("g g", () => goToPrevious(TOP));
        keyboardHandler.on("G", () => goToNext(BOTTOM));
        keyboardHandler.on("ArrowLeft", goToPrevious);
        keyboardHandler.on("ArrowRight", goToNext);
        keyboardHandler.on("k", goToPrevious);
        keyboardHandler.on("p", goToPrevious);
        keyboardHandler.on("j", goToNext);
        keyboardHandler.on("n", goToNext);
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
        keyboardHandler.on("F", goToFeed);
        keyboardHandler.on("R", handleRefreshAllFeeds);
        keyboardHandler.on("?", showKeyboardShortcuts);
        keyboardHandler.on("+", goToAddSubscription);
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

    const touchHandler = new TouchHandler();
    touchHandler.listen();

    if (WebAuthnHandler.isWebAuthnSupported()) {
        const webauthnHandler = new WebAuthnHandler();

        onClick("#webauthn-delete", () => { webauthnHandler.removeAllCredentials(); });

        const registerButton = document.getElementById("webauthn-register");
        if (registerButton != null) {
            registerButton.disabled = false;

            onClick("#webauthn-register", () => {
                webauthnHandler.register().catch((err) => WebAuthnHandler.showErrorMessage(err));
            });
        }

        const loginButton = document.getElementById("webauthn-login");
        if (loginButton != null) {
            const abortController = new AbortController();
            loginButton.disabled = false;

            onClick("#webauthn-login", () => {
                const usernameField = document.getElementById("form-username");
                if (usernameField != null) {
                    abortController.abort();
                    webauthnHandler.login(usernameField.value).catch(err => WebAuthnHandler.showErrorMessage(err));
                }
            });

            webauthnHandler.conditionalLogin(abortController).catch(err => WebAuthnHandler.showErrorMessage(err));
        }
    }

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
        if (event.button == 1) {
            handleEntryStatus("next", event.target, true);
        }
    }, true);

    checkMenuToggleModeByLayout();
    window.addEventListener("resize", checkMenuToggleModeByLayout, { passive: true });

    fixVoiceOverDetailsSummaryBug();

    const logoElement = document.querySelector(".logo");
    if (logoElement) {
        logoElement.addEventListener("click", toggleMainMenu);
        logoElement.addEventListener("keydown", toggleMainMenu);
    }

    onClick(".header nav li", (event) => onClickMainMenuListItem(event));

    if ("serviceWorker" in navigator) {
        const scriptElement = document.getElementById("service-worker-script");
        if (scriptElement) {
	    navigator.serviceWorker.register(ttpolicy.createScriptURL(scriptElement.src));
        }
    }

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

    // Save and resume media position
    const lastPositionElements = document.querySelectorAll("audio[data-last-position],video[data-last-position]");
    lastPositionElements.forEach((element) => {
        if (element.dataset.lastPosition) {
            element.currentTime = element.dataset.lastPosition;
        }
        element.ontimeupdate = () => handlePlayerProgressionSave(element);
    });

    // Set media playback rate
    const playbackRateElements = document.querySelectorAll("audio[data-playback-rate],video[data-playback-rate]");
    playbackRateElements.forEach((element) => {
        if (element.dataset.playbackRate) {
            element.playbackRate = element.dataset.playbackRate;
        }
    });
});
