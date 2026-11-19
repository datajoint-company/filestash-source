import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";

import { Icon } from "../../components/";
import { http_get, URL_FILES, settings_get, settings_put } from "../../helpers/";
import { t } from "../../locales/";
import {
    treeEntries, fileLink, visibleEntries, fileCount, TREE_FILE_LIMIT, isOnPath, needsListing, isCurrentListing,
} from "./folder_tree.helper";

import "./folder_tree.scss";

// Each level is listed only when it is first expanded. Raw entries are kept so that
// toggling hidden files only re-filters, without listing again.
function listEntries(path) {
    return http_get("/api/files/ls?path=" + encodeURIComponent(path))
        .then((res) => res.results || [], () => []);
}

// Keeps the newest listing of `path` for the given tree version.
function useListing(path, version, enabled) {
    const [listing, setListing] = useState({ at: null, entries: null });
    useEffect(() => {
        if (!needsListing(enabled, listing.at, version)) return;
        listEntries(path).then((entries) => setListing((shown) => (
            isCurrentListing(shown.at, version) ? { at: version, entries } : shown
        )));
    }, [enabled, version]);
    return listing.entries;
}

function Node({ name, path, current, version, showHidden }) {
    const [open, setOpen] = useState(isOnPath(current, path));
    useEffect(() => {
        if (isOnPath(current, path)) setOpen(true);
    }, [current]);
    const entries = useListing(path, version, open);
    const children = entries && treeEntries(entries, showHidden);
    const files = children ? fileCount(children) : 0;
    return (
        <li>
            <div className={"row" + (current === path ? " active" : "")}>
                <button
                    type="button" className="caret" onClick={() => setOpen(!open)}
                    aria-expanded={open} aria-label={open ? t("Collapse") : t("Expand")}>
                    {open ? "▾" : "▸"}
                </button>
                <Link to={URL_FILES + path} title={name}><Icon name="directory" />{name}</Link>
                {files > 0 && (
                    <span className="count" title={files + " " + t(files === 1 ? "file" : "files")}>{files}</span>
                )}
            </div>
            {open && children && children.length > 0 && (
                <Level entries={children} parent={path} current={current} version={version} showHidden={showHidden} />
            )}
        </li>
    );
}

function FileRow({ name, path }) {
    return (
        <li>
            <div className="row file">
                <span className="caret-space" />
                <Link to={fileLink(path)} title={name}><Icon name="file" />{name}</Link>
            </div>
        </li>
    );
}

function Entry({ entry, parent, current, version, showHidden }) {
    if (entry.type === "file") return <FileRow name={entry.name} path={parent + entry.name} />;
    return (
        <Node
            name={entry.name} path={parent + entry.name + "/"} current={current}
            version={version} showHidden={showHidden} />
    );
}

function Level({ entries, parent, current, version, showHidden }) {
    const { shown, more } = visibleEntries(entries, TREE_FILE_LIMIT);
    return (
        <ul>
            {shown.map((e) => (
                <Entry
                    key={e.type + e.name} entry={e} parent={parent} current={current}
                    version={version} showHidden={showHidden} />
            ))}
            {more > 0 && (
                <li>
                    <div className="row more">
                        <span className="caret-space" />
                        <Link to={URL_FILES + parent}>{"… " + more + " " + t("more")}</Link>
                    </div>
                </li>
            )}
        </ul>
    );
}

// `version` changes after a file or folder is created, renamed or deleted, and when the file list is refreshed (as after an upload); open levels are then listed again.
export function FolderTree({ path, version = 0, showHidden = false }) {
    const [collapsed, setCollapsed] = useState(settings_get("filespage_tree_collapsed") === true);
    const toggle = () => {
        settings_put("filespage_tree_collapsed", !collapsed);
        setCollapsed(!collapsed);
    };
    const entries = useListing("/", version, true);
    const roots = entries ? treeEntries(entries, showHidden) : [];
    return (
        <nav className={"component_folder_tree no-select" + (collapsed ? " collapsed" : "")}>
            <div className="title">
                {collapsed ? null : <span>{t("Files")}</span>}
                <button type="button" onClick={toggle} title={collapsed ? t("Show files") : t("Hide files")}>
                    {collapsed ? "»" : "«"}
                </button>
            </div>
            {collapsed ? null : (
                <Level entries={roots} parent="/" current={path} version={version} showHidden={showHidden} />
            )}
        </nav>
    );
}
