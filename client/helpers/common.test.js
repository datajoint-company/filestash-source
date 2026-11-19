import { copyToClipboard } from "./common";

// Defined once with a getter: on Node 18, redefining a property of jest's window does not
// change what the module under test reads.
let secureContext = true;
Object.defineProperty(window, "isSecureContext", { get: () => secureContext, configurable: true });

function setClipboard(writeText, secure) {
    Object.defineProperty(navigator, "clipboard", {
        value: writeText ? { writeText } : undefined, configurable: true,
    });
    secureContext = secure;
}

describe("copyToClipboard", () => {
    beforeEach(() => {
        document.execCommand = jest.fn(() => true);
    });

    it("uses the Clipboard API in a secure context", async () => {
        const writeText = jest.fn(() => Promise.resolve());
        setClipboard(writeText, true);
        await expect(copyToClipboard("inbox/a.txt")).resolves.toBe(true);
        expect(writeText).toHaveBeenCalledWith("inbox/a.txt");
        expect(document.execCommand).not.toHaveBeenCalled();
    });

    it("falls back to execCommand when the Clipboard API rejects", async () => {
        setClipboard(jest.fn(() => Promise.reject(new Error("denied"))), true);
        await expect(copyToClipboard("inbox/a.txt")).resolves.toBe(true);
        expect(document.execCommand).toHaveBeenCalledWith("copy");
    });

    it("uses execCommand outside a secure context", async () => {
        const writeText = jest.fn(() => Promise.resolve());
        setClipboard(writeText, false);
        await expect(copyToClipboard("inbox/a.txt")).resolves.toBe(true);
        expect(writeText).not.toHaveBeenCalled();
    });

    it("reports failure when execCommand returns false or throws", async () => {
        setClipboard(null, false);
        document.execCommand = jest.fn(() => false);
        await expect(copyToClipboard("inbox/a.txt")).resolves.toBe(false);
        document.execCommand = jest.fn(() => {
            throw new Error("blocked");
        });
        await expect(copyToClipboard("inbox/a.txt")).resolves.toBe(false);
    });

    it("removes its temporary textarea", async () => {
        setClipboard(null, false);
        await copyToClipboard("inbox/a.txt");
        expect(document.querySelectorAll("textarea").length).toBe(0);
    });

    it("copies nothing for an empty string", async () => {
        setClipboard(null, false);
        await expect(copyToClipboard("")).resolves.toBe(false);
        expect(document.execCommand).not.toHaveBeenCalled();
    });
});
