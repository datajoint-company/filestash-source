export function nop() {}

export function leftPad(str, length, pad = "0") {
    if (typeof str !== "string" || typeof pad !== "string" || str.length >= length ||
       !pad.length > 0) {
        return str;
    }
    return leftPad(pad + str, length, pad);
}

// Resolves to whether the text reached the clipboard. navigator.clipboard needs a secure
// context (and clipboard-write permission inside an iframe); anything else uses execCommand.
export function copyToClipboard(str) {
    if (!str) return Promise.resolve(false);
    if (navigator.clipboard && window.isSecureContext) {
        return navigator.clipboard.writeText(str).then(() => true, () => execCopy(str));
    }
    return Promise.resolve(execCopy(str));
}

// execCommand reports failure by returning false, not by throwing.
function execCopy(str) {
    const $t = document.createElement("textarea");
    $t.value = str;
    $t.setAttribute("readonly", "");
    $t.style.position = "fixed";
    $t.style.opacity = "0";
    document.body.appendChild($t);
    $t.select();
    let copied = false;
    try {
        copied = document.execCommand("copy");
    } catch (err) {
        copied = false;
    }
    document.body.removeChild($t);
    return copied === true;
}

export function format(str = "") {
    if (str.length === 0) return str;
    return str.split("_")
        .map((word, index) => {
            if (index != 0) return word;
            return (word[0] || "").toUpperCase() + word.substring(1);
        })
        .join(" ");
}

export function objectGet(obj, paths) {
    let value = obj;
    for (let i=0; i<paths.length; i++) {
        if (typeof value !== "object" || value === null) return null;
        value = value[paths[i]];
    }
    return value;
}
