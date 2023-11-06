class WebAuthnHandler {
    static isWebAuthnSupported() {
        return window.PublicKeyCredential;
    }

    static showErrorMessage(errorMessage) {
        console.log("webauthn error: " + errorMessage);
        let alertElement = document.getElementById("webauthn-error");
        if (alertElement) {
            alertElement.textContent += " (" + errorMessage + ")";
            alertElement.classList.remove("hidden");
        }
    }

    async isConditionalLoginSupported() {
        return WebAuthnHandler.isWebAuthnSupported() &&
            window.PublicKeyCredential.isConditionalMediationAvailable &&
            window.PublicKeyCredential.isConditionalMediationAvailable();
    }

    async conditionalLogin(abortController) {
        if (await this.isConditionalLoginSupported()) {
            this.login("", abortController);
        }
    }

    decodeBuffer(value) {
        return Uint8Array.from(atob(value.replace(/-/g, "+").replace(/_/g, "/")), c => c.charCodeAt(0));
    }

    encodeBuffer(value) {
        return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
            .replace(/\+/g, "-")
            .replace(/\//g, "_")
            .replace(/=/g, "");
    }

    async post(urlKey, username, data) {
        let url = document.body.dataset[urlKey];
        if (username) {
            url += "?username=" + username;
        }

        return fetch(url, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "X-Csrf-Token": getCsrfToken()
            },
            body: JSON.stringify(data),
        });
    }

    async get(urlKey, username) {
        let url = document.body.dataset[urlKey];
        if (username) {
            url += "?username=" + username;
        }
        return fetch(url);
    }

    async removeAllCredentials() {
        try {
            await this.post("webauthnDeleteAllUrl", null, {});
        } catch (err) {
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

        window.location.reload();
    }

    async register() {
        let registerBeginResponse;
        try {
            registerBeginResponse = await this.get("webauthnRegisterBeginUrl");
        } catch (err) {
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

        let credentialCreationOptions = await registerBeginResponse.json();
        credentialCreationOptions.publicKey.challenge = this.decodeBuffer(credentialCreationOptions.publicKey.challenge);
        credentialCreationOptions.publicKey.user.id = this.decodeBuffer(credentialCreationOptions.publicKey.user.id);
        if (Object.hasOwn(credentialCreationOptions.publicKey, 'excludeCredentials')) {
            credentialCreationOptions.publicKey.excludeCredentials.forEach((credential) => credential.id = this.decodeBuffer(credential.id));
        }

        let attestation = await navigator.credentials.create(credentialCreationOptions);

        let registrationFinishResponse;
        try {
            registrationFinishResponse = await this.post("webauthnRegisterFinishUrl", null, {
                id: attestation.id,
                rawId: this.encodeBuffer(attestation.rawId),
                type: attestation.type,
                response: {
                    attestationObject: this.encodeBuffer(attestation.response.attestationObject),
                    clientDataJSON: this.encodeBuffer(attestation.response.clientDataJSON),
                },
            });
        } catch (err) {
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

        if (!registrationFinishResponse.ok) {
            throw new Error("Login failed with HTTP status code " + response.status);
        }

        let jsonData = await registrationFinishResponse.json();
        window.location.href = jsonData.redirect;
    }

    async login(username, abortController) {
        let loginBeginResponse;
        try {
            loginBeginResponse = await this.get("webauthnLoginBeginUrl", username);
        } catch (err) {
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

        let credentialRequestOptions = await loginBeginResponse.json();
        credentialRequestOptions.publicKey.challenge = this.decodeBuffer(credentialRequestOptions.publicKey.challenge);

        if (Object.hasOwn(credentialRequestOptions.publicKey, 'allowCredentials')) {
            credentialRequestOptions.publicKey.allowCredentials.forEach((credential) => credential.id = this.decodeBuffer(credential.id));
        }

        if (abortController) {
            credentialRequestOptions.signal = abortController.signal;
            credentialRequestOptions.mediation = "conditional";
        }

        let assertion;
        try {
            assertion = await navigator.credentials.get(credentialRequestOptions);
        }
        catch (err) {
            // Swallow aborted conditional logins
            if (err instanceof DOMException && err.name == "AbortError") {
                return;
            }
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

        if (!assertion) {
            return;
        }

        let loginFinishResponse;
        try {
            loginFinishResponse = await this.post("webauthnLoginFinishUrl", username, {
                id: assertion.id,
                rawId: this.encodeBuffer(assertion.rawId),
                type: assertion.type,
                response: {
                    authenticatorData: this.encodeBuffer(assertion.response.authenticatorData),
                    clientDataJSON: this.encodeBuffer(assertion.response.clientDataJSON),
                    signature: this.encodeBuffer(assertion.response.signature),
                    userHandle: this.encodeBuffer(assertion.response.userHandle),
                },
            });
        } catch (err) {
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

        if (!loginFinishResponse.ok) {
            throw new Error("Login failed with HTTP status code " + loginFinishResponse.status);
        }

        window.location.reload();
    }
}
