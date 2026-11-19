// One tree level from an /api/files/ls response: folders first, then files, each
// sorted by name. Dot-entries are dropped unless hidden files are shown, as in the file list.
export function treeEntries(results, showHidden) {
    const byName = (a, b) => a.name.localeCompare(b.name);
    const visible = (results || [])
        .filter((f) => f.type === "directory" || f.type === "file")
        .filter((f) => showHidden === true || f.name[0] !== ".")
        .map((f) => ({ name: f.name, type: f.type }));
    return visible.filter((f) => f.type === "directory").sort(byName)
        .concat(visible.filter((f) => f.type === "file").sort(byName));
}

// The viewer link for a file, escaped the way the file list escapes its links.
export function fileLink(path) {
    return "/view" + path
        .replace(/%/g, "%2525")
        .replace(/\?/g, "%3F")
        .replace(/#/g, "%23");
}

// Files listed per folder before the rest are summed up in a link to the folder.
export const TREE_FILE_LIMIT = 200;

// One level as drawn: every folder and at most `limit` files; `more` counts the files left out.
export function visibleEntries(entries, limit) {
    let files = 0;
    const shown = entries.filter((e) => e.type !== "file" || files++ < limit);
    return { shown, more: entries.length - shown.length };
}

// Both paths end in "/", so "/inbox/" is not treated as a parent of "/inbox2/".
export function isOnPath(current, folder) {
    return typeof current === "string" && current.startsWith(folder);
}

// An open level is listed again whenever the tree version has moved past the one it was listed at.
export function needsListing(open, listedAt, version) {
    return open === true && listedAt !== version;
}

// A listing applies only if it is not older than the one already shown; a slow
// response for an earlier version must not overwrite a newer one.
export function isCurrentListing(shownAt, listedAt) {
    return shownAt === null || listedAt >= shownAt;
}

export function fileCount(entries) {
    return entries.filter((e) => e.type === "file").length;
}
