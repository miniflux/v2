class WebAuthnHandler {
    static isWebAuthnSupported() {
        return typeof PublicKeyCredential !== "undefined";
    }

    static showErrorMessage(errorMessage) {
        console.error("WebAuthn error:", errorMessage);

        const alertElement = document.getElementById("webauthn-error-alert");
        if (alertElement) {
            alertElement.remove();
        }

        const alertTemplateElement = document.getElementById("webauthn-error");
        if (alertTemplateElement) {
            const clonedElement = alertTemplateElement.content.cloneNode(true);
            const errorMessageElement = clonedElement.getElementById("webauthn-error-message");
            if (errorMessageElement) {
                errorMessageElement.textContent = errorMessage;
            }
            alertTemplateElement.parentNode.insertBefore(clonedElement, alertTemplateElement);
        }
    }

    static async isConditionalLoginSupported() {
        return WebAuthnHandler.isWebAuthnSupported() &&
            window.PublicKeyCredential.isConditionalMediationAvailable &&
            await window.PublicKeyCredential.isConditionalMediationAvailable();
    }

    async conditionalLogin(abortController) {
        if (await WebAuthnHandler.isConditionalLoginSupported()) {
            return this.login("", abortController);
        }
    }

    decodeBuffer(value) {
        return Uint8Array.from(atob(value.replace(/-/g, "+").replace(/_/g, "/")), c => c.charCodeAt(0));
    }

    encodeBuffer(value) {
        return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
            .replace(/\+/g, "-")
            .replace(/\//g, "_")
            .replace(/=+$/g, "");
    }

    async post(urlKey, username, data) {
        let url = document.body.dataset[urlKey];
        if (username) {
            url += `?username=${encodeURIComponent(username)}`;
        }

        return sendPOSTRequest(url, data);
    }

    async get(urlKey, username) {
        let url = document.body.dataset[urlKey];
        if (username) {
            url += `?username=${encodeURIComponent(username)}`;
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

        let credentialCreationOptions;
        try {
            credentialCreationOptions = await registerBeginResponse.json();
        } catch (err) {
            WebAuthnHandler.showErrorMessage("Failed to parse registration options");
            return;
        }

        credentialCreationOptions.publicKey.challenge = this.decodeBuffer(credentialCreationOptions.publicKey.challenge);
        credentialCreationOptions.publicKey.user.id = this.decodeBuffer(credentialCreationOptions.publicKey.user.id);
        if (Object.hasOwn(credentialCreationOptions.publicKey, 'excludeCredentials')) {
            credentialCreationOptions.publicKey.excludeCredentials.forEach((credential) => {
                credential.id = this.decodeBuffer(credential.id);
            });
        }

        let attestation;
        try {
            attestation = await navigator.credentials.create(credentialCreationOptions);
        } catch (err) {
            WebAuthnHandler.showErrorMessage(err);
            return;
        }

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
            throw new Error(`Registration failed with HTTP status code ${registrationFinishResponse.status}`);
        }

        const jsonData = await registrationFinishResponse.json();
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

        let credentialRequestOptions;
        try {
            credentialRequestOptions = await loginBeginResponse.json();
        } catch (err) {
            WebAuthnHandler.showErrorMessage("Failed to parse login options");
            return;
        }

        credentialRequestOptions.publicKey.challenge = this.decodeBuffer(credentialRequestOptions.publicKey.challenge);

        if (Object.hasOwn(credentialRequestOptions.publicKey, 'allowCredentials')) {
            credentialRequestOptions.publicKey.allowCredentials.forEach((credential) => {
                credential.id = this.decodeBuffer(credential.id);
            });
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
            if (err instanceof DOMException && err.name === "AbortError") {
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
            throw new Error(`Login failed with HTTP status code ${loginFinishResponse.status}`);
        }

        window.location.reload();
    }
}
