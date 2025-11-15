import { Toast } from './toast.js';

// ID Pill component - displays IDs in monospace with click to copy
export const IdPill = {
  view: (vnode) => {
    const id = vnode.attrs.id;

    return m(
      ".id-pill",
      {
        onclick: () => {
          navigator.clipboard
            .writeText(id)
            .then(() => {
              Toast.show("Copied to clipboard", "success");
            })
            .catch(() => {
              Toast.show("Failed to copy", "error");
            });
        },
        style: "cursor: pointer; transition: all 0.15s;",
        title: "Click to copy",
        onmouseover: (e) => {
          e.target.style.background = "#404040";
        },
        onmouseout: (e) => {
          e.target.style.background = "#262626";
        },
      },
      id,
    );
  },
};
