import { Icons } from './icons.js';
import { Toast } from './components/toast.js';
import { FunctionsList } from './views/functions-list.js';
import { FunctionCreate } from './views/function-create.js';
import { FunctionDetail } from './views/function-detail.js';
import { FunctionEdit } from './views/function-edit.js';
import { FunctionEnv } from './views/function-env.js';
import { ExecutionDetail } from './views/execution-detail.js';
import { VersionDiff } from './views/version-diff.js';

// Layout component
const Layout = {
  view: (vnode) => [
    m(
      "header",
      m(".container", [
        m("a.logo", { href: "#!/functions" }, [
          Icons.moon(),
          " FaaS Dashboard",
        ]),
      ]),
    ),
    m("main", vnode.children),
    m(Toast),
  ],
};

// Routes
m.route(document.getElementById("app"), "/functions", {
  "/functions": {
    render: () => m(Layout, m(FunctionsList)),
  },
  "/functions/new": {
    render: () => m(Layout, m(FunctionCreate)),
  },
  "/functions/:id": {
    render: (vnode) => m(Layout, m(FunctionDetail, vnode.attrs)),
  },
  "/functions/:id/edit": {
    render: (vnode) => m(Layout, m(FunctionEdit, vnode.attrs)),
  },
  "/functions/:id/env": {
    render: (vnode) => m(Layout, m(FunctionEnv, vnode.attrs)),
  },
  "/functions/:id/diff/:v1/:v2": {
    render: (vnode) => m(Layout, m(VersionDiff, vnode.attrs)),
  },
  "/executions/:id": {
    render: (vnode) => m(Layout, m(ExecutionDetail, vnode.attrs)),
  },
});
