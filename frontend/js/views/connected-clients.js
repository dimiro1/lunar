/**
 * @fileoverview Connected clients view for managing API tokens.
 */

import { API } from "../api.js";
import { t } from "../i18n/index.js";
import { Button, ButtonVariant } from "../components/button.js";
import { Card, CardContent, CardHeader } from "../components/card.js";
import { Badge, BadgeSize, BadgeVariant } from "../components/badge.js";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "../components/table.js";

function formatDate(timestamp) {
  if (!timestamp) return t("clients.neverUsed");
  return new Date(timestamp * 1000).toLocaleString();
}

export const ConnectedClients = {
  oninit() {
    this.tokens = [];
    this.loading = true;
    this.error = null;
    this.loadTokens();
  },

  async loadTokens() {
    try {
      const response = await API.tokens.list();
      this.tokens = response.tokens || [];
      this.loading = false;
    } catch (e) {
      this.error = e.message;
      this.loading = false;
    }
    m.redraw();
  },

  async revokeToken(id) {
    if (!confirm(t("clients.revokeConfirm"))) return;
    try {
      await API.tokens.revoke(id);
      await this.loadTokens();
    } catch (e) {
      this.error = e.message;
      m.redraw();
    }
  },

  view() {
    return m(".fade-in", [
      m(".page-header", [
        m(".page-header__title", [
          m("div", [
            m("h1", t("clients.title")),
          ]),
        ]),
      ]),

      m(Card, [
        m(CardHeader, {
          title: t("clients.allClients"),
          subtitle: t("clients.totalCount", { count: this.tokens.length }),
        }),

        this.loading
          ? m(CardContent, m("p", t("common.loading")))
          : this.error
          ? m(CardContent, m("p.text--danger", this.error))
          : this.tokens.length === 0
          ? m(CardContent, m("p.text-muted", t("clients.noClients")))
          : m(Table, [
            m(TableHeader, [
              m(TableRow, [
                m(TableHead, t("clients.name")),
                m(TableHead, t("clients.created")),
                m(TableHead, t("clients.lastUsed")),
                m(TableHead, t("clients.status")),
                m(TableHead, t("clients.actions")),
              ]),
            ]),
            m(
              TableBody,
              this.tokens.map((token) =>
                m(TableRow, { key: token.id }, [
                  m(TableCell, { mono: true }, token.name),
                  m(TableCell, formatDate(token.created_at)),
                  m(TableCell, formatDate(token.last_used)),
                  m(
                    TableCell,
                    m(
                      Badge,
                      {
                        variant: token.revoked
                          ? BadgeVariant.DEFAULT
                          : BadgeVariant.SUCCESS,
                        size: BadgeSize.SM,
                      },
                      token.revoked
                        ? t("clients.revoked")
                        : t("clients.active"),
                    ),
                  ),
                  m(
                    TableCell,
                    !token.revoked &&
                      m(
                        Button,
                        {
                          variant: ButtonVariant.DESTRUCTIVE,
                          size: "sm",
                          onclick: (e) => {
                            e.stopPropagation();
                            this.revokeToken(token.id);
                          },
                        },
                        t("clients.revoke"),
                      ),
                  ),
                ])
              ),
            ),
          ]),
      ]),
    ]);
  },
};
