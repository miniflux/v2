/**
 * Unit tests for handleConfirmationMessage function
 *
 * Run with: node app_test.js
 *
 * Tests the bulk operations confirmation bypass logic.
 */

// Mock DOM environment
global.document = {
    createTextNode: (text) => ({ textContent: text }),
    createElement: (tag) => ({
        tagName: tag.toUpperCase(),
        className: '',
        style: {},
        dataset: {},
        appendChild: function(child) { this.children = this.children || []; this.children.push(child); },
        remove: function() {},
    }),
};

// The handleConfirmationMessage function extracted from app.js
function handleConfirmationMessage(linkElement, callback) {
    if (linkElement.tagName !== 'A' && linkElement.tagName !== "BUTTON") {
        linkElement = linkElement.parentNode;
    }

    // If confirmation labels are not present, execute callback immediately
    // (user has disabled confirmation prompts in settings)
    if (linkElement.dataset.labelQuestion === undefined) {
        const url = linkElement.dataset.url;
        const redirectUrl = linkElement.dataset.redirectUrl;
        if (url) {
            callback(url, redirectUrl);
        } else {
            callback();
        }
        return;
    }

    linkElement.style.display = "none";

    const containerElement = linkElement.parentNode;
    const questionElement = document.createElement("span");

    function createLoadingElement() {
        const loadingElement = document.createElement("span");
        loadingElement.className = "loading";
        loadingElement.appendChild(document.createTextNode(linkElement.dataset.labelLoading));

        questionElement.remove();
        containerElement.appendChild(loadingElement);
    }

    const yesElement = document.createElement("button");
    yesElement.appendChild(document.createTextNode(linkElement.dataset.labelYes));
    yesElement.onclick = (event) => {
        event.preventDefault();

        createLoadingElement();

        callback(linkElement.dataset.url, linkElement.dataset.redirectUrl);
    };

    const noElement = document.createElement("button");
    noElement.appendChild(document.createTextNode(linkElement.dataset.labelNo));
    noElement.onclick = (event) => {
        event.preventDefault();

        const noActionUrl = linkElement.dataset.noActionUrl;
        if (noActionUrl) {
            createLoadingElement();

            callback(noActionUrl, linkElement.dataset.redirectUrl);
        } else {
            linkElement.style.display = "inline";
            questionElement.remove();
        }
    };

    questionElement.className = "confirm";
    questionElement.appendChild(document.createTextNode(`${linkElement.dataset.labelQuestion} `));
    questionElement.appendChild(yesElement);
    questionElement.appendChild(document.createTextNode(", "));
    questionElement.appendChild(noElement);

    containerElement.appendChild(questionElement);
}

// Test runner
let testsPassed = 0;
let testsFailed = 0;

function test(name, fn) {
    try {
        fn();
        console.log(`✓ ${name}`);
        testsPassed++;
    } catch (e) {
        console.log(`✗ ${name}`);
        console.log(`  Error: ${e.message}`);
        testsFailed++;
    }
}

function assertEqual(actual, expected, msg) {
    if (actual !== expected) {
        throw new Error(`${msg || 'Assertion failed'}: expected ${expected}, got ${actual}`);
    }
}

function assertTrue(value, msg) {
    if (!value) {
        throw new Error(msg || 'Expected true, got false');
    }
}

// Tests
console.log('\nRunning handleConfirmationMessage tests...\n');

// Test 1: When labelQuestion is missing, callback executes immediately (no confirmation)
test('executes callback immediately when labelQuestion is missing (bulk confirm disabled)', () => {
    let callbackCalled = false;
    let callbackUrl = null;
    let callbackRedirectUrl = null;

    const mockElement = {
        tagName: 'BUTTON',
        style: {},
        dataset: {
            // labelQuestion is intentionally missing - simulates DisableBulkOperationsConfirmations=true
            url: '/mark-all-as-read',
            redirectUrl: '/unread',
            labelLoading: 'Loading...'
        },
        parentNode: {
            appendChild: () => {},
        }
    };

    handleConfirmationMessage(mockElement, (url, redirectUrl) => {
        callbackCalled = true;
        callbackUrl = url;
        callbackRedirectUrl = redirectUrl;
    });

    assertTrue(callbackCalled, 'Callback should be called immediately');
    assertEqual(callbackUrl, '/mark-all-as-read', 'URL should be passed to callback');
    assertEqual(callbackRedirectUrl, '/unread', 'Redirect URL should be passed to callback');
});

// Test 2: When labelQuestion is present, callback does NOT execute immediately (shows confirmation)
test('does not execute callback when labelQuestion is present (bulk confirm enabled)', () => {
    let callbackCalled = false;

    const mockElement = {
        tagName: 'BUTTON',
        style: { display: 'inline' },
        dataset: {
            labelQuestion: 'Are you sure?',
            labelYes: 'Yes',
            labelNo: 'No',
            labelLoading: 'Loading...',
            url: '/mark-all-as-read',
        },
        parentNode: {
            appendChild: () => {},
        }
    };

    handleConfirmationMessage(mockElement, () => {
        callbackCalled = true;
    });

    assertTrue(!callbackCalled, 'Callback should NOT be called immediately');
    assertEqual(mockElement.style.display, 'none', 'Element should be hidden to show confirmation UI');
});

// Test 3: Callback without URL argument (markPageAsRead action)
test('executes callback without URL when no URL dataset attribute', () => {
    let callbackCalled = false;
    let argsCount = 0;

    const mockElement = {
        tagName: 'BUTTON',
        style: {},
        dataset: {
            // No URL - simulates markPageAsRead action
            labelLoading: 'Loading...'
        },
        parentNode: {
            appendChild: () => {},
        }
    };

    handleConfirmationMessage(mockElement, (...args) => {
        callbackCalled = true;
        argsCount = args.length;
    });

    assertTrue(callbackCalled, 'Callback should be called');
    assertEqual(argsCount, 0, 'Callback should be called with no arguments when no URL');
});

// Test 4: Non-BUTTON/A elements traverse up to parent
test('traverses to parent when element is not A or BUTTON', () => {
    let callbackCalled = false;

    const parentNode = {
        tagName: 'BUTTON',
        style: {},
        dataset: {
            url: '/test',
            labelLoading: 'Loading...'
        },
        parentNode: {
            appendChild: () => {},
        }
    };

    const mockElement = {
        tagName: 'SPAN',
        parentNode: parentNode,
    };

    handleConfirmationMessage(mockElement, () => {
        callbackCalled = true;
    });

    assertTrue(callbackCalled, 'Callback should be called using parent element');
});

// Test 5: Empty string should NOT bypass confirmation (must be undefined)
test('does not bypass confirmation when labelQuestion is empty string', () => {
    let callbackCalled = false;

    const mockElement = {
        tagName: 'BUTTON',
        style: { display: 'inline' },
        dataset: {
            labelQuestion: '',  // Empty string - should still show confirmation
            labelYes: 'Yes',
            labelNo: 'No',
            labelLoading: 'Loading...',
            url: '/mark-all-as-read',
        },
        parentNode: {
            appendChild: () => {},
        }
    };

    handleConfirmationMessage(mockElement, () => {
        callbackCalled = true;
    });

    assertTrue(!callbackCalled, 'Callback should NOT be called for empty string');
    assertEqual(mockElement.style.display, 'none', 'Element should be hidden to show confirmation UI');
});

// Summary
console.log('\n' + '='.repeat(50));
console.log(`Tests passed: ${testsPassed}`);
console.log(`Tests failed: ${testsFailed}`);
console.log('='.repeat(50) + '\n');

process.exit(testsFailed > 0 ? 1 : 0);
