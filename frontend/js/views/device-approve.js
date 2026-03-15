/**
 * @fileoverview Device authorization approval view.
 * Allows logged-in users to approve or deny CLI device auth requests.
 */

import { API } from "../api.js";
import { t } from "../i18n/index.js";
import { Button, ButtonVariant } from "../components/button.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
import { CodeDisplay } from "../components/code-display.js";

function renderContainer(...children) {
  return m(
    ".fade-in",
    { style: "max-width: 480px; margin: 0 auto;" },
    children,
  );
}

export const DeviceApprove = {
  oninit(vnode) {
    this.code = vnode.attrs.code;
    this.loading = true;
    this.error = null;
    this.auth = null;
    this.completed = false;
    this.completedAction = null;
    this.submitting = false;

    this.loadStatus();
  },

  async loadStatus() {
    try {
      this.auth = await API.auth.getDeviceApproveStatus(this.code);
      this.loading = false;
    } catch (e) {
      this.error = e.message || t("deviceApprove.notFound");
      this.loading = false;
    }
    m.redraw();
  },

  async handleAction(action) {
    this.submitting = true;
    try {
      await API.auth.approveDevice(this.code, action);
      this.completed = true;
      this.completedAction = action;
    } catch (e) {
      this.error = e.message;
    }
    this.submitting = false;
    m.redraw();
  },

  view() {
    if (this.loading) {
      return renderContainer(
        m(Card, [
          m(CardHeader, { title: t("deviceApprove.title") }),
          m(CardContent, m("p", t("common.loading"))),
        ]),
      );
    }

    if (this.error) {
      return renderContainer(
        m(Card, [
          m(CardHeader, { title: t("deviceApprove.title") }),
          m(CardContent, m("p.text--danger", this.error)),
        ]),
      );
    }

    if (this.completed) {
      const message = this.completedAction === "allow"
        ? t("deviceApprove.approved")
        : t("deviceApprove.denied");

      return renderContainer(
        m(Card, [
          m(CardHeader, { title: t("deviceApprove.title") }),
          m(CardContent, m("p", message)),
        ]),
      );
    }

    if (this.auth.status !== "pending") {
      return renderContainer(
        m(Card, [
          m(CardHeader, { title: t("deviceApprove.title") }),
          m(CardContent, m("p", t("deviceApprove.expired"))),
        ]),
      );
    }

    return renderContainer(
      m(Card, [
        m(CardHeader, { title: t("deviceApprove.title") }),
        m(CardContent, [
          m(
            "p",
            { style: "margin-bottom: 1.5rem;" },
            t("deviceApprove.description"),
          ),
          m(CodeDisplay, {
            label: t("deviceApprove.userCode"),
            value: this.auth.user_code,
          }),
          m(
            "div",
            {
              style:
                "display: flex; gap: 0.75rem; justify-content: flex-end; margin-top: 1.5rem;",
            },
            [
              m(
                Button,
                {
                  variant: ButtonVariant.SECONDARY,
                  onclick: () => this.handleAction("deny"),
                  disabled: this.submitting,
                },
                t("deviceApprove.deny"),
              ),
              m(
                Button,
                {
                  variant: ButtonVariant.PRIMARY,
                  onclick: () => this.handleAction("allow"),
                  disabled: this.submitting,
                },
                t("deviceApprove.allow"),
              ),
            ],
          ),
        ]),
      ]),
    );
  },
};
