import { treeEntries, fileLink, visibleEntries, isOnPath, needsListing, isCurrentListing, fileCount } from "./folder_tree.helper";

describe("treeEntries", () => {
    it("lists folders first, then files, each sorted by name", () => {
        const results = [
            { name: "notes.txt", type: "file" },
            { name: "outbox", type: "directory" },
            { name: "a.csv", type: "file" },
            { name: "inbox", type: "directory" },
        ];
        expect(treeEntries(results)).toEqual([
            { name: "inbox", type: "directory" },
            { name: "outbox", type: "directory" },
            { name: "a.csv", type: "file" },
            { name: "notes.txt", type: "file" },
        ]);
    });
    it("drops dot-folders and dot-files unless hidden files are shown", () => {
        const results = [
            { name: ".git", type: "directory" },
            { name: ".env", type: "file" },
            { name: "inbox", type: "directory" },
            { name: "notes.txt", type: "file" },
        ];
        const visible = [{ name: "inbox", type: "directory" }, { name: "notes.txt", type: "file" }];
        expect(treeEntries(results, false)).toEqual(visible);
        expect(treeEntries(results)).toEqual(visible);
        expect(treeEntries(results, true).map((e) => e.name)).toEqual([".git", "inbox", ".env", "notes.txt"]);
    });
    it("returns an empty list for a missing response", () => {
        expect(treeEntries(undefined)).toEqual([]);
    });
});

describe("fileLink", () => {
    it("opens the file in the viewer", () => {
        expect(fileLink("/inbox/session_1/notes.txt")).toBe("/view/inbox/session_1/notes.txt");
    });
    it("escapes %, ? and # as the file list does", () => {
        expect(fileLink("/inbox/50% done?#1.txt")).toBe("/view/inbox/50%2525 done%3F%231.txt");
    });
});

describe("visibleEntries", () => {
    const dir = (name) => ({ name, type: "directory" });
    const file = (name) => ({ name, type: "file" });
    it("keeps every entry when the files fit", () => {
        const entries = [dir("raw"), file("a.txt"), file("b.txt")];
        expect(visibleEntries(entries, 2)).toEqual({ shown: entries, more: 0 });
    });
    it("keeps every folder and the first files up to the limit, and counts the rest", () => {
        const entries = [dir("raw"), dir("sorted"), file("a"), file("b"), file("c"), file("d")];
        expect(visibleEntries(entries, 2)).toEqual({
            shown: [dir("raw"), dir("sorted"), file("a"), file("b")],
            more: 2,
        });
    });
    it("never hides folders, however many there are", () => {
        const entries = [dir("a"), dir("b"), dir("c"), file("x")];
        expect(visibleEntries(entries, 1)).toEqual({ shown: entries, more: 0 });
    });
});

describe("isOnPath", () => {
    it("is true for the folder itself and anything below it", () => {
        expect(isOnPath("/inbox/", "/inbox/")).toBe(true);
        expect(isOnPath("/inbox/session_1/", "/inbox/")).toBe(true);
    });
    it("is false for a sibling sharing a name prefix", () => {
        expect(isOnPath("/inbox2/", "/inbox/")).toBe(false);
    });
});

describe("needsListing", () => {
    it("lists an open level that was never listed", () => {
        expect(needsListing(true, null, 0)).toBe(true);
    });
    it("does not list a closed level", () => {
        expect(needsListing(false, null, 0)).toBe(false);
    });
    it("does not list again at the same version", () => {
        expect(needsListing(true, 3, 3)).toBe(false);
    });
    it("lists again once the version moves on", () => {
        expect(needsListing(true, 3, 4)).toBe(true);
    });
});

describe("isCurrentListing", () => {
    it("applies the first listing", () => {
        expect(isCurrentListing(null, 0)).toBe(true);
    });
    it("applies a listing for the same or a newer version", () => {
        expect(isCurrentListing(2, 2)).toBe(true);
        expect(isCurrentListing(2, 3)).toBe(true);
    });
    it("ignores a listing for an older version", () => {
        expect(isCurrentListing(3, 2)).toBe(false);
    });
});

describe("fileCount", () => {
    it("counts the files, not the folders", () => {
        const entries = [
            { name: "raw", type: "directory" },
            { name: "a.txt", type: "file" },
            { name: "b.txt", type: "file" },
        ];
        expect(fileCount(entries)).toBe(2);
    });
    it("counts every file, past the drawing limit", () => {
        const entries = Array.from({ length: 450 }, (_, i) => ({ name: "f" + i, type: "file" }));
        expect(fileCount(entries)).toBe(450);
    });
    it("is zero for a folder with no files", () => {
        expect(fileCount([{ name: "raw", type: "directory" }])).toBe(0);
        expect(fileCount([])).toBe(0);
    });
});
