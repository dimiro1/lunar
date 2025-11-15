// Toast notification system
export const Toast = {
  messages: [],
  nextId: 0,

  show: (message, type = 'success', duration = 2000) => {
    const id = Toast.nextId++;
    Toast.messages.push({ id, message, type });
    m.redraw();

    setTimeout(() => {
      Toast.remove(id);
    }, duration);
  },

  remove: (id) => {
    const index = Toast.messages.findIndex(msg => msg.id === id);
    if (index !== -1) {
      Toast.messages.splice(index, 1);
      m.redraw();
    }
  },

  view: () => {
    if (Toast.messages.length === 0) return null;

    return m('.toast-container',
      Toast.messages.map(msg =>
        m('.toast', {
          key: msg.id,
          class: `toast-${msg.type}`,
        }, [
          m('span', msg.message),
          m('button.toast-close', {
            onclick: () => Toast.remove(msg.id)
          }, '×')
        ])
      )
    );
  }
};
