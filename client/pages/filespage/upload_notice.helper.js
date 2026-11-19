// The pipeline only picks up what lands under inbox/ at the root of the connection.
export function isInboxPath(path) {
    return /^\/inbox(\/|$)/.test(path || "");
}
