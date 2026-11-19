import React from "react";

import { Container, Icon } from "../../components/";
import { t } from "../../locales/";
import "./upload_notice.scss";

export function UploadNotice({ text, onDismiss }) {
    if (!text) return null;
    return (
        <Container>
            <div className="component_upload_notice" role="status">
                <span className="upload_notice_text">{text}</span>
                <button
                    type="button"
                    className="upload_notice_dismiss"
                    aria-label={t("Dismiss")}
                    title={t("Dismiss")}
                    onClick={onDismiss}>
                    <Icon name="close" />
                </button>
            </div>
        </Container>
    );
}
