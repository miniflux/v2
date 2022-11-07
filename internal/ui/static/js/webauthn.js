function isWebAuthnSupported() {
    return window.PublicKeyCredential;
}

async function isConditionalLoginSupported() {
    return isWebAuthnSupported() && 
     window.PublicKeyCredential.isConditionalMediationAvailable &&
     window.PublicKeyCredential.isConditionalMediationAvailable();
}

// URLBase64 to ArrayBuffer
function bufferDecode(value) {
    return Uint8Array.from(atob(value), c => c.charCodeAt(0));
}

// ArrayBuffer to URLBase64
function bufferEncode(value) {
    return btoa(String.fromCharCode.apply(null, new Uint8Array(value)))
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}

function getCsrfToken() {
    let element = document.querySelector("body[data-csrf-token]");
    if (element !== null) {
        return element.dataset.csrfToken;
    }
    return "";
}

async function post(urlKey, username, data) {
    var url = document.body.dataset[urlKey];
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

async function get(urlKey, username) {
    var url = document.body.dataset[urlKey];
    if (username) {
        url += "?username=" + username;
    }
    return fetch(url);
}

function showError(error) {
    console.log("webauthn error: " + error);
    let alert = document.getElementById("webauthn-error");
    if (alert) {
        alert.classList.remove("hidden");
    }
}

async function register() {
    let beginRegisterURL = "webauthnRegisterBeginUrl";
    let r = await get(beginRegisterURL);
    let credOptions = await r.json();
    credOptions.publicKey.challenge = bufferDecode(credOptions.publicKey.challenge);
    credOptions.publicKey.user.id = bufferDecode(credOptions.publicKey.user.id);
    let attestation = await navigator.credentials.create(credOptions);
    let cred = {
        id: attestation.id,
        rawId: bufferEncode(attestation.rawId),
        type: attestation.type,
        response: {
            attestationObject: bufferEncode(attestation.response.attestationObject),
            clientDataJSON: bufferEncode(attestation.response.clientDataJSON),
        },
    };
    let finishRegisterURL = "webauthnRegisterFinishUrl";
    let response = await post(finishRegisterURL, null, cred);
    if (!response.ok) {
        throw new Error("Login failed with HTTP status " + response.status);
    }
    console.log("registration successful");
    window.location.reload();
}

async function login(username) {
    let beginLoginURL = "webauthnLoginBeginUrl";
    let r = await get(beginLoginURL, username);
    let c = await r.json();
    let credOptions = {
        publicKey: {
            challenge: bufferDecode(c.publicKey.challenge)
        },
    };
    if (!username) {
        credOptions.signal = abortController.signal;
        credOptions.mediation = "conditional";
    }
    
    let assertion = await navigator.credentials.get(credOptions);
    if (!assertion) {
        return;
    }
    
    let assertionResponse = {
        id: assertion.id,
        rawId: bufferEncode(assertion.rawId),
        type: assertion.type,
        response: {
            authenticatorData: bufferEncode(assertion.response.authenticatorData),
            clientDataJSON: bufferEncode(assertion.response.clientDataJSON),
            signature: bufferEncode(assertion.response.signature),
            userHandle: bufferEncode(assertion.response.userHandle),
        },
    };
    
    let finishLoginURL = "webauthnLoginFinishUrl";
    let response = await post(finishLoginURL, username, assertionResponse);
    if (!response.ok) {
        throw new Error("Login failed with HTTP status " + response.status);
    }
    window.location.reload();
}

async function conditionalLogin() {
    if (isConditionalLoginSupported()) {
        login();
    }
}

async function removeCreds() {
    let removeCredsURL = "webauthnDeleteAllUrl";
    await post(removeCredsURL, null, {});
    window.location.reload();
}

let abortController = new AbortController();
document.addEventListener("DOMContentLoaded", function () {
    if (!isWebAuthnSupported()) {
        return;
    }

    let registerButton = document.getElementById("webauthn-register");
    if (registerButton != null) {
        registerButton.disabled = false;
        registerButton.addEventListener("click", (e) => {
            register().catch((err) => showError(err));
        });
    }

    let removeCredsButton = document.getElementById("webauthn-delete");
    if (removeCredsButton != null) {
        removeCredsButton.addEventListener("click", removeCreds);
    }
    
    let loginButton = document.getElementById("webauthn-login");
    if (loginButton != null) {
        loginButton.disabled = false;
        let usernameField = document.getElementById("form-username");
        if (usernameField != null) {
            usernameField.autocomplete += " webauthn";
        }
        let passwordField = document.getElementById("form-password");
        if (passwordField != null) {
            passwordField.autocomplete += " webauthn";
        }

        loginButton.addEventListener("click", (e) => {
            e.preventDefault();
            abortController.abort();
            login(usernameField.value).catch(err => showError(err));
        });
        
        conditionalLogin().catch(err => {
            // ignore aborted conditional login requests
            if (!(err instanceof DOMException)) {
                showError(err);
            }
        });
    }
});
