let ttpolicy;
if (window.trustedTypes && trustedTypes.createPolicy) {
    //TODO: use an allow-list for `createScriptURL`
    if (!ttpolicy) {
        ttpolicy = trustedTypes.createPolicy('ttpolicy', {
            createScriptURL: src => src,
            createHTML: html => html,
        });
    }
} else {
    ttpolicy = {
        createScriptURL: src => src,
        createHTML: html => html,
    };
}
