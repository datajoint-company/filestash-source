import { isInboxPath } from "./upload_notice.helper";

describe("isInboxPath", () => {
    it("matches inbox/ and anything below it", () => {
        expect(isInboxPath("/inbox/")).toBe(true);
        expect(isInboxPath("/inbox")).toBe(true);
        expect(isInboxPath("/inbox/session_1/")).toBe(true);
    });
    it("does not match other folders", () => {
        expect(isInboxPath("/")).toBe(false);
        expect(isInboxPath("/outbox/")).toBe(false);
        expect(isInboxPath("/inbox_old/")).toBe(false);
        expect(isInboxPath("/data/inbox/")).toBe(false);
        expect(isInboxPath("")).toBe(false);
        expect(isInboxPath(undefined)).toBe(false);
    });
});
