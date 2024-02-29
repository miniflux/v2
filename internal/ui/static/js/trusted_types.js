//TODO: this is catastrophic
const ttpolicy = trustedTypes.createPolicy("ttpolicy", {
	createScriptURL: () => document.getElementById("service-worker-script").src,
	createHTML : (data) => data, 
});
