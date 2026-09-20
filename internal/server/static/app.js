import { Idiomorph } from './idiomorph.esm.js';

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text != null) node.textContent = text;
  return node;
}

/** Stable HTML id for Idiomorph matching (must be unique in the document). */
function morphId(...parts) {
  const body = parts
    .map((p) => String(p ?? '').replace(/[^a-zA-Z0-9_-]+/g, '_'))
    .filter(Boolean)
    .join('-');
  return `gb-${body || 'x'}`;
}

const ICONS = {
  github: `<svg class="host-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09 2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8"/></svg>`,
  gitlab: `<svg class="host-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M15.97 9.058l-.895-2.756L13.404.525a.507.507 0 0 0-.968-.007L10.764 6.3H5.236L3.564.518a.507.507 0 0 0-.968.007L.924 6.302.03 9.058a.98.98 0 0 0 .355 1.096l7.62 5.536 7.61-5.536a.98.98 0 0 0 .355-1.096M12.887 1.89l1.24 3.812H11.65zm-9.774 0 1.24 3.812H2.387zm-.72 5.072.976 3.005-2.392-1.738zm1.356 4.177L5.98 6.302h4.04l2.23 4.837L8 14.345zm7.688-1.172.976-3.005 1.416 1.267z"/></svg>`,
  branch: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M9.5 3.25a2.25 2.25 0 1 1 3 2.122V6A2.5 2.5 0 0 1 10 8.5H6a1 1 0 0 0-1 1v1.128a2.251 2.251 0 1 1-1.5 0V5.372a2.25 2.25 0 1 1 1.5 0v1.836A2.493 2.493 0 0 1 6 7h4a1 1 0 0 0 1-1v-.628A2.25 2.25 0 0 1 9.5 3.25M4.75 3.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5M11.5 3.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5m-7 9a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5"/></svg>`,
  star: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.751.751 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25"/></svg>`,
  lock: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M4 4a4 4 0 0 1 8 0v2h.25c.966 0 1.75.784 1.75 1.75v5.5A1.75 1.75 0 0 1 12.25 15h-8.5A1.75 1.75 0 0 1 2 13.25v-5.5C2 6.784 2.784 6 3.75 6H4Zm8.25 3.5h-8.5a.25.25 0 0 0-.25.25v5.5c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25v-5.5a.25.25 0 0 0-.25-.25ZM10.5 4v2h-5V4a2.5 2.5 0 0 1 5 0Z"/></svg>`,
  laptop: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M2.75 2.5a.25.25 0 0 0-.25.25v7.5c0 .138.112.25.25.25h10.5a.25.25 0 0 0 .25-.25v-7.5a.25.25 0 0 0-.25-.25ZM.75 2.75A1.75 1.75 0 0 1 2.5 1h11a1.75 1.75 0 0 1 1.75 1.75v7.5A1.75 1.75 0 0 1 13.5 12H2.5A1.75 1.75 0 0 1 .75 10.25ZM0 13.75a.75.75 0 0 1 .75-.75h14.5a.75.75 0 0 1 0 1.5H.75a.75.75 0 0 1-.75-.75Z"/></svg>`,
  clipboard: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M5.5 2.75h1.1a1.4 1.4 0 0 0 2.8 0h1.1A1.25 1.25 0 0 1 11.75 4v8.25A1.25 1.25 0 0 1 10.5 13.5h-5A1.25 1.25 0 0 1 4.25 12.25V4A1.25 1.25 0 0 1 5.5 2.75Z"/><path fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" d="M9.25 8H14m0 0-1.75-1.75M14 8l-1.75 1.75"/></svg>`,
  pr: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25m5.677-.425A.75.75 0 0 1 7.75 3h3.5a.75.75 0 0 1 0 1.5h-2.69l3.22 3.22a.75.75 0 0 1 0 1.06l-3.22 3.22h2.69a.75.75 0 0 1 0 1.5h-3.5a.75.75 0 0 1-.53-1.28L10.44 9 6.72 5.28a.75.75 0 0 1 .457-1.455M4 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5m0 9a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5"/></svg>`,
  external: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M3.75 2h3.5a.75.75 0 0 1 0 1.5h-3.5a.25.25 0 0 0-.25.25v8.5c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25v-3.5a.75.75 0 0 1 1.5 0v3.5A1.75 1.75 0 0 1 12.25 14h-8.5A1.75 1.75 0 0 1 2 12.25v-8.5C2 2.784 2.784 2 3.75 2Zm6.854-1h4.146a.25.25 0 0 1 .25.25v4.146a.25.25 0 0 1-.427.177L13.03 4.03 9.28 7.78a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042l3.75-3.75-1.543-1.543A.25.25 0 0 1 10.604 1Z"/></svg>`,
  alert: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"/></svg>`,
  clock: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Zm7-3.25v2.992l2.028.812a.75.75 0 0 1-.557 1.392l-2.5-1A.751.751 0 0 1 7 8.25v-3.5a.75.75 0 0 1 1.5 0Z"/></svg>`,
  draft: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M0 2.75C0 1.784.784 1 1.75 1h12.5c.966 0 1.75.784 1.75 1.75v1.5A1.75 1.75 0 0 1 14.25 6H1.75A1.75 1.75 0 0 1 0 4.25ZM1.75 2.5a.25.25 0 0 0-.25.25v1.5c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25v-1.5a.25.25 0 0 0-.25-.25ZM0 11.25c0-.966.784-1.75 1.75-1.75h12.5c.966 0 1.75.784 1.75 1.75v1.5A1.75 1.75 0 0 1 14.25 15H1.75A1.75 1.75 0 0 1 0 13.25Zm1.75-.25a.25.25 0 0 0-.25.25v1.5c0 .138.112.25.25.25h12.5a.25.25 0 0 0 .25-.25v-1.5a.25.25 0 0 0-.25-.25Z"/></svg>`,
  check: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0"/></svg>`,
  x: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M3.72 3.72a.75.75 0 0 1 1.06 0L8 6.94l3.22-3.22a.749.749 0 0 1 1.275.326.749.749 0 0 1-.215.734L9.06 8l3.22 3.22a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215L8 9.06l-3.22 3.22a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L6.94 8 3.72 4.78a.75.75 0 0 1 0-1.06"/></svg>`,
  dot: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 4a4 4 0 1 1 0 8 4 4 0 0 1 0-8"/></svg>`,
  trash: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M6.5 1.75a.25.25 0 0 1 .25-.25h2.5a.25.25 0 0 1 .25.25V3h-3ZM2.25 3.75a.75.75 0 0 1 0-1.5h11.5a.75.75 0 0 1 0 1.5H13v9.5A1.75 1.75 0 0 1 11.25 15h-6.5A1.75 1.75 0 0 1 3 13.25v-9.5Zm1.5 0v9.5c0 .138.112.25.25.25h6.5a.25.25 0 0 0 .25-.25v-9.5Zm2 1.75a.75.75 0 0 1 .75.75v5.5a.75.75 0 0 1-1.5 0v-5.5a.75.75 0 0 1 .75-.75Zm3 0a.75.75 0 0 1 .75.75v5.5a.75.75 0 0 1-1.5 0v-5.5a.75.75 0 0 1 .75-.75Z"/></svg>`,
  search: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M10.68 11.74a6 6 0 0 1-7.922-8.982 6 6 0 0 1 8.982 7.922l3.04 3.04a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215ZM11.5 7a4.499 4.499 0 1 0-8.997 0A4.499 4.499 0 0 0 11.5 7Z"/></svg>`,
  info: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M0 8a8 8 0 1 1 16 0A8 8 0 0 1 0 8Zm8-6.5a6.5 6.5 0 1 0 0 13 6.5 6.5 0 0 0 0-13ZM6.5 7.75A.75.75 0 0 1 7.25 7h1a.75.75 0 0 1 .75.75v2.75h.25a.75.75 0 0 1 0 1.5h-2a.75.75 0 0 1 0-1.5h.25v-2h-.25a.75.75 0 0 1-.75-.75ZM8 6a1 1 0 1 1 0-2 1 1 0 0 1 0 2Z"/></svg>`,
  pull: `<svg class="mark-icon" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8.75 1.75a.75.75 0 0 0-1.5 0v7.19L4.72 6.41a.75.75 0 0 0-1.06 1.06l3.75 3.75a.75.75 0 0 0 1.06 0l3.75-3.75a.75.75 0 0 0-1.06-1.06L8.75 8.94ZM2.75 13.5a.75.75 0 0 0 0 1.5h10.5a.75.75 0 0 0 0-1.5Z"/></svg>`,
  spinner: `<svg class="mark-icon mark-icon--spin" viewBox="0 0 16 16" aria-hidden="true"><path fill="currentColor" d="M8 1.5a6.5 6.5 0 1 0 6.5 6.5h-1.5A5 5 0 1 1 8 3V1.5Z"/></svg>`,
};

function setButtonLabel(btn, svg, label) {
  if (!btn) return;
  btn.replaceChildren();
  if (svg) btn.insertAdjacentHTML('beforeend', svg);
  const text = el('span', 'btn-label', label);
  btn.appendChild(text);
}

/** Disable a button and show the spinner glyph until setButtonIdle. */
function setButtonBusy(btn, label, title) {
  if (!btn) return;
  btn.disabled = true;
  btn.classList.add('is-busy');
  btn.setAttribute('aria-busy', 'true');
  setButtonLabel(btn, ICONS.spinner, label);
  if (title != null) btn.title = title;
}

/** Re-enable a button and restore its normal icon + label. */
function setButtonIdle(btn, { svg, label, title } = {}) {
  if (!btn) return;
  btn.disabled = false;
  btn.classList.remove('is-busy');
  btn.removeAttribute('aria-busy');
  if (svg != null || label != null) setButtonLabel(btn, svg, label);
  if (title != null) btn.title = title;
}

/** @type {null | ((ok: boolean) => void)} */
let modalResolve = null;
/** @type {null | ((ok: boolean) => boolean | Promise<boolean>)} */
let modalBeforeClose = null;

function closeModal(ok) {
  void closeModalAsync(ok);
}

async function closeModalAsync(ok) {
  if (modalBeforeClose) {
    let allow = true;
    try {
      allow = await modalBeforeClose(Boolean(ok));
    } catch (_) {
      allow = false;
    }
    if (!allow) return;
  }
  modalBeforeClose = null;
  const root = document.getElementById('modal-root');
  const modal = root?.querySelector('.modal');
  const cancelBtn = document.getElementById('modal-cancel');
  if (root) root.hidden = true;
  if (modal) {
    modal.classList.remove('modal--wide');
    modal.classList.remove('modal--manage');
  }
  const body = document.getElementById('modal-body');
  if (body) body.classList.remove('modal-body--rich');
  if (cancelBtn) cancelBtn.hidden = false;
  document.removeEventListener('keydown', onModalKeydown);
  const resolve = modalResolve;
  modalResolve = null;
  if (resolve) resolve(Boolean(ok));
}

function onModalKeydown(ev) {
  if (ev.key === 'Escape') {
    ev.preventDefault();
    closeModal(false);
    return;
  }
  if (ev.key !== 'Enter') return;
  const confirmBtn = document.getElementById('modal-confirm');
  const danger = confirmBtn?.classList.contains('modal-btn--danger');
  // Enter confirms only non-destructive dialogs.
  if (danger) return;
  // Manage is a multi-field editor; Enter must not dismiss it (textarea newlines, path fields).
  const modal = document.querySelector('#modal-root .modal');
  if (modal?.classList.contains('modal--manage')) return;
  const target = ev.target;
  if (
    target instanceof HTMLTextAreaElement
    || target instanceof HTMLSelectElement
    || (target instanceof HTMLInputElement && target.type !== 'button' && target.type !== 'submit')
  ) {
    return;
  }
  ev.preventDefault();
  closeModal(true);
}

function safeHref(url) {
  const raw = String(url || '').trim();
  if (!raw) return '';
  try {
    const u = new URL(raw, window.location.origin);
    if (u.protocol === 'http:' || u.protocol === 'https:') return u.href;
  } catch (_) {
    /* ignore */
  }
  return '';
}

function modalIsOpen() {
  const root = document.getElementById('modal-root');
  return Boolean(root && !root.hidden);
}

function setModalBodyContent(content) {
  const body = document.getElementById('modal-body');
  if (!body) return;
  body.replaceChildren();
  if (content == null) return;
  if (typeof content === 'string') {
    body.textContent = content;
    return;
  }
  if (content instanceof Node) body.appendChild(content);
}

function setModalDetailText(text) {
  const detail = document.getElementById('modal-detail');
  if (!detail) return;
  const detailText = String(text || '').trim();
  if (detailText) {
    detail.hidden = false;
    detail.textContent = detailText;
  } else {
    detail.hidden = true;
    detail.textContent = '';
  }
}

/**
 * Theme confirm dialog. Resolves true when confirmed.
 * @param {{ title: string, body?: string, bodyNode?: Node, detail?: string, confirmLabel?: string, cancelLabel?: string, danger?: boolean, info?: boolean, wide?: boolean, manage?: boolean, beforeClose?: (ok: boolean) => boolean | Promise<boolean> }} opts
 */
function confirmDialog(opts) {
  const root = document.getElementById('modal-root');
  const modal = root?.querySelector('.modal');
  const title = document.getElementById('modal-title');
  const body = document.getElementById('modal-body');
  const detail = document.getElementById('modal-detail');
  const cancelBtn = document.getElementById('modal-cancel');
  const confirmBtn = document.getElementById('modal-confirm');
  if (!root || !title || !body || !detail || !cancelBtn || !confirmBtn) {
    return Promise.resolve(window.confirm([opts.title, opts.body, opts.detail].filter(Boolean).join('\n\n')));
  }
  if (modalResolve) closeModal(false);

  const info = Boolean(opts.info);
  title.textContent = opts.title || (info ? 'Details' : 'Confirm');
  if (opts.bodyNode instanceof Node) {
    setModalBodyContent(opts.bodyNode);
  } else {
    setModalBodyContent(opts.body || '');
  }
  if (opts.manage || opts.bodyNode instanceof Node) {
    body.classList.add('modal-body--rich');
  } else {
    body.classList.remove('modal-body--rich');
  }
  setModalDetailText(opts.detail || '');

  cancelBtn.hidden = info;
  if (modal) {
    modal.classList.toggle('modal--wide', info || Boolean(opts.wide) || Boolean(opts.manage));
    modal.classList.toggle('modal--manage', Boolean(opts.manage));
  }

  const danger = info ? false : opts.danger !== false;
  confirmBtn.className = `modal-btn ${danger ? 'modal-btn--danger' : 'modal-btn--ok'}`;
  if (!info) {
    setButtonLabel(cancelBtn, ICONS.x, opts.cancelLabel || 'Cancel');
  }
  setButtonLabel(confirmBtn, danger ? ICONS.trash : ICONS.check, opts.confirmLabel || (info ? 'Close' : 'Confirm'));

  modalBeforeClose = typeof opts.beforeClose === 'function' ? opts.beforeClose : null;

  root.hidden = false;
  document.addEventListener('keydown', onModalKeydown);
  confirmBtn.focus();

  return new Promise((resolve) => {
    modalResolve = resolve;
  });
}

/** Info-only modal (single Close). */
function infoDialog(opts) {
  return confirmDialog({ ...opts, info: true, danger: false, confirmLabel: opts.confirmLabel || 'Close' });
}

/**
 * Confirm, then keep the modal open in a busy state until work() settles.
 * Returns false on cancel; true after work() completes (even if work throws, caller handles).
 */
async function confirmAndRun(opts, work) {
  const ok = await confirmDialog({
    title: opts.title,
    body: opts.body,
    detail: opts.detail,
    confirmLabel: opts.confirmLabel,
    cancelLabel: opts.cancelLabel,
    danger: opts.danger,
  });
  if (!ok) return false;

  const root = document.getElementById('modal-root');
  const modal = root?.querySelector('.modal');
  const title = document.getElementById('modal-title');
  const body = document.getElementById('modal-body');
  const cancelBtn = document.getElementById('modal-cancel');
  const confirmBtn = document.getElementById('modal-confirm');
  if (!root || !confirmBtn) {
    await work();
    return true;
  }

  if (title) title.textContent = opts.busyTitle || opts.title || 'Working…';
  if (body) setModalBodyContent(opts.busyBody || 'Please wait…');
  setModalDetailText(opts.detail || '');
  if (cancelBtn) {
    cancelBtn.hidden = true;
    cancelBtn.disabled = true;
  }
  confirmBtn.className = 'modal-btn modal-btn--ok';
  setButtonBusy(confirmBtn, opts.busyLabel || 'Working…');
  if (modal) modal.classList.toggle('modal--wide', Boolean(opts.wide));
  root.hidden = false;

  try {
    await work();
  } finally {
    root.hidden = true;
    if (modal) modal.classList.remove('modal--wide');
    setButtonIdle(confirmBtn);
    if (cancelBtn) {
      cancelBtn.hidden = false;
      cancelBtn.disabled = false;
    }
  }
  return true;
}

function bindModal() {
  const root = document.getElementById('modal-root');
  const cancelBtn = document.getElementById('modal-cancel');
  const confirmBtn = document.getElementById('modal-confirm');
  cancelBtn?.addEventListener('click', () => closeModal(false));
  confirmBtn?.addEventListener('click', () => closeModal(true));
  root?.querySelectorAll('[data-modal-dismiss]').forEach((node) => {
    node.addEventListener('click', () => closeModal(false));
  });
}

function iconMark(svg, className, title) {
  const wrap = el('span', `mark ${className}`);
  wrap.title = title;
  wrap.setAttribute('aria-label', title);
  wrap.insertAdjacentHTML('beforeend', svg);
  return wrap;
}

function ciMark(status) {
  const s = String(status || '').toLowerCase();
  if (['success', 'passed'].includes(s)) return iconMark(ICONS.check, 'mark--ok', s);
  if (['failed', 'failure', 'error', 'cancelled', 'canceled'].includes(s)) return iconMark(ICONS.x, 'mark--bad', s);
  if (['running', 'pending', 'in_progress', 'queued'].includes(s)) return iconMark(ICONS.dot, 'mark--run', s);
  if (!s) return null;
  return iconMark(ICONS.dot, 'mark--muted', s);
}

function hostPrefix(host, path) {
  const key = String(host || '').toLowerCase();
  const wrap = el('span', `project-host${key ? ` project-host--${key}` : ''}`);
  const label = key || 'unknown host';
  const tip = path ? `${label} · ${path}` : label;
  wrap.title = tip;
  wrap.setAttribute('aria-label', tip);
  if (ICONS[key]) {
    wrap.insertAdjacentHTML('beforeend', ICONS[key]);
  } else {
    wrap.appendChild(el('span', 'project-host-fallback', host || '?'));
  }
  return wrap;
}

function relativeTime(iso) {
  if (!iso) return '';
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return '';
  const sec = Math.round((Date.now() - t) / 1000);
  if (sec < 60) return 'just now';
  const min = Math.round(sec / 60);
  if (min < 60) return `${min}m ago`;
  const hr = Math.round(min / 60);
  if (hr < 48) return `${hr}h ago`;
  const day = Math.round(hr / 24);
  if (day < 60) return `${day}d ago`;
  const mo = Math.round(day / 30);
  return `${mo}mo ago`;
}

function localByBranch(local) {
  /** @type {Map<string, object[]>} */
  const map = new Map();
  if (!local || !local.mapped) return map;
  const trees = (Array.isArray(local.worktrees) ? local.worktrees : []).filter((w) => !w.bare);
  if (trees.length) {
    for (const wt of trees) {
      if (!wt.branch || wt.detached) continue;
      const list = map.get(wt.branch) || [];
      list.push(wt);
      map.set(wt.branch, list);
    }
    return map;
  }
  if (local.branch && !local.detached && !local.error) {
    map.set(local.branch, [{
      path: local.path,
      branch: local.branch,
      dirty: local.dirty,
      ahead: local.ahead,
      behind: local.behind,
      main: true,
      appearance_path: local.path,
    }]);
  }
  return map;
}

/** Prefer a worktree on branchName, else primary appearance, else main, else first. */
function pickLocalWorktree(wts, local, branchName) {
  const list = Array.isArray(wts) ? wts : (wts ? [wts] : []);
  if (!list.length) return null;
  const branch = String(branchName || '').trim();
  const pool = branch
    ? list.filter((w) => !w.detached && !w.bare && String(w.branch || '') === branch)
    : list;
  const use = pool.length ? pool : list;
  const primaryPath = String(local?.path || '');
  const primary = use.find((w) => primaryPath && w.appearance_path === primaryPath);
  if (primary) return primary;
  const main = use.find((w) => w.main);
  return main || use[0];
}

/** Origin sync for the appearance that owns wt; falls back to project-level sync. */
function originSyncForWorktree(local, branchName, wt) {
  const apps = localAppearances(local);
  const want = String(wt?.appearance_path || wt?.path || '').trim();
  if (want) {
    const app = apps.find((a) => String(a.path || '') === want);
    if (app) {
      return originSyncFor(app, branchName);
    }
  }
  return originSyncFor(local, branchName);
}

function originSyncFor(local, branchName) {
  const name = String(branchName || '');
  if (!name) return null;
  const list = Array.isArray(local?.origin_sync) ? local.origin_sync : [];
  return list.find((s) => s && s.name === name) || null;
}

function originSyncBits(sync) {
  const bits = [];
  if (!sync) return bits;
  if (sync.ahead) bits.push(`↑${sync.ahead}`);
  if (sync.behind) bits.push(`↓${sync.behind}`);
  return bits;
}

/** Drop stale ↓N while this checkout is pulling or waiting for a fresh board paint. */
function originSyncBitsForPull(sync, projectID, branchName, repoPath) {
  const bits = originSyncBits(sync);
  if (!sync?.behind || !projectID || !branchName || !repoPath) return bits;
  const key = pullActionKey(projectID, branchName, repoPath);
  if (!pendingPulls.has(key) && !recentPulled.has(key)) return bits;
  return bits.filter((b) => !String(b).startsWith('↓'));
}

function originSyncTitle(sync) {
  if (!sync) return '';
  const parts = [];
  if (sync.ahead) parts.push(`${sync.ahead} ahead of origin`);
  if (sync.behind) parts.push(`${sync.behind} behind origin`);
  if (parts.length) return parts.join(', ');
  return 'matches origin';
}

async function copyText(text, node) {
  const value = String(text || '').trim();
  if (!value) return;
  try {
    await navigator.clipboard.writeText(value);
  } catch {
    const ta = document.createElement('textarea');
    ta.value = value;
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.left = '-9999px';
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand('copy');
    } finally {
      ta.remove();
    }
  }
  if (!node) return;
  const prev = node.title;
  node.classList.add('is-copied');
  node.title = 'Copied';
  window.setTimeout(() => {
    node.classList.remove('is-copied');
    node.title = prev;
  }, 1000);
}

function copyPathButton(fullPath) {
  const path = String(fullPath || '').trim();
  if (!path) return null;
  const btn = el('button', 'copy-path-btn');
  btn.type = 'button';
  btn.title = `Copy path: ${path}`;
  btn.setAttribute('aria-label', btn.title);
  btn.dataset.action = 'copy-path';
  btn.dataset.path = path;
  btn.insertAdjacentHTML('beforeend', ICONS.clipboard);
  return btn;
}

function localColumn(wts, local, branchName, project) {
  const cell = el('div', 'branch-local');
  if (!local) {
    cell.appendChild(el('span', 'branch-local-empty', '—'));
    return cell;
  }
  if (!local.mapped) {
    cell.title = 'Local checkout not mapped';
    return cell;
  }
  const list = Array.isArray(wts) ? wts : (wts ? [wts] : []);
  const preferred = pickLocalWorktree(list, local, branchName);
  const repoPath = preferred?.path || local.path || '';
  const sync = originSyncForWorktree(local, branchName, preferred) || originSyncFor(local, branchName);
  const syncBits = originSyncBitsForPull(sync, project?.id, branchName, repoPath);

  if (!list.length) {
    if (local.error) {
      cell.appendChild(el('span', 'branch-local-empty', '!'));
      cell.title = local.error;
      return cell;
    }
    if (syncBits.length) {
      const span = el('span', 'branch-local-sync is-divergent', syncBits.join(' '));
      if (project?.id) span.id = morphId('sync', project.id, branchName, repoPath);
      span.title = `${originSyncTitle(sync)} (local branch not checked out)`;
      cell.appendChild(span);
      appendSyncInvestigateButton(cell, {
        project,
        branchName,
        repoPath,
      });
      appendPullButton(cell, {
        sync,
        project,
        branchName,
        repoPath,
        dirty: false,
      });
      return cell;
    }
    cell.appendChild(el('span', 'branch-local-empty', '—'));
    cell.title = sync
      ? 'Local branch not checked out (matches origin)'
      : 'No local checkout on this branch';
    return cell;
  }

  const marks = el('span', 'branch-local-marks');
  const primaryPath = String(local.path || '');
  for (const wt of list) {
    const label = wt.appearance_label || wt.appearance_path || wt.path || 'local checkout';
    const isPrimary = primaryPath && wt.appearance_path === primaryPath;
    let cls = 'mark--muted';
    let title = label;
    if (wt.dirty) {
      cls = 'mark--warn';
      title = `dirty: ${label}`;
    } else if (isPrimary || wt.main) {
      cls = 'mark--default';
    }
    marks.appendChild(iconMark(ICONS.laptop, cls, title));
  }
  if (preferred?.path) {
    const copyBtn = copyPathButton(preferred.path);
    if (copyBtn) marks.appendChild(copyBtn);
  }
  if (local.error) {
    marks.appendChild(iconMark(ICONS.alert, 'mark--bad', local.error));
    cell.title = local.error;
  }
  cell.appendChild(marks);

  // Skip ↑/↓ and pull when origin freshness is unknown (for example fetch failed).
  if (!local.error && syncBits.length) {
    const span = el('span', 'branch-local-sync is-divergent', syncBits.join(' '));
    if (project?.id) span.id = morphId('sync', project.id, branchName, repoPath);
    span.title = originSyncTitle(sync);
    cell.appendChild(span);
  }
  const syncIssue = Boolean(sync?.ahead || sync?.behind);
  const dirtyWorktree = list.some((wt) => wt.dirty);
  if (!local.error && (syncIssue || dirtyWorktree)) {
    appendSyncInvestigateButton(cell, {
      project,
      branchName,
      repoPath,
    });
  }
  if (!local.error) {
    appendPullButton(cell, {
      sync,
      project,
      branchName,
      repoPath,
      dirty: Boolean(preferred?.dirty),
      whyDirty: preferred?.dirty ? `Pull blocked: ${preferred?.appearance_label || preferred?.path || 'worktree'} is dirty` : '',
    });
  }
  return cell;
}

/** @type {Set<string>} */
const pendingPulls = new Set();

/** Keys that just finished pull successfully; cleared after a successful fresh board paint. */
/** @type {Set<string>} */
const recentPulled = new Set();

/** True until a fresh dashboard paint lands after one or more successful pulls. */
let pullBoardDirty = false;

/** Quiet polls must not abort an in-flight ?fresh=1 load (post-pull / Refresh). */
let dashboardFreshInFlight = false;

/** @type {{ label: string, path: string, message: string }[]} */
const pullFailures = [];

let pullFlushScheduled = false;

function pullStatusText() {
  const n = pendingPulls.size;
  if (n <= 0) return '';
  if (n === 1) return 'Pulling 1 checkout…';
  return `Pulling ${n} checkouts…`;
}

function schedulePullFlush() {
  if (pendingPulls.size > 0 || pullFlushScheduled) return;
  pullFlushScheduled = true;
  queueMicrotask(() => {
    void flushPullBatch();
  });
}

async function flushPullBatch() {
  pullFlushScheduled = false;
  if (pendingPulls.size > 0) return;
  const failures = pullFailures.splice(0, pullFailures.length);
  const status = document.getElementById('status');
  // Prefer a fresh load so origin ahead/behind matches the pulls that just finished.
  let ok = await loadDashboard({ quiet: true, fresh: true });
  if (!ok && pullBoardDirty) {
    ok = await loadDashboard({ quiet: true, fresh: true });
  }
  if (!ok && pullBoardDirty) {
    // Keep "pulled" until a later fresh paint; do not clear recentPulled on abort races.
    if (status && !String(status.textContent || '').startsWith('Error:')) {
      status.textContent = 'Pulls finished; refresh pending…';
    }
  }
  if (failures.length === 0) {
    if (status && !String(status.textContent || '').startsWith('Error:') && ok) {
      status.textContent = status.textContent || 'Pulls finished';
    }
    return;
  }
  if (status) {
    status.textContent = failures.length === 1
      ? `Pull failed: ${failures[0].message}`
      : `${failures.length} pulls failed`;
  }
  const body = failures.map((f) => {
    const where = f.path ? `${f.label}\n${f.path}` : f.label;
    return `${where}\n${f.message}`;
  }).join('\n\n');
  await infoDialog({
    title: failures.length === 1 ? `Pull failed: ${failures[0].label}` : `${failures.length} pulls failed`,
    body,
  });
}

function pullActionKey(projectID, branch, repoPath) {
  return `${projectID}\0${branch}\0${repoPath}`;
}

function appendPullButton(cell, { sync, project, branchName, repoPath, dirty, whyDirty }) {
  if (!project?.id || !branchName || !repoPath) return;
  const key = pullActionKey(project.id, branchName, repoPath);
  // Drop ephemeral "pulled" once this checkout is no longer behind.
  if (recentPulled.has(key) && (!sync?.behind || sync.ahead)) {
    recentPulled.delete(key);
    if (recentPulled.size === 0) pullBoardDirty = false;
  }
  if (!sync?.behind || sync.ahead) return;
  const pending = pendingPulls.has(key);
  const justPulled = recentPulled.has(key);
  const btn = el('button', 'branch-pull-ff');
  btn.type = 'button';
  btn.id = morphId('pull', project.id, branchName, repoPath);
  btn.dataset.action = 'pull';
  btn.dataset.projectId = project.id;
  btn.dataset.branch = branchName;
  btn.dataset.repoPath = repoPath;
  btn.dataset.behind = String(sync.behind);
  if (pending) {
    setButtonBusy(btn, 'pulling…', `Pulling ${branchName} from origin…`);
  } else if (justPulled) {
    setButtonLabel(btn, ICONS.pull, 'pulled');
    btn.disabled = true;
    btn.title = `Fast-forwarded ${branchName} from origin`;
  } else if (dirty) {
    setButtonLabel(btn, ICONS.pull, `pull ↓${sync.behind}`);
    btn.disabled = true;
    btn.title = whyDirty || `Pull blocked: working tree has local changes`;
    btn.classList.add('is-blocked');
  } else {
    setButtonLabel(btn, ICONS.pull, `pull ↓${sync.behind}`);
    btn.title = `Fast-forward local ${branchName} from origin (${sync.behind} behind)`;
  }
  cell.appendChild(btn);
}

function appendSyncInvestigateButton(cell, { project, branchName, repoPath }) {
  if (!project?.id || !branchName || !repoPath) return;
  const btn = el('button', 'branch-sync-inspect');
  btn.type = 'button';
  btn.id = morphId('inspect', project.id, branchName, repoPath);
  btn.dataset.action = 'inspect-sync';
  btn.dataset.projectId = project.id;
  btn.dataset.branch = branchName;
  btn.dataset.repoPath = repoPath;
  setButtonLabel(btn, ICONS.search, 'inspect');
  btn.title = `Investigate local ${branchName} versus origin`;
  cell.appendChild(btn);
}

/** Appearances list used for project title (same shape as renderAppearances). */
function localAppearances(local) {
  if (!local?.mapped) return [];
  if (Array.isArray(local.appearances) && local.appearances.length) {
    return local.appearances;
  }
  if (!local.path) return [];
  return [{
    path: local.path,
    display_id: local.path,
    primary: true,
    branch: local.branch,
    tag: local.tag,
    dirty: local.dirty,
    ahead: local.ahead,
    behind: local.behind,
    detached: local.detached,
    error: local.error,
    origin_sync: local.origin_sync,
    worktrees: local.worktrees,
  }];
}

function branchesCell(branches, host, project) {
  const cell = el('td', 'branches-cell');
  const local = project?.local;
  const byLocal = localByBranch(local);
  const list = el('ul', 'branch-list');

  const rows = Array.isArray(branches) ? branches : [];
  const remoteNames = new Set(rows.map((b) => b.name).filter(Boolean));
  if (!rows.length && byLocal.size === 0) {
    cell.appendChild(document.createTextNode('—'));
    return cell;
  }

  const reviewKind = String(host || '').toLowerCase() === 'gitlab' ? 'MR' : 'PR';
  const head = el('li', 'branch-item branch-item--head');
  head.appendChild(el('span', 'branch-marks'));
  head.appendChild(el('span', 'branch-head-label', 'branch'));
  head.appendChild(el('span', 'branch-head-label branch-head-label--local', 'local'));
  head.appendChild(el('span', 'branch-head-label branch-head-label--end', 'ci'));
  list.appendChild(head);

  for (const b of rows) {
    list.appendChild(branchRow({
      remote: b,
      localWts: byLocal.get(b.name) || [],
      local,
      host,
      project,
      reviewKind,
      localOnly: false,
    }));
    byLocal.delete(b.name);
  }

  for (const [name, wts] of byLocal) {
    if (remoteNames.has(name)) continue;
    if (branchNameHidden(name, hideBranchPatterns)) continue;
    list.appendChild(branchRow({
      remote: { name, ci_status: '', web_url: '' },
      localWts: wts,
      local,
      host,
      project,
      reviewKind,
      localOnly: true,
    }));
  }

  cell.appendChild(list);
  return cell;
}

function branchRow({ remote: b, localWts, local, host, project, reviewKind, localOnly }) {
  const localWt = pickLocalWorktree(localWts, local, b.name);
  const failed = ['failed', 'failure', 'error'].includes(String(b.ci_status || '').toLowerCase());
  const pruneHint = String(localWt?.prune_hint || '');
  const itemClass = [
    'branch-item',
    b.open_review ? 'branch-item--review' : '',
    b.conflict ? 'branch-item--conflict' : '',
    failed ? 'branch-item--ci-bad' : '',
    localOnly ? 'branch-item--local-only' : '',
    pruneHint === 'safe' ? 'branch-item--prune-safe' : '',
    pruneHint === 'likely' ? 'branch-item--prune-likely' : '',
  ].filter(Boolean).join(' ');
  const item = el('li', itemClass);
  if (project?.id && b.name) {
    item.id = morphId('br', project.id, b.name);
    item.dataset.projectId = project.id;
    item.dataset.branch = b.name;
  }

  const marks = el('span', 'branch-marks');
  if (b.open_review) marks.appendChild(iconMark(ICONS.pr, 'mark--review', `open ${reviewKind}`));
  if (b.draft) marks.appendChild(iconMark(ICONS.draft, 'mark--muted', 'draft'));
  if (b.conflict) marks.appendChild(iconMark(ICONS.alert, 'mark--bad', 'merge conflicts'));
  if (b.stale) marks.appendChild(iconMark(ICONS.clock, 'mark--stale', 'stale (>14 days)'));
  item.appendChild(marks);

  const body = el('div', 'branch-body');
  const title = el('div', 'branch-title');
  if (b.default) title.appendChild(iconMark(ICONS.lock, 'mark--default', 'default branch'));
  const nameHref = safeHref(b.web_url || localWt?.merged_url || '');
  const name = el(nameHref ? 'a' : 'span', 'branch-name', b.name || '');
  if (nameHref) {
    name.href = nameHref;
    name.target = '_blank';
    name.rel = 'noopener';
    name.title = localWt?.merged_url
      ? `Open merged ${reviewKind}`
      : (b.open_review ? `Open ${reviewKind}` : 'Open branch');
  }
  title.appendChild(name);
  body.appendChild(title);

  const meta = el('div', 'branch-meta');
  if (localOnly) meta.appendChild(el('span', 'branch-chip branch-chip--local', 'local only'));
  const pruneCmd = pruneCommand(b.name, localWt, local);
  if (pruneHint === 'safe') {
    const btn = el('button', 'branch-chip branch-chip--ok branch-prune-safe');
    btn.type = 'button';
    btn.id = morphId('prune', project.id, b.name, localWt?.path || local?.path || '');
    btn.dataset.action = 'prune-safe';
    btn.dataset.projectId = project.id;
    btn.dataset.branch = b.name;
    btn.dataset.worktreePath = localWt?.path || local?.path || '';
    setButtonLabel(btn, ICONS.trash, 'safe to remove');
    const bits = [];
    if (localWt.merged_id) bits.push(`merged ${reviewKind} #${localWt.merged_id}`);
    const mergedWhen = relativeTime(localWt.merged_at);
    if (mergedWhen) bits.push(mergedWhen);
    bits.push('Click to remove local branch');
    btn.title = bits.join(' · ');
    meta.appendChild(btn);
  } else if (pruneHint === 'likely') {
    const chip = el('span', 'branch-chip branch-chip--warn', 'likely removable');
    chip.title = pruneCmd
      ? `Remote head gone · check then: ${pruneCmd}`
      : 'Remote head gone';
    meta.appendChild(chip);
    const inv = el('button', 'branch-investigate');
    inv.type = 'button';
    inv.id = morphId('inv-prune', project.id, b.name, localWt?.path || local?.path || '');
    inv.dataset.action = 'investigate-prune';
    inv.dataset.projectId = project.id;
    inv.dataset.branch = b.name;
    inv.dataset.worktreePath = localWt?.path || local?.path || '';
    inv.dataset.defaultBranch = local?.default_branch || 'main';
    setButtonLabel(inv, ICONS.search, 'Investigate');
    inv.title = 'Gather evidence into an agent session';
    meta.appendChild(inv);
  }
  if (b.open_review) {
    meta.appendChild(el('span', 'branch-chip branch-chip--review', b.review_id ? `${reviewKind} #${b.review_id}` : reviewKind));
  }
  if (b.conflict) meta.appendChild(el('span', 'branch-chip branch-chip--bad', 'conflicts'));
  if (b.stale) meta.appendChild(el('span', 'branch-chip branch-chip--stale', 'stale'));
  if (b.draft) meta.appendChild(el('span', 'branch-chip', 'draft'));
  const when = relativeTime(localOnly && localWt?.merged_at ? localWt.merged_at : b.updated_at);
  if (when) meta.appendChild(el('span', 'branch-when', when));
  if (meta.childNodes.length) body.appendChild(meta);
  item.appendChild(body);

  item.appendChild(localColumn(localWts, local, b.name, project));

  const actions = el('div', 'branch-actions');
  const ci = ciMark(b.ci_status);
  if (ci) {
    if (failed && (b.run_id || project.ci?.run_id)) {
      const triageBtn = el('button', 'branch-triage');
      triageBtn.type = 'button';
      triageBtn.id = morphId('triage', project.id, b.name || '', b.run_id || project.ci?.run_id || '');
      triageBtn.dataset.action = 'triage';
      triageBtn.dataset.projectId = project.id;
      triageBtn.dataset.branch = b.name || '';
      triageBtn.dataset.runId = b.run_id || project.ci?.run_id || '';
      triageBtn.title = 'Show failed jobs and AI triage';
      triageBtn.setAttribute('aria-label', triageBtn.title);
      triageBtn.appendChild(ci);
      actions.appendChild(triageBtn);
    } else {
      actions.appendChild(ci);
    }
  }

  const ciURL = safeHref(b.ci_url);
  if (ciURL) {
    const go = el('a', 'branch-goto');
    go.href = ciURL;
    go.target = '_blank';
    go.rel = 'noopener';
    go.title = b.ci_status ? `Open CI run (${b.ci_status})` : 'Open CI run';
    go.setAttribute('aria-label', go.title);
    go.insertAdjacentHTML('beforeend', ICONS.external);
    actions.appendChild(go);
  } else if (nameHref) {
    const go = el('a', 'branch-goto');
    go.href = nameHref;
    go.target = '_blank';
    go.rel = 'noopener';
    go.title = localWt?.merged_url
      ? `Open merged ${reviewKind}`
      : (b.open_review ? `Open ${reviewKind}` : 'Open branch');
    go.setAttribute('aria-label', go.title);
    go.insertAdjacentHTML('beforeend', ICONS.external);
    actions.appendChild(go);
  }

  item.appendChild(actions);
  return item;
}

function shellQuote(value) {
  const s = String(value || '');
  if (/^[A-Za-z0-9_./:@+-]+$/.test(s)) return s;
  return `'${s.replace(/'/g, `'\\''`)}'`;
}

// Primary checkout: switch off the branch then delete it.
// Linked worktree: remove the worktree path.
function pruneCommand(branchName, localWt, local) {
  const branch = shellQuote(branchName);
  if (localWt?.main) {
    const def = shellQuote(local?.default_branch || 'main');
    return `git switch ${def} && git branch -d ${branch}`;
  }
  if (localWt?.path) {
    return `git worktree remove ${shellQuote(localWt.path)}`;
  }
  return `git branch -d ${branch}`;
}

/** Origin sync for an appearance's current branch (list entry or top-level ahead/behind). */
function appearanceBranchSync(app) {
  if (!app || app.detached || app.error) return null;
  const branch = String(app.branch || '').trim();
  if (!branch) return null;
  const fromList = originSyncFor(app, branch);
  if (fromList) return fromList;
  if (app.ahead || app.behind) {
    return { name: branch, ahead: app.ahead || 0, behind: app.behind || 0 };
  }
  return null;
}

/** True when the appearance checkout of branch is dirty (blocks ff pull). */
function appearancePullDirty(app, branch) {
  const name = String(branch || '').trim();
  if (!name || !app) return true;
  if (app.dirty && !app.detached && String(app.branch || '') === name) return true;
  const trees = Array.isArray(app.worktrees) ? app.worktrees : [];
  return trees.some((w) => !w.bare && String(w.branch || '') === name && w.dirty);
}

/** Display label for an appearance checkout (branch, tag, or detached SHA). */
function appearanceRefLabel(app) {
  if (!app) return '';
  if (app.detached) {
    const tag = String(app.tag || '').trim();
    if (tag) return tag;
    const sha = String(app.branch || '').trim();
    return sha ? `detached ${sha}` : 'detached';
  }
  return String(app.branch || '').trim();
}

function appearanceStatusMarks(app, project) {
  const marks = el('span', 'appearance-marks');
  if (app.error) {
    marks.appendChild(iconMark(ICONS.alert, 'mark--bad', app.error));
    return marks;
  }
  if (app.branch || app.tag) {
    const label = appearanceRefLabel(app);
    const br = el('span', 'appearance-branch', label);
    if (app.detached && app.tag && app.branch) {
      br.title = `detached at ${app.branch}`;
    } else if (app.detached && app.branch) {
      br.title = 'detached HEAD';
    }
    marks.appendChild(br);
  }
  if (app.dirty) {
    marks.appendChild(iconMark(ICONS.laptop, 'mark--warn', 'dirty working tree'));
  }
  const sync = appearanceBranchSync(app);
  const bits = originSyncBitsForPull(sync, project?.id, app.branch, app.path);
  if (bits.length) {
    const span = el('span', 'appearance-sync is-divergent', bits.join(' '));
    if (project?.id) span.id = morphId('app-sync', project.id, app.path || '', app.branch || '');
    span.title = originSyncTitle(sync) || 'versus upstream';
    marks.appendChild(span);
  }
  return marks;
}

function renderAppearanceDetail(app) {
  const detail = el('div', 'appearance-detail');
  detail.appendChild(el('div', 'appearance-detail-path', app.path || ''));
  const trees = (Array.isArray(app.worktrees) ? app.worktrees : []).filter((w) => !w.bare);
  if (!trees.length) {
    if (app.error) detail.appendChild(el('div', 'error', app.error));
    return detail;
  }
  const list = el('ul', 'appearance-worktrees');
  for (const wt of trees) {
    const li = el('li', 'appearance-worktree');
    const label = wt.main ? 'main' : (wt.branch || 'worktree');
    const bits = [label];
    if (wt.branch && !wt.main) bits[0] = wt.branch;
    if (wt.dirty) bits.push('dirty');
    const sync = originSyncFor(app, wt.branch);
    const syncBits = originSyncBits(sync);
    if (syncBits.length) bits.push(...syncBits);
    else {
      if (wt.ahead) bits.push(`↑${wt.ahead}`);
      if (wt.behind) bits.push(`↓${wt.behind}`);
    }
    li.appendChild(el('span', '', bits.join(' · ')));
    if (wt.path && wt.path !== app.path) {
      li.appendChild(el('span', 'appearance-wt-path', wt.path));
    }
    list.appendChild(li);
  }
  detail.appendChild(list);
  return detail;
}

function renderAppearances(local, project) {
  const apps = localAppearances(local);
  if (!apps.length) return null;

  const wrap = el('div', 'project-appearances');
  for (const app of apps) {
    const row = el('div', 'appearance-row' + (app.primary ? ' appearance-row--primary' : ''));
    if (project?.id) {
      row.id = morphId('app', project.id, app.path || app.display_id || '');
      row.dataset.projectId = project.id;
    }
    const line = el('div', 'appearance-line');
    const copyBtn = copyPathButton(app.path);
    if (copyBtn) line.appendChild(copyBtn);
    const head = el('button', 'appearance-head');
    head.type = 'button';
    head.setAttribute('aria-expanded', 'false');
    head.dataset.action = 'toggle-appearance';
    const id = el('span', 'project-local-path appearance-id', app.display_id || app.path || '');
    id.title = app.path || '';
    head.appendChild(id);
    head.appendChild(appearanceStatusMarks(app, project));
    const detail = renderAppearanceDetail(app);
    detail.hidden = true;
    line.appendChild(head);
    const sync = appearanceBranchSync(app);
    const branchName = String(app.branch || '').trim();
    appendPullButton(line, {
      sync,
      project,
      branchName,
      repoPath: app.path,
      dirty: appearancePullDirty(app, branchName),
      whyDirty: app.dirty ? `Pull blocked: ${app.display_id || app.path || 'checkout'} is dirty` : '',
    });
    row.appendChild(line);
    row.appendChild(detail);
    wrap.appendChild(row);
  }
  return wrap;
}

function toolStatusMark(cli) {
  if (!cli?.installed) {
    return iconMark(ICONS.x, 'mark--bad', 'missing');
  }
  if (!cli.authed) {
    return iconMark(ICONS.alert, 'mark--warn', 'login required');
  }
  return iconMark(ICONS.check, 'mark--ok', 'ready');
}

function toolRow(hostKey, label, cli) {
  const row = el('span', `tool-row tool-row--${hostKey}`);
  const state = !cli?.installed
    ? 'missing'
    : (!cli.authed ? 'login required' : 'ready');
  const detail = cli?.detail ? ` (${cli.detail})` : '';
  const tip = `${label}: ${state}${detail}`;
  row.title = tip;
  row.setAttribute('aria-label', tip);
  row.insertAdjacentHTML('beforeend', ICONS[hostKey] || '');
  row.appendChild(toolStatusMark(cli));
  return row;
}

function renderTooling(tool) {
  const root = document.getElementById('tooling');
  root.innerHTML = '';
  root.appendChild(toolRow('github', 'gh', tool?.github || {}));
  root.appendChild(toolRow('gitlab', 'glab', tool?.gitlab || {}));
}

const CI_FAILED = new Set(['failed', 'failure', 'error', 'cancelled', 'canceled']);

let allProjects = [];

/** @type {string[]} Active ui.hide_branches patterns from the last dashboard payload. */
let hideBranchPatterns = [];

const VIEW_STORAGE_KEY = 'gitboard.activeView';
/** @type {string} */
let activeViewId = '';
/** @type {{ id: string, label: string, count: number, implicit?: boolean }[]} */
let boardViews = [];

try {
  activeViewId = String(localStorage.getItem(VIEW_STORAGE_KEY) || '').trim();
} catch (_) {
  activeViewId = '';
}

const filters = {
  q: '',
  /** @type {Set<string>} empty = all host/org scopes */
  scopes: new Set(),
  chips: {
    ci_failed: false,
    open_review: false,
    dirty: false,
    prune: false,
    conflicts: false,
  },
};

function normalizeQ(s) {
  return String(s || '').toLowerCase().trim();
}

function scopeKey(host, org) {
  return `${String(host || '').toLowerCase()}\0${String(org || '').trim()}`;
}

function parseScopeKey(key) {
  const i = String(key).indexOf('\0');
  if (i < 0) return { host: String(key || ''), org: '' };
  return { host: key.slice(0, i), org: key.slice(i + 1) };
}

function hostLabel(host) {
  const key = String(host || '').toLowerCase();
  if (key === 'github') return 'GitHub';
  if (key === 'gitlab') return 'GitLab';
  return key || 'unknown';
}

function collectScopeOptions(projects) {
  /** @type {Map<string, { host: string, org: string }>} */
  const map = new Map();
  for (const row of projects || []) {
    const host = String(row.host || '').toLowerCase();
    if (!host) continue;
    const org = String(row.org || '').trim();
    const key = scopeKey(host, org);
    if (!map.has(key)) map.set(key, { host, org });
  }
  return [...map.values()].sort((a, b) => {
    if (a.host !== b.host) return a.host.localeCompare(b.host);
    return a.org.localeCompare(b.org);
  });
}

/** Ranked search fields: higher weight wins when sorting matches. */
function searchFields(row) {
  const branches = [];
  for (const b of row.branches || []) {
    if (b.name) branches.push(b.name);
  }
  const local = row.local;
  const localBits = [];
  if (local?.path) localBits.push(local.path);
  if (local?.branch) localBits.push(local.branch);
  if (local?.tag) localBits.push(local.tag);
  for (const app of local?.appearances || []) {
    if (app.display_id) localBits.push(app.display_id);
    if (app.path) localBits.push(app.path);
    if (app.branch) localBits.push(app.branch);
    if (app.tag) localBits.push(app.tag);
    if (app.parent_label) localBits.push(app.parent_label);
  }
  for (const wt of local?.worktrees || []) {
    if (wt.branch) localBits.push(wt.branch);
    if (wt.tag) localBits.push(wt.tag);
    if (wt.path) localBits.push(wt.path);
  }
  return {
    label: normalizeQ(row.label),
    id: normalizeQ(row.id),
    path: normalizeQ(row.path),
    org: normalizeQ(row.org),
    host: normalizeQ(row.host),
    branch: normalizeQ(branches.join(' ')),
    local: normalizeQ(localBits.join(' ')),
  };
}

const SEARCH_FIELD_WEIGHTS = {
  label: 1000,
  id: 700,
  path: 400,
  org: 250,
  branch: 120,
  local: 80,
  host: 20,
};

function tokenFieldScore(token, value) {
  if (!token || !value) return 0;
  if (value === token) return 100;
  // Word / path segment prefix (consil → consilium).
  if (value.split(/[\s/_-]+/).some((part) => part.startsWith(token))) return 75;
  if (value.startsWith(token)) return 70;
  if (value.includes(token)) return 35;
  return 0;
}

/** Higher is better. Returns -1 when any token misses. */
function searchScore(row, q) {
  const tokens = normalizeQ(q).split(/\s+/).filter(Boolean);
  if (!tokens.length) return 0;
  const fields = searchFields(row);
  let total = 0;
  for (const tok of tokens) {
    let best = 0;
    for (const [name, weight] of Object.entries(SEARCH_FIELD_WEIGHTS)) {
      const hit = tokenFieldScore(tok, fields[name]);
      if (!hit) continue;
      const score = hit * weight;
      if (score > best) best = score;
    }
    if (!best) return -1;
    total += best;
  }
  return total;
}

function projectDirty(row) {
  const local = row.local;
  if (!local?.mapped) return false;
  if (local.dirty) return true;
  for (const app of local.appearances || []) {
    if (app.dirty) return true;
    for (const wt of app.worktrees || []) {
      if (wt.dirty) return true;
    }
  }
  for (const wt of local.worktrees || []) {
    if (wt.dirty) return true;
  }
  return false;
}

function projectPrune(row) {
  const local = row.local;
  if (!local?.mapped) return false;
  const trees = [];
  for (const wt of local.worktrees || []) trees.push(wt);
  for (const app of local.appearances || []) {
    for (const wt of app.worktrees || []) trees.push(wt);
  }
  return trees.some((wt) => wt.prune_hint === 'safe' || wt.prune_hint === 'likely');
}

function projectCIFailed(row) {
  return (row.branches || []).some((b) => CI_FAILED.has(String(b.ci_status || '').toLowerCase()));
}

function projectOpenReview(row) {
  // Only branches still shown after ui.hide_branches; ignore forge open_items
  // totals that may still count hidden PR/MR heads.
  return (row.branches || []).some((b) => b.open_review);
}

function projectConflicts(row) {
  return (row.branches || []).some((b) => b.conflict);
}

/** True when any mapped checkout is behind-only vs origin (ff pull candidate). */
function projectBehind(row) {
  const local = row.local;
  if (!local?.mapped) return false;
  const syncBehind = (sync) => Boolean(sync?.behind) && !sync?.ahead;
  // Detached pins (common for submodules) often have origin_sync rows for
  // unrelated local branches; only count behind on a real pull candidate.
  if (!local.detached) {
    for (const sync of local.origin_sync || []) {
      if (syncBehind(sync)) return true;
    }
    if (syncBehind({ ahead: local.ahead, behind: local.behind })) return true;
  }
  for (const app of local.appearances || []) {
    const sync = appearanceBranchSync(app);
    if (syncBehind(sync)) return true;
    for (const wt of app.worktrees || []) {
      if (wt.bare || wt.detached) continue;
      if (syncBehind({ ahead: wt.ahead, behind: wt.behind })) return true;
    }
  }
  for (const wt of local.worktrees || []) {
    if (wt.bare || wt.detached) continue;
    if (syncBehind({ ahead: wt.ahead, behind: wt.behind })) return true;
  }
  return false;
}

/**
 * Higher means more board attention. Weights mirror filter chips:
 * failed CI > conflicts > open review > dirty > prune > behind.
 */
function attentionRank(row) {
  let score = 0;
  if (projectCIFailed(row)) score += 1000;
  if (projectConflicts(row)) score += 500;
  if (projectOpenReview(row)) score += 200;
  if (projectDirty(row)) score += 100;
  if (projectPrune(row)) score += 50;
  if (projectBehind(row)) score += 25;
  return score;
}

function projectSortName(row) {
  return String(row.label || row.id || row.path || '').toLowerCase();
}

function projectMatches(row, state) {
  if (state.scopes.size) {
    const key = scopeKey(row.host, row.org);
    if (!state.scopes.has(key)) return false;
  }
  if (state.chips.ci_failed && !projectCIFailed(row)) return false;
  if (state.chips.open_review && !projectOpenReview(row)) return false;
  if (state.chips.dirty && !projectDirty(row)) return false;
  if (state.chips.prune && !projectPrune(row)) return false;
  if (state.chips.conflicts && !projectConflicts(row)) return false;
  if (state.q) {
    if (searchScore(row, state.q) < 0) return false;
  }
  return true;
}

function filtersActive(state) {
  if (state.q || state.scopes.size) return true;
  return Object.values(state.chips).some(Boolean);
}

function filteredProjects() {
  const matched = (allProjects || []).filter((row) => projectMatches(row, filters));
  return matched
    .map((row, index) => ({
      row,
      index,
      search: filters.q ? searchScore(row, filters.q) : 0,
      attention: attentionRank(row),
      name: projectSortName(row),
    }))
    .sort((a, b) => {
      if (filters.q && b.search !== a.search) return b.search - a.search;
      if (b.attention !== a.attention) return b.attention - a.attention;
      if (a.name !== b.name) return a.name < b.name ? -1 : 1;
      return a.index - b.index;
    })
    .map((item) => item.row);
}

function updateFilterChrome(visibleCount) {
  const total = allProjects.length;
  const countEl = document.getElementById('filter-count');
  const clearBtn = document.getElementById('filter-clear');
  const active = filtersActive(filters);
  if (clearBtn) clearBtn.hidden = !active;
  if (!countEl) return;
  if (!total) {
    countEl.textContent = '';
    return;
  }
  if (!active) {
    countEl.textContent = `${total} project${total === 1 ? '' : 's'}`;
    return;
  }
  countEl.textContent = `Showing ${visibleCount} of ${total}`;
}

function updateScopeLabel() {
  const label = document.getElementById('filter-scope-label');
  if (!label) return;
  const n = filters.scopes.size;
  if (!n) {
    label.textContent = 'All hosts / orgs';
    return;
  }
  if (n === 1) {
    const { host, org } = parseScopeKey([...filters.scopes][0]);
    label.textContent = org ? `${hostLabel(host)} · ${org}` : hostLabel(host);
    return;
  }
  label.textContent = `${n} hosts / orgs`;
}

function rebuildScopeMenu() {
  const menu = document.getElementById('filter-scope-menu');
  if (!menu) return;
  const options = collectScopeOptions(allProjects);
  const keep = new Set();
  for (const opt of options) keep.add(scopeKey(opt.host, opt.org));
  for (const key of [...filters.scopes]) {
    if (!keep.has(key)) filters.scopes.delete(key);
  }
  menu.innerHTML = '';
  if (!options.length) {
    menu.appendChild(el('div', 'filter-scope-empty', 'No hosts / orgs yet'));
    updateScopeLabel();
    return;
  }
  for (const opt of options) {
    const key = scopeKey(opt.host, opt.org);
    const btn = el('button', 'filter-scope-option');
    btn.type = 'button';
    btn.setAttribute('role', 'option');
    btn.dataset.scope = key;
    const selected = filters.scopes.has(key);
    btn.setAttribute('aria-selected', selected ? 'true' : 'false');
    btn.appendChild(hostPrefix(opt.host, opt.org || opt.host));
    const name = el('span', 'filter-scope-option-name', opt.org || '(no org)');
    btn.appendChild(name);
    const check = el('span', 'filter-scope-check');
    check.setAttribute('aria-hidden', 'true');
    check.textContent = '✓';
    btn.appendChild(check);
    btn.addEventListener('click', (ev) => {
      ev.preventDefault();
      ev.stopPropagation();
      if (filters.scopes.has(key)) filters.scopes.delete(key);
      else filters.scopes.add(key);
      btn.setAttribute('aria-selected', filters.scopes.has(key) ? 'true' : 'false');
      updateScopeLabel();
      applyBoard();
    });
    menu.appendChild(btn);
  }
  updateScopeLabel();
}

function setScopeMenuOpen(open) {
  const btn = document.getElementById('filter-scope-btn');
  const menu = document.getElementById('filter-scope-menu');
  if (!btn || !menu) return;
  btn.setAttribute('aria-expanded', open ? 'true' : 'false');
  menu.hidden = !open;
}

function setConfigInfoOpen(open) {
  const btn = document.getElementById('config-info-btn');
  const menu = document.getElementById('config-info-menu');
  if (!btn || !menu) return;
  btn.setAttribute('aria-expanded', open ? 'true' : 'false');
  menu.hidden = !open;
}

/** @param {{ hide_branches?: string[] } | null | undefined} ui */
function setHideBranchPatterns(ui) {
  hideBranchPatterns = Array.isArray(ui?.hide_branches)
    ? ui.hide_branches.map((p) => String(p || '').trim()).filter(Boolean)
    : [];
}

/** @param {{ hide_branches?: string[] } | null | undefined} ui */
function renderConfigInfo(ui) {
  setHideBranchPatterns(ui);
  const btn = document.getElementById('config-info-btn');
  const menu = document.getElementById('config-info-menu');
  if (!btn || !menu) return;
  const patterns = hideBranchPatterns;
  btn.classList.toggle('has-patterns', patterns.length > 0);
  btn.title = patterns.length
    ? `${patterns.length} hide_branches pattern${patterns.length === 1 ? '' : 's'} active`
    : 'Board config from your config file';
  if (!btn.querySelector('.mark-icon')) {
    btn.insertAdjacentHTML('afterbegin', ICONS.info);
  }
  const section = el('div', 'config-info-section');
  section.appendChild(el('p', 'config-info-key', 'ui.hide_branches'));
  if (patterns.length === 0) {
    section.appendChild(el('p', 'config-info-empty', 'None (showing all branches)'));
  } else {
    const list = el('ul', 'config-info-list');
    for (const p of patterns) {
      const li = document.createElement('li');
      const code = document.createElement('code');
      code.textContent = p;
      li.appendChild(code);
      list.appendChild(li);
    }
    section.appendChild(list);
  }
  menu.replaceChildren(el('p', 'config-info-title', 'Board config'), section);
}

/**
 * Go path.Match subset used by ui.hide_branches: * and ? do not cross '/'.
 * Invalid patterns never match (same as the server).
 * @param {string} pattern
 * @param {string} name
 */
function pathMatch(pattern, name) {
  const pat = String(pattern || '');
  const text = String(name || '');
  if (!pat) return false;
  let re = '^';
  for (let i = 0; i < pat.length; i++) {
    const c = pat[i];
    if (c === '*') {
      re += '[^/]*';
      continue;
    }
    if (c === '?') {
      re += '[^/]';
      continue;
    }
    if (c === '\\') {
      i++;
      if (i >= pat.length) return false;
      re += pat[i].replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      continue;
    }
    if ('^$+(){}|[]'.includes(c) || c === '.') {
      re += `\\${c}`;
      continue;
    }
    re += c;
  }
  re += '$';
  try {
    return new RegExp(re).test(text);
  } catch {
    return false;
  }
}

/**
 * @param {string} name
 * @param {string[]} patterns
 */
function branchNameHidden(name, patterns) {
  const n = String(name || '');
  if (!n || !Array.isArray(patterns) || patterns.length === 0) return false;
  for (const pat of patterns) {
    const p = String(pat || '').trim();
    if (!p) continue;
    if (pathMatch(p, n)) return true;
  }
  return false;
}

function applyBoard() {
  const rows = filteredProjects();
  renderRows(rows);
  updateFilterChrome(rows.length);
  updateScopeLabel();
}

function readFiltersFromDom() {
  const qInput = document.getElementById('filter-q');
  filters.q = normalizeQ(qInput?.value);
  for (const btn of document.querySelectorAll('.filter-chip[data-filter]')) {
    const key = btn.getAttribute('data-filter');
    if (key && key in filters.chips) {
      filters.chips[key] = btn.getAttribute('aria-pressed') === 'true';
    }
  }
}

function clearFilters() {
  const qInput = document.getElementById('filter-q');
  if (qInput) qInput.value = '';
  for (const btn of document.querySelectorAll('.filter-chip[data-filter]')) {
    btn.setAttribute('aria-pressed', 'false');
  }
  filters.q = '';
  filters.scopes.clear();
  for (const key of Object.keys(filters.chips)) filters.chips[key] = false;
  setScopeMenuOpen(false);
  applyBoard();
}

function bindFilters() {
  const qInput = document.getElementById('filter-q');
  const clearBtn = document.getElementById('filter-clear');
  const scopeBtn = document.getElementById('filter-scope-btn');
  const scopeRoot = document.getElementById('filter-scope');
  const configBtn = document.getElementById('config-info-btn');
  const configRoot = document.getElementById('config-info');
  const onChange = () => {
    readFiltersFromDom();
    applyBoard();
  };
  qInput?.addEventListener('input', onChange);
  clearBtn?.addEventListener('click', () => clearFilters());
  scopeBtn?.addEventListener('click', (ev) => {
    ev.preventDefault();
    ev.stopPropagation();
    const open = scopeBtn.getAttribute('aria-expanded') === 'true';
    setConfigInfoOpen(false);
    setScopeMenuOpen(!open);
  });
  configBtn?.addEventListener('click', (ev) => {
    ev.preventDefault();
    ev.stopPropagation();
    const open = configBtn.getAttribute('aria-expanded') === 'true';
    setScopeMenuOpen(false);
    setConfigInfoOpen(!open);
  });
  document.addEventListener('click', (ev) => {
    if (scopeRoot && !scopeRoot.contains(ev.target)) setScopeMenuOpen(false);
    if (configRoot && !configRoot.contains(ev.target)) setConfigInfoOpen(false);
  });
  document.addEventListener('keydown', (ev) => {
    if (ev.key === 'Escape') {
      setScopeMenuOpen(false);
      setConfigInfoOpen(false);
    }
  });
  for (const btn of document.querySelectorAll('.filter-chip[data-filter]')) {
    btn.addEventListener('click', () => {
      const wasOn = btn.getAttribute('aria-pressed') === 'true';
      for (const other of document.querySelectorAll('.filter-chip[data-filter]')) {
        other.setAttribute('aria-pressed', 'false');
      }
      if (!wasOn) btn.setAttribute('aria-pressed', 'true');
      onChange();
    });
  }
  renderConfigInfo({ hide_branches: [] });
}

function renderRows(projects) {
  const tbody = document.getElementById('rows');
  // Build rows under a detached tbody, then morph its *children* into #rows.
  // Morphing a second tbody (especially with id="rows") makes Idiomorph try to
  // insert that element under the live #rows and throws HierarchyRequestError.
  // Sync/pull nodes use morphId so ephemeral ↓N / pull busy state updates in place
  // instead of leaving stale marks from a prior paint.
  const scratch = document.createElement('tbody');
  for (const row of projects || []) {
    const tr = el('tr');
    if (row.id) {
      tr.id = morphId('proj', row.id);
      tr.dataset.projectId = row.id;
    }
    const titleCell = el('td', 'project-cell');
    if (row.org) {
      const orgLine = el('div', 'project-org');
      orgLine.appendChild(hostPrefix(row.host, row.path));
      orgLine.appendChild(el('span', 'project-org-name', row.org));
      titleCell.appendChild(orgLine);
    }
    const titleRow = el('div', 'project-title');
    if (!row.org) {
      titleRow.appendChild(hostPrefix(row.host, row.path));
    }
    titleRow.appendChild(el('strong', '', row.label || row.id));
    if (row.open_url) {
      const href = safeHref(row.open_url);
      if (href) {
        const link = el('a', 'project-open');
        link.href = href;
        link.target = '_blank';
        link.rel = 'noopener';
        link.title = 'Open repository';
        link.setAttribute('aria-label', `Open ${row.label || row.id}`);
        link.insertAdjacentHTML('beforeend', ICONS.external);
        titleRow.appendChild(link);
      }
    }
    titleCell.appendChild(titleRow);
    const local = row.local;
    const appearanceBlock = renderAppearances(local, row);
    if (appearanceBlock) {
      titleCell.appendChild(appearanceBlock);
    } else if (local && local.mapped === false) {
      titleCell.appendChild(el('div', 'project-local-path', 'local not mapped'));
    } else if (local?.error) {
      titleCell.appendChild(el('div', 'error', local.error));
    }
    if (row.error) {
      titleCell.appendChild(el('div', 'error', row.error));
    }
    tr.appendChild(titleCell);
    tr.appendChild(branchesCell(row.branches, row.host, row));
    scratch.appendChild(tr);
  }
  Idiomorph.morph(tbody, Array.from(scratch.children), {
    morphStyle: 'innerHTML',
    callbacks: {
      beforeNodeMorphed(oldNode, newNode) {
        if (!(oldNode instanceof HTMLElement) || !(newNode instanceof HTMLElement)) return;
        if (!oldNode.classList.contains('appearance-row') || !oldNode.classList.contains('is-open')) return;
        newNode.classList.add('is-open');
        const head = newNode.querySelector(':scope > .appearance-line > .appearance-head');
        const detail = newNode.querySelector(':scope > .appearance-detail');
        if (head) head.setAttribute('aria-expanded', 'true');
        if (detail) detail.hidden = false;
      },
    },
  });
}

let pollSeconds = 30;
let pollTimer = null;
let loading = false;
/** Monotonic id for in-flight dashboard fetches; only the latest may paint. */
let dashboardGen = 0;
/** @type {AbortController | null} */
let dashboardAbort = null;
/** Last successful dashboard payload per view id (instant paint on switch). */
const viewPayloadCache = new Map();
/** View id whose tab shows a busy spinner while its dashboard fetch is in flight. */
let loadingViewId = '';

function updatePollLabel() {
  const label = document.getElementById('poll-label');
  if (!label) return;
  if (pollSeconds <= 0) {
    label.textContent = 'auto-refresh off';
    return;
  }
  label.textContent = `every ${pollSeconds}s`;
}

function schedulePoll() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
  if (pollSeconds <= 0) return;
  pollTimer = setInterval(() => {
    if (document.hidden || loading || pendingPulls.size > 0 || pullFlushScheduled) return;
    void loadDashboard({ quiet: true, fresh: pullBoardDirty });
  }, pollSeconds * 1000);
}

function dashboardURL(fresh) {
  const params = new URLSearchParams();
  if (activeViewId) params.set('view', activeViewId);
  if (fresh) params.set('fresh', '1');
  const q = params.toString();
  return q ? `/api/dashboard?${q}` : '/api/dashboard';
}

function persistActiveView(id) {
  activeViewId = String(id || '').trim();
  try {
    if (activeViewId) localStorage.setItem(VIEW_STORAGE_KEY, activeViewId);
    else localStorage.removeItem(VIEW_STORAGE_KEY);
  } catch (_) {
    /* ignore */
  }
}

/**
 * Paint board UI from a dashboard JSON payload (live or cached).
 * @param {object} data
 * @param {{ fresh?: boolean, fromCache?: boolean }} [opts]
 */
function paintDashboardData(data, { fresh = false, fromCache = false } = {}) {
  const status = document.getElementById('status');
  if (typeof data.poll_interval_seconds === 'number') {
    const next = data.poll_interval_seconds;
    if (next !== pollSeconds) {
      pollSeconds = next;
      updatePollLabel();
      schedulePoll();
    }
  }
  renderConfigInfo(data.ui);
  if (Array.isArray(data.views)) {
    const active = String(data.active_view || '').trim();
    if (!fromCache && active && active !== activeViewId) persistActiveView(active);
    renderViewSwitcher(data.views, activeViewId || active);
  }
  if (fresh) {
    recentPulled.clear();
    pullBoardDirty = false;
  } else if (pullBoardDirty) {
    schedulePullFlush();
  } else if (!fromCache) {
    recentPulled.clear();
  }
  renderTooling(data.tooling);
  allProjects = Array.isArray(data.projects) ? data.projects : [];
  rebuildScopeMenu();
  applyBoard();
  if (status) {
    if (fromCache) {
      status.textContent = `Cached ${data.generated_at || ''}…`;
    } else {
      status.textContent = `Updated ${data.generated_at || ''}`;
    }
  }
  const cacheKey = String(data.active_view || activeViewId || '').trim();
  if (cacheKey && !fromCache) {
    viewPayloadCache.set(cacheKey, data);
  }
}

function renderViewSwitcher(views, active) {
  const root = document.getElementById('view-switcher');
  if (!root) return;
  boardViews = Array.isArray(views) ? views : [];
  root.replaceChildren();
  if (boardViews.length === 0) return;
  for (const v of boardViews) {
    const btn = el('button', 'view-tab');
    btn.type = 'button';
    btn.setAttribute('role', 'tab');
    const label = v.label || v.id;
    const selected = v.id === active;
    btn.setAttribute('aria-selected', selected ? 'true' : 'false');
    btn.title = `${label} (${v.count ?? 0})`;
    if (loadingViewId && v.id === loadingViewId) {
      setButtonBusy(btn, label, `Loading ${label}`);
      btn.setAttribute('aria-selected', 'true');
    } else {
      btn.textContent = label;
    }
    btn.addEventListener('click', () => {
      if (v.id === activeViewId) return;
      persistActiveView(v.id);
      loadingViewId = v.id;
      const cached = viewPayloadCache.get(v.id);
      if (cached) {
        paintDashboardData(cached, { fromCache: true });
        void loadDashboard({ quiet: false, fresh: false, viewBusy: true });
        return;
      }
      renderViewSwitcher(boardViews, v.id);
      void loadDashboard({ quiet: false, fresh: false, viewBusy: true });
    });
    root.appendChild(btn);
  }
}

/**
 * Fetch and paint the board. Returns true when this generation painted successfully.
 * @param {{ quiet?: boolean, fresh?: boolean, viewBusy?: boolean }} [opts]
 */
async function loadDashboard({ quiet = false, fresh = false, viewBusy = false } = {}) {
  // Do not let a quiet poll abort a post-pull / Refresh fresh load mid-flight.
  if (!fresh && dashboardFreshInFlight) {
    return false;
  }
  if (!viewBusy && loadingViewId) {
    loadingViewId = '';
  }
  const gen = ++dashboardGen;
  if (dashboardAbort) {
    dashboardAbort.abort();
  }
  const ac = new AbortController();
  dashboardAbort = ac;
  loading = true;
  const trackingFresh = fresh;
  if (trackingFresh) dashboardFreshInFlight = true;
  const status = document.getElementById('status');
  if (viewBusy && loadingViewId && boardViews.length) {
    renderViewSwitcher(boardViews, loadingViewId);
  }
  if (!quiet) {
    status.textContent = loadingViewId ? 'Loading…' : 'Refreshing…';
  }
  try {
    const url = dashboardURL(fresh);
    const res = await fetch(url, { cache: 'no-store', signal: ac.signal });
    if (gen !== dashboardGen) return false;
    if (!res.ok) {
      if (res.status === 400 && activeViewId) {
        persistActiveView('');
        loadingViewId = '';
        if (gen === dashboardGen) {
          loading = false;
          if (dashboardAbort === ac) dashboardAbort = null;
          if (trackingFresh) dashboardFreshInFlight = false;
        }
        return loadDashboard({ quiet, fresh });
      }
      throw new Error(`HTTP ${res.status}`);
    }
    const data = await res.json();
    if (gen !== dashboardGen) return false;
    loadingViewId = '';
    paintDashboardData(data, { fresh });
    return true;
  } catch (err) {
    if (gen !== dashboardGen) return false;
    if (err && typeof err === 'object' && /** @type {{ name?: string }} */ (err).name === 'AbortError') {
      return false;
    }
    if (loadingViewId) {
      loadingViewId = '';
      renderViewSwitcher(boardViews, activeViewId);
    }
    status.textContent = `Error: ${err instanceof Error ? err.message : String(err)}`;
    return false;
  } finally {
    if (gen === dashboardGen) {
      loading = false;
      if (dashboardAbort === ac) dashboardAbort = null;
      if (trackingFresh) dashboardFreshInFlight = false;
    }
  }
}

async function pruneSafeCheckout({ project_id, branch, worktree_path, button }) {
  const status = document.getElementById('status');
  if (!project_id || !branch || !worktree_path) {
    if (status) status.textContent = 'Missing project, branch, or worktree path';
    return;
  }
  const label = `${project_id} / ${branch}`;
  const ok = await confirmDialog({
    title: 'Remove local checkout?',
    body: `Remove the local branch checkout for ${label}. This only affects your machine.`,
    detail: worktree_path,
    confirmLabel: 'Remove',
    cancelLabel: 'Cancel',
    danger: true,
  });
  if (!ok) return;
  setButtonBusy(button, 'removing…');
  if (status) status.textContent = `Removing ${label}…`;
  try {
    await postJSONWithIndexLockConfirm('/api/prune/safe', {
      project_id,
      branch,
      worktree_path,
    }, {
      title: `${label}: stale git lock`,
      body: 'A leftover git index.lock is blocking remove. Remove the lock and continue?',
    });
    if (status) status.textContent = `Removed ${label}`;
    await loadDashboard({ quiet: true, fresh: true });
  } catch (err) {
    if (err && err.name === 'AbortError') {
      setButtonIdle(button, {
        svg: ICONS.trash,
        label: 'safe to remove',
      });
      if (status) status.textContent = '';
      return;
    }
    const msg = err instanceof Error ? err.message : String(err);
    if (status) status.textContent = '';
    setButtonIdle(button, {
      svg: ICONS.trash,
      label: 'safe to remove',
    });
    await infoDialog({
      title: `${label}: remove failed`,
      body: msg,
      detail: worktree_path,
    });
  }
}

async function pullFFCheckout({ project_id, branch, repo_path, behind, button }) {
  const status = document.getElementById('status');
  if (!project_id || !branch || !repo_path) {
    if (status) status.textContent = 'Missing project, branch, or repo path';
    return;
  }
  const label = `${project_id} / ${branch}`;
  const key = pullActionKey(project_id, branch, repo_path);
  if (pendingPulls.has(key)) return;

  pendingPulls.add(key);
  // Repaint so the static ↓N hides while the button shows pulling…
  applyBoard();
  if (status) status.textContent = pullStatusText();

  try {
    await postJSONWithIndexLockConfirm('/api/pull/ff', {
      project_id,
      branch,
      repo_path,
    }, {
      title: `${label}: stale git lock`,
      body: 'A leftover git index.lock is blocking pull. Remove the lock and continue?',
    });
    recentPulled.add(key);
    pullBoardDirty = true;
  } catch (err) {
    if (err && err.name === 'AbortError') {
      // User declined lock clear; treat as cancel, not a pull failure.
    } else {
      const msg = err instanceof Error ? err.message : String(err);
      pullFailures.push({ label, path: repo_path, message: msg });
    }
  } finally {
    pendingPulls.delete(key);
    // Pending cleared: show "pulled" (or restore ↓) before the fresh fetch returns.
    applyBoard();
    if (pendingPulls.size > 0) {
      if (status) status.textContent = pullStatusText();
    } else {
      if (status && recentPulled.size > 0) {
        status.textContent = 'Pulls finished; refreshing…';
      }
      schedulePullFlush();
    }
  }
}

/**
 * POST JSON; on 409 stale_index_lock ask to clear and retry once with clear_index_lock.
 * @param {string} url
 * @param {Record<string, unknown>} body
 * @param {{ title: string, body: string }} lockPrompt
 */
async function postJSONWithIndexLockConfirm(url, body, lockPrompt) {
  const send = async (payload) => {
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const text = await res.text();
    let data = null;
    if (text) {
      try {
        data = JSON.parse(text);
      } catch {
        data = null;
      }
    }
    return { res, text, data };
  };

  let { res, text, data } = await send(body);
  if (res.status === 409 && data && data.code === 'stale_index_lock') {
    const age = data.age_seconds ? ` (about ${data.age_seconds}s old)` : '';
    const ok = await confirmDialog({
      title: lockPrompt.title,
      body: `${lockPrompt.body}${age}`,
      detail: data.lock_path || data.repo_path || '',
      confirmLabel: 'Remove lock',
      cancelLabel: 'Cancel',
      danger: true,
    });
    if (!ok) {
      const cancel = new Error('cancelled');
      cancel.name = 'AbortError';
      throw cancel;
    }
    ({ res, text, data } = await send({ ...body, clear_index_lock: true }));
  }
  if (!res.ok) {
    const msg = (data && data.message) || text.trim() || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return data;
}

function syncRelationLabel(relation) {
  switch (relation) {
    case 'up_to_date':
      return 'Up to date';
    case 'behind_only':
      return 'Behind only';
    case 'ahead_only':
      return 'Ahead only';
    case 'diverged':
      return 'Diverged';
    case 'unrelated':
      return 'Unrelated histories';
    default:
      return relation || 'Unknown';
  }
}

function syncRelationMessage(result) {
  if (result.dirty) {
    if (result.relation === 'behind_only') {
      return 'Origin is ahead, but the working tree is dirty. Clean or stash local changes before fast-forward.';
    }
    return 'The working tree has local changes. Gitboard will not change this checkout automatically.';
  }
  switch (result.relation) {
    case 'up_to_date':
      return 'The local branch and origin point to the same history.';
    case 'behind_only':
      return 'Origin has commits that the local branch does not have. A fast-forward is safe.';
    case 'ahead_only':
      return 'The local branch has commits that are not on origin. Review them before publishing or discarding them.';
    case 'diverged':
      return 'Both sides have unique commits. Choose whether to preserve, rebase, merge, or discard local work before changing this checkout.';
    case 'unrelated':
      return 'The local branch and origin do not share a history. Automatic resolution is blocked.';
    default:
      return 'Git could not classify this branch safely.';
  }
}

function appendSyncCommitGroup(root, label, count, commits) {
  if (!count && (!commits || !commits.length)) return;
  const group = el('section', 'sync-commit-group');
  const heading = el('h3', 'sync-commit-heading', `${label} (${count})`);
  group.appendChild(heading);
  const list = el('ul', 'sync-commit-list');
  for (const commit of commits || []) {
    const item = el('li', 'sync-commit');
    item.appendChild(el('code', 'sync-commit-sha', commit.sha || ''));
    item.appendChild(el('span', '', commit.subject || '(no subject)'));
    list.appendChild(item);
  }
  if (count > (commits || []).length) {
    list.appendChild(el('li', 'sync-commit-more', `${count - commits.length} more not shown`));
  }
  group.appendChild(list);
  root.appendChild(group);
}

function syncInspectionNode(result, request, project) {
  const root = el('div', 'sync-investigation');
  const summary = el('p', 'sync-summary', syncRelationMessage(result));
  root.appendChild(summary);

  const counts = el('div', 'sync-counts');
  counts.appendChild(el('span', 'sync-count', `State: ${syncRelationLabel(result.relation)}`));
  counts.appendChild(el('span', 'sync-count', `Ahead ↑${result.ahead_count || 0}`));
  counts.appendChild(el('span', 'sync-count', `Behind ↓${result.behind_count || 0}`));
  root.appendChild(counts);

  appendSyncCommitGroup(root, 'Local-only commits', result.ahead_count || 0, result.ahead);
  appendSyncCommitGroup(root, 'Origin-only commits', result.behind_count || 0, result.behind);

  const behindOnly = result.relation === 'behind_only' && (result.behind_count || 0) > 0;
  if (behindOnly) {
    const actions = el('div', 'sync-actions');
    const pull = el('button', 'modal-btn modal-btn--ok sync-action');
    pull.type = 'button';
    setButtonLabel(pull, ICONS.pull, `Fast-forward ↓${result.behind_count}`);
    if (result.dirty || !result.can_fast_forward) {
      pull.disabled = true;
      pull.classList.add('is-blocked');
      pull.title = result.dirty
        ? 'Pull blocked: working tree has local changes'
        : 'Pull blocked: fast-forward is not available';
    } else {
      pull.title = `Fast-forward local ${request.branch} from origin`;
      pull.addEventListener('click', (ev) => {
        ev.preventDefault();
        ev.stopPropagation();
        void pullFFCheckout({
          project_id: project.id,
          branch: request.branch,
          repo_path: request.repo_path,
          behind: result.behind_count,
          button: pull,
        });
      });
    }
    actions.appendChild(pull);
    root.appendChild(actions);
  }
  return root;
}

async function investigateSync(request, btn) {
  const status = document.getElementById('status');
  setButtonBusy(btn, 'loading…', `Investigating local ${request.branch} versus origin`);
  if (status) status.textContent = `Investigating ${request.project_id} / ${request.branch}…`;
  const closed = infoDialog({
    title: `${request.project_id} / ${request.branch}: local sync`,
    body: 'Refreshing origin and comparing commits…',
    wide: true,
  });
  try {
    const res = await fetch('/api/local/sync/investigate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });
    const text = await res.text();
    if (!modalIsOpen()) {
      await closed;
      return;
    }
    if (!res.ok) {
      setModalBodyContent(text.trim() || `HTTP ${res.status}`);
      setModalDetailText(request.repo_path);
      await closed;
      return;
    }
    const data = JSON.parse(text);
    const result = data.result || {};
    setModalBodyContent(syncInspectionNode(result, request, { id: request.project_id }));
    const detailLines = [
      result.path && `Path: ${result.path}`,
      result.current_branch && `Checked out: ${result.current_branch}`,
      result.upstream && `Upstream: ${result.upstream}`,
      result.local_sha && `Local: ${result.local_sha}`,
      result.remote_sha && `Origin: ${result.remote_sha}`,
      result.merge_base && `Merge base: ${result.merge_base}`,
      result.dirty_files?.length && `Changed files:\n${result.dirty_files.join('\n')}`,
    ].filter(Boolean);
    setModalDetailText(detailLines.join('\n\n'));
    if (status) status.textContent = '';
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    if (modalIsOpen()) {
      setModalBodyContent(msg);
      setModalDetailText(request.repo_path);
    }
    if (status) status.textContent = '';
  } finally {
    setButtonIdle(btn, {
      svg: ICONS.search,
      label: 'inspect',
      title: `Investigate local ${request.branch} versus origin`,
    });
  }
  await closed;
}

async function investigatePrune(body, btn) {
  const status = document.getElementById('status');
  setButtonBusy(btn, 'investigating…');
  if (status) status.textContent = `Investigating ${body.project_id} / ${body.branch}…`;
  try {
    const res = await fetch('/api/agents/prune/investigate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const text = await res.text();
    if (!res.ok) {
      if (status) status.textContent = '';
      await infoDialog({
        title: `${body.project_id} / ${body.branch}: investigate failed`,
        body: text.trim() || `HTTP ${res.status}`,
      });
      return;
    }
    const data = JSON.parse(text);
    const card = data.card || {};
    const ev = data.evidence || {};
    const detailLines = [
      card.bullets?.length && `Evidence:\n${card.bullets.map((b) => `• ${b}`).join('\n')}`,
      card.command && `Proposed: ${card.command}`,
      data.source && `Source: ${data.source}${data.model ? ` (${data.model})` : ''}`,
      data.session_id && `Session: ${data.session_id}`,
      ev.unique_commit_count != null && `Unique commits: ${ev.unique_commit_count}`,
      ev.unique_file_count != null && `Unique files: ${ev.unique_file_count}`,
    ].filter(Boolean);
    if (status) status.textContent = '';
    await infoDialog({
      title: `${body.project_id} / ${body.branch}: investigate`,
      body: `Verdict: ${card.verdict || '?'}${card.summary ? `\n\n${card.summary}` : ''}`,
      detail: detailLines.join('\n\n'),
    });
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    if (status) status.textContent = '';
    await infoDialog({
      title: `${body.project_id} / ${body.branch}: investigate failed`,
      body: msg,
    });
  } finally {
    setButtonIdle(btn, {
      svg: ICONS.search,
      label: 'Investigate',
      title: 'Gather evidence into an agent session',
    });
  }
}

async function showFailures(project, branch) {
  const branchName = branch?.name ? ` / ${branch.name}` : '';
  const title = `${project.label}${branchName}: failed jobs`;
  const runID = branch?.run_id || project.ci?.run_id || '';
  const closed = infoDialog({ title, body: 'Loading jobs…' });

  try {
    const params = new URLSearchParams({ project: project.id, run_id: runID });
    const res = await fetch(`/api/failures?${params}`);
    if (!modalIsOpen()) {
      await closed;
      return;
    }
    if (!res.ok) {
      setModalBodyContent(await res.text());
      await closed;
      return;
    }
    const data = await res.json();
    if (!modalIsOpen()) {
      await closed;
      return;
    }
    const jobs = data.jobs || [];
    if (!jobs.length) {
      setModalBodyContent('No failed jobs returned. Try opening the pipeline link.');
      await closed;
      return;
    }
    const jobsRoot = el('div', 'jobs');
    for (const job of jobs) {
      const row = el('div', 'job-row');
      row.appendChild(el('span', '', job.name || job.id));
      if (job.web_url) {
        const href = safeHref(job.web_url);
        if (href) {
          const a = el('a', '', 'log');
          a.href = href;
          a.target = '_blank';
          a.rel = 'noopener';
          row.appendChild(a);
        }
      }
      const ai = el('button', '');
      ai.type = 'button';
      setButtonLabel(ai, ICONS.search, 'AI triage');
      ai.addEventListener('click', () => {
        void runTriage(project, job, runID, ai);
      });
      row.appendChild(ai);
      jobsRoot.appendChild(row);
    }
    setModalBodyContent(jobsRoot);
  } catch (err) {
    if (modalIsOpen()) {
      const msg = err instanceof Error ? err.message : String(err);
      setModalBodyContent(msg);
    }
  }
  await closed;
}

async function runTriage(project, job, runID, button) {
  if (!modalIsOpen()) return;
  setModalDetailText('Running triage…');
  setButtonBusy(button, 'triaging…');
  try {
    const res = await fetch('/api/triage', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        project_id: project.id,
        run_id: runID || project.ci?.run_id || '',
        job_id: job.id,
        job_name: job.name,
      }),
    });
    const text = await res.text();
    if (!modalIsOpen()) return;
    if (!res.ok) {
      setModalDetailText(text);
      return;
    }
    const data = JSON.parse(text);
    if (data.unavailable) {
      setModalDetailText(data.unavailable);
      return;
    }
    const lines = [
      data.summary && `Summary: ${data.summary}`,
      data.root_cause && `Root cause: ${data.root_cause}`,
      data.fix_steps?.length && `Fix steps:\n${data.fix_steps.map((s, i) => `${i + 1}. ${s}`).join('\n')}`,
      data.confidence && `Confidence: ${data.confidence}`,
      data.model && `Model: ${data.model}`,
    ].filter(Boolean);
    setModalDetailText(lines.join('\n\n'));
  } catch (err) {
    if (modalIsOpen()) {
      const msg = err instanceof Error ? err.message : String(err);
      setModalDetailText(msg);
    }
  } finally {
    setButtonIdle(button, { svg: ICONS.search, label: 'AI triage' });
  }
}

async function apiJSON(url, opts = {}) {
  const res = await fetch(url, {
    cache: 'no-store',
    headers: opts.body ? { 'Content-Type': 'application/json' } : undefined,
    ...opts,
  });
  const text = await res.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch (_) {
      data = null;
    }
  }
  if (!res.ok) {
    const msg = (data && (data.error || data.message)) || text || `HTTP ${res.status}`;
    throw new Error(typeof msg === 'string' ? msg : `HTTP ${res.status}`);
  }
  return data;
}

async function openManageSync() {
  const state = {
    tab: 'tracked',
    /** Host key for Tracked subtabs (github, gitlab, …). */
    trackedHost: '',
    payload: null,
    candidates: null,
    /** @type {{ id: string, label: string, projects: string[] }[] | null} */
    viewDraft: null,
    /** @type {{ orgs: string, groups: string } | null} */
    sourcesDraft: null,
    /** @type {{ roots: string } | null} */
    localDraft: null,
    dirty: false,
    status: '',
  };

  const root = el('div', 'manage-root');
  const tabs = el('div', 'manage-tabs');
  const panel = el('div', 'manage-panel');
  const statusEl = el('p', 'manage-empty');
  root.appendChild(tabs);
  root.appendChild(statusEl);
  root.appendChild(panel);

  const setStatus = (msg) => {
    state.status = msg || '';
    statusEl.textContent = state.status;
    statusEl.hidden = !state.status;
  };

  const markDirty = () => {
    state.dirty = true;
  };

  const clearDirty = () => {
    state.dirty = false;
  };

  const cloneViewDefs = (defs) => defs.map((v) => ({
    id: String(v.id || ''),
    label: String(v.label || ''),
    projects: Array.isArray(v.projects) ? [...v.projects] : [],
  }));

  const ensureViewDraft = () => {
    if (state.viewDraft) return state.viewDraft;
    const projects = Array.isArray(state.payload?.projects) ? state.payload.projects : [];
    let defs = Array.isArray(state.payload?.view_defs)
      ? cloneViewDefs(state.payload.view_defs)
      : [];
    if (defs.length === 0) {
      defs = [{
        id: 'default',
        label: 'Default',
        projects: projects.map((p) => p.id),
      }];
    }
    state.viewDraft = defs;
    return state.viewDraft;
  };

  const ensureSourcesDraft = () => {
    if (state.sourcesDraft) return state.sourcesDraft;
    const sync = state.payload?.sync || {};
    state.sourcesDraft = {
      orgs: (sync.github_orgs || []).join(', '),
      groups: (sync.gitlab_groups || []).join(', '),
    };
    return state.sourcesDraft;
  };

  const ensureLocalDraft = () => {
    if (state.localDraft) return state.localDraft;
    const roots = Array.isArray(state.payload?.local?.roots) ? state.payload.local.roots : [];
    state.localDraft = { roots: roots.join('\n') };
    return state.localDraft;
  };

  const refreshState = async () => {
    setStatus('Loading…');
    state.payload = await apiJSON('/api/views');
    setStatus('');
    renderPanel();
  };

  const tabDefs = [
    { id: 'tracked', label: 'Tracked' },
    { id: 'discover', label: 'Discover' },
    { id: 'views', label: 'Views' },
    { id: 'local', label: 'Local' },
    { id: 'sources', label: 'Sources' },
  ];

  const renderTabs = () => {
    tabs.replaceChildren();
    for (const t of tabDefs) {
      const btn = el('button', 'manage-tab');
      btn.type = 'button';
      btn.setAttribute('aria-selected', t.id === state.tab ? 'true' : 'false');
      btn.textContent = t.label;
      btn.addEventListener('click', () => {
        state.tab = t.id;
        renderTabs();
        renderPanel();
      });
      tabs.appendChild(btn);
    }
  };

  const renderTracked = () => {
    panel.replaceChildren();
    const projects = Array.isArray(state.payload?.projects) ? state.payload.projects : [];
    if (projects.length === 0) {
      panel.appendChild(el('p', 'manage-empty', 'No tracked projects yet. Use Discover or add sources first.'));
      return;
    }

    /** @type {Map<string, typeof projects>} */
    const byHost = new Map();
    for (const p of projects) {
      const host = String(p.host || '').toLowerCase() || 'unknown';
      let group = byHost.get(host);
      if (!group) {
        group = [];
        byHost.set(host, group);
      }
      group.push(p);
    }
    const hosts = [...byHost.keys()].sort((a, b) => a.localeCompare(b));
    if (!state.trackedHost || !byHost.has(state.trackedHost)) {
      state.trackedHost = hosts[0];
    }

    const subtabs = el('div', 'manage-subtabs');
    subtabs.setAttribute('role', 'tablist');
    subtabs.setAttribute('aria-label', 'Tracked by host');
    for (const host of hosts) {
      const btn = el('button', 'manage-subtab');
      btn.type = 'button';
      btn.setAttribute('role', 'tab');
      btn.setAttribute('aria-selected', host === state.trackedHost ? 'true' : 'false');
      const n = byHost.get(host)?.length ?? 0;
      btn.textContent = `${hostLabel(host)} (${n})`;
      btn.addEventListener('click', () => {
        if (host === state.trackedHost) return;
        state.trackedHost = host;
        renderTracked();
      });
      subtabs.appendChild(btn);
    }
    panel.appendChild(subtabs);

    const list = el('div', 'manage-tracked-list');
    const group = [...(byHost.get(state.trackedHost) || [])].sort((a, b) => {
      const la = String(a.label || a.id || '').toLowerCase();
      const lb = String(b.label || b.id || '').toLowerCase();
      return la.localeCompare(lb);
    });

    /** Serialize local_path saves so rapid edits do not race. */
    let localPathChain = Promise.resolve();

    const saveLocalPath = async (projectId, localPath, busyBtn) => {
      if (busyBtn) setButtonBusy(busyBtn, 'saving…');
      try {
        state.payload = await apiJSON('/api/sync/projects', {
          method: 'POST',
          body: JSON.stringify({
            action: 'set_local_path',
            id: projectId,
            local_path: localPath,
          }),
        });
        setStatus(localPath ? `Local path set for ${projectId}` : `Local path cleared for ${projectId}`);
        renderPanel();
      } catch (err) {
        setStatus(err instanceof Error ? err.message : String(err));
        throw err;
      } finally {
        if (busyBtn) setButtonIdle(busyBtn, { label: busyBtn.dataset.idleLabel || 'Save path' });
      }
    };

    for (const p of group) {
      const row = el('div', 'manage-row manage-row--local');
      const main = el('div', 'manage-row-main');
      main.appendChild(el('div', '', p.label || p.id));
      main.appendChild(el('div', 'manage-row-meta', `${p.host} · ${p.path}`));

      const pathField = el('div', 'manage-field manage-field--inline');
      pathField.appendChild(el('label', '', 'Local path'));
      const pathInput = document.createElement('input');
      pathInput.type = 'text';
      pathInput.placeholder = '~/code/repo or leave empty to use scan roots';
      pathInput.value = p.local_path || '';
      pathField.appendChild(pathInput);
      main.appendChild(pathField);

      const actions = el('div', 'manage-actions');
      const savePath = el('button', 'modal-btn modal-btn--ghost');
      savePath.type = 'button';
      savePath.dataset.idleLabel = 'Save path';
      setButtonLabel(savePath, null, 'Save path');
      savePath.addEventListener('click', () => {
        if (savePath.disabled) return;
        const next = pathInput.value.trim();
        localPathChain = localPathChain.then(
          () => saveLocalPath(p.id, next, savePath),
          () => saveLocalPath(p.id, next, savePath),
        );
      });
      pathInput.addEventListener('keydown', (ev) => {
        if (ev.key !== 'Enter') return;
        ev.preventDefault();
        savePath.click();
      });

      const clearPath = el('button', 'modal-btn modal-btn--ghost');
      clearPath.type = 'button';
      clearPath.dataset.idleLabel = 'Clear path';
      setButtonLabel(clearPath, null, 'Clear path');
      clearPath.disabled = !String(p.local_path || '').trim();
      clearPath.addEventListener('click', () => {
        if (clearPath.disabled) return;
        pathInput.value = '';
        localPathChain = localPathChain.then(
          () => saveLocalPath(p.id, '', clearPath),
          () => saveLocalPath(p.id, '', clearPath),
        );
      });

      const rm = el('button', 'modal-btn modal-btn--ghost');
      rm.type = 'button';
      setButtonLabel(rm, null, 'Remove');
      rm.addEventListener('click', async () => {
        if (rm.disabled) return;
        setButtonBusy(rm, 'removing…', `Removing ${p.id}`);
        try {
          state.payload = await apiJSON('/api/sync/projects', {
            method: 'POST',
            body: JSON.stringify({ action: 'remove', id: p.id }),
          });
          if (state.viewDraft) {
            for (const v of state.viewDraft) {
              v.projects = v.projects.filter((id) => id !== p.id);
            }
          }
          setStatus(`Removed ${p.id}`);
          renderPanel();
        } catch (err) {
          setButtonIdle(rm, { label: 'Remove' });
          setStatus(err instanceof Error ? err.message : String(err));
        }
      });
      actions.appendChild(savePath);
      actions.appendChild(clearPath);
      actions.appendChild(rm);
      row.appendChild(main);
      row.appendChild(actions);
      list.appendChild(row);
    }
    panel.appendChild(list);
  };

  const renderDiscover = () => {
    panel.replaceChildren();
    const actions = el('div', 'manage-actions');
    const loadBtn = el('button', 'modal-btn modal-btn--ok');
    loadBtn.type = 'button';
    setButtonLabel(loadBtn, null, 'Load candidates');
    actions.appendChild(loadBtn);
    panel.appendChild(actions);
    const list = el('div', 'manage-check-list');
    panel.appendChild(list);

    const trackKey = (host, path) => `${String(host || '').toLowerCase()}|${String(path || '').replace(/^\/+|\/+$/g, '')}`;

    const findTracked = (host, path) => {
      const key = trackKey(host, path);
      const projects = Array.isArray(state.payload?.projects) ? state.payload.projects : [];
      return projects.find((p) => trackKey(p.host || p.Host, p.path || p.Path) === key) || null;
    };

    const setCandidateTracked = (host, path, tracked) => {
      if (!Array.isArray(state.candidates)) return;
      const key = trackKey(host, path);
      for (const c of state.candidates) {
        if (trackKey(c.Host || c.host, c.Path || c.path) === key) {
          c.Tracked = tracked;
          c.tracked = tracked;
        }
      }
    };

    /** Serialize checkbox saves so rapid toggles do not race config writes. */
    let saveChain = Promise.resolve();

    const paintCandidates = () => {
      list.replaceChildren();
      const cands = Array.isArray(state.candidates) ? state.candidates : [];
      if (cands.length === 0) {
        list.appendChild(el('p', 'manage-empty', 'No candidates. Configure sync sources, then load.'));
        return;
      }
      for (const c of cands) {
        const host = c.Host || c.host || '';
        const path = c.Path || c.path || '';
        const lab = el('label', '');
        const cb = document.createElement('input');
        cb.type = 'checkbox';
        cb.checked = Boolean(c.Tracked || c.tracked);
        cb.dataset.host = host;
        cb.dataset.path = path;
        cb.addEventListener('change', () => {
          const wantTracked = cb.checked;
          cb.disabled = true;
          saveChain = saveChain.then(async () => {
            try {
              if (wantTracked) {
                state.payload = await apiJSON('/api/sync/projects', {
                  method: 'POST',
                  body: JSON.stringify({ action: 'add', host, path }),
                });
                setCandidateTracked(host, path, true);
                setStatus(`Tracked ${host} ${path}`);
              } else {
                const proj = findTracked(host, path);
                if (!proj?.id) {
                  throw new Error(`not tracked: ${host} ${path}`);
                }
                state.payload = await apiJSON('/api/sync/projects', {
                  method: 'POST',
                  body: JSON.stringify({ action: 'remove', id: proj.id }),
                });
                if (state.viewDraft) {
                  for (const v of state.viewDraft) {
                    v.projects = v.projects.filter((id) => id !== proj.id);
                  }
                }
                setCandidateTracked(host, path, false);
                setStatus(`Untracked ${host} ${path}`);
              }
            } catch (err) {
              cb.checked = !wantTracked;
              setStatus(err instanceof Error ? err.message : String(err));
            } finally {
              cb.disabled = false;
            }
          });
        });
        lab.appendChild(cb);
        lab.appendChild(document.createTextNode(`${host} ${path}`));
        list.appendChild(lab);
      }
    };

    loadBtn.addEventListener('click', async () => {
      if (loadBtn.disabled) return;
      setButtonBusy(loadBtn, 'loading…', 'Discovering forge repositories');
      try {
        const data = await apiJSON('/api/sync/candidates');
        state.candidates = data.candidates || [];
        const warnings = Array.isArray(data.warnings) ? data.warnings.filter(Boolean) : [];
        let msg = `${state.candidates.length} candidates`;
        if (warnings.length) {
          msg += `; ${warnings.join(' · ')}`;
        }
        setStatus(msg);
        paintCandidates();
      } catch (err) {
        setStatus(err instanceof Error ? err.message : String(err));
      } finally {
        setButtonIdle(loadBtn, { label: 'Load candidates', title: '' });
      }
    });

    paintCandidates();
  };

  const renderViews = () => {
    panel.replaceChildren();
    const projects = Array.isArray(state.payload?.projects) ? state.payload.projects : [];
    const defs = ensureViewDraft();

    const list = el('div', '');
    panel.appendChild(list);

    /** Serialize view PUTs so rapid toggles do not race config writes. */
    let viewsSaveChain = Promise.resolve();
    let viewsDebounceTimer = 0;

    const syncDefsFromPayload = () => {
      const next = cloneViewDefs(state.payload?.view_defs || defs);
      defs.splice(0, defs.length, ...next);
      state.viewDraft = defs;
    };

    const persistViews = async (busyBtn, busyLabel) => {
      for (const v of defs) {
        if (!String(v.id || '').trim() || !String(v.label || '').trim()) {
          setStatus('View id and label required');
          return;
        }
      }
      if (busyBtn) setButtonBusy(busyBtn, busyLabel || 'saving…');
      try {
        state.payload = await apiJSON('/api/views', {
          method: 'PUT',
          body: JSON.stringify({ views: defs }),
        });
        syncDefsFromPayload();
        setStatus('Views saved');
      } catch (err) {
        setStatus(err instanceof Error ? err.message : String(err));
        throw err;
      } finally {
        if (busyBtn) setButtonIdle(busyBtn, { label: busyBtn.dataset.idleLabel || 'Delete view' });
      }
    };

    const queueViewsSave = (busyBtn, busyLabel) => {
      viewsSaveChain = viewsSaveChain.then(
        () => persistViews(busyBtn, busyLabel),
        () => persistViews(busyBtn, busyLabel),
      );
      return viewsSaveChain;
    };

    const scheduleViewsSave = () => {
      window.clearTimeout(viewsDebounceTimer);
      viewsDebounceTimer = window.setTimeout(() => {
        void queueViewsSave();
      }, 400);
    };

    const paint = () => {
      list.replaceChildren();
      defs.forEach((view, idx) => {
        const block = el('div', 'manage-view-block');
        const fields = el('div', 'manage-field-row');

        const idField = el('div', 'manage-field');
        const idLab = el('label', '', 'View id');
        const idInput = document.createElement('input');
        idInput.value = view.id;
        idInput.addEventListener('input', () => {
          defs[idx].id = idInput.value.trim();
          scheduleViewsSave();
        });
        idInput.addEventListener('blur', () => {
          window.clearTimeout(viewsDebounceTimer);
          void queueViewsSave();
        });
        idField.appendChild(idLab);
        idField.appendChild(idInput);
        fields.appendChild(idField);

        const labelField = el('div', 'manage-field');
        const labelLab = el('label', '', 'Label');
        const labelInput = document.createElement('input');
        labelInput.value = view.label;
        labelInput.addEventListener('input', () => {
          defs[idx].label = labelInput.value.trim();
          scheduleViewsSave();
        });
        labelInput.addEventListener('blur', () => {
          window.clearTimeout(viewsDebounceTimer);
          void queueViewsSave();
        });
        labelField.appendChild(labelLab);
        labelField.appendChild(labelInput);
        fields.appendChild(labelField);
        block.appendChild(fields);

        const checks = el('div', 'manage-check-list');
        for (const p of projects) {
          const lab = el('label', '');
          const cb = document.createElement('input');
          cb.type = 'checkbox';
          cb.checked = view.projects.includes(p.id);
          cb.addEventListener('change', () => {
            const want = cb.checked;
            const set = new Set(defs[idx].projects);
            if (want) set.add(p.id);
            else set.delete(p.id);
            defs[idx].projects = [...set];
            cb.disabled = true;
            void queueViewsSave().then(
              () => { cb.disabled = false; },
              () => {
                cb.checked = !want;
                cb.disabled = false;
              },
            );
          });
          lab.appendChild(cb);
          lab.appendChild(document.createTextNode(`${p.label || p.id} (${p.path})`));
          checks.appendChild(lab);
        }
        block.appendChild(checks);

        const rm = el('button', 'modal-btn modal-btn--ghost manage-view-delete');
        rm.type = 'button';
        rm.dataset.idleLabel = 'Delete view';
        setButtonLabel(rm, null, 'Delete view');
        rm.addEventListener('click', () => {
          if (rm.disabled) return;
          if (defs.length <= 1) {
            setStatus('Keep at least one view');
            return;
          }
          const removed = defs.splice(idx, 1)[0];
          void queueViewsSave(rm, 'deleting…').then(
            () => paint(),
            () => {
              defs.splice(idx, 0, removed);
              paint();
            },
          );
        });
        block.appendChild(rm);
        list.appendChild(block);
      });
    };

    const actions = el('div', 'manage-actions manage-actions--footer');
    const addBtn = el('button', 'modal-btn modal-btn--ghost');
    addBtn.type = 'button';
    addBtn.dataset.idleLabel = 'Add view';
    setButtonLabel(addBtn, null, 'Add view');
    addBtn.addEventListener('click', () => {
      if (addBtn.disabled) return;
      defs.push({ id: `view-${defs.length + 1}`, label: 'New view', projects: [] });
      void queueViewsSave(addBtn, 'saving…').then(
        () => paint(),
        () => paint(),
      );
    });
    actions.appendChild(addBtn);
    panel.appendChild(actions);
    paint();
  };

  const renderSources = () => {
    panel.replaceChildren();
    const draft = ensureSourcesDraft();
    const orgField = el('div', 'manage-field');
    orgField.appendChild(el('label', '', 'GitHub orgs (comma-separated)'));
    const orgInput = document.createElement('input');
    orgInput.value = draft.orgs;
    orgInput.addEventListener('input', () => {
      draft.orgs = orgInput.value;
      markDirty();
    });
    orgField.appendChild(orgInput);
    const groupField = el('div', 'manage-field');
    groupField.appendChild(el('label', '', 'GitLab groups (comma-separated)'));
    const groupInput = document.createElement('input');
    groupInput.value = draft.groups;
    groupInput.addEventListener('input', () => {
      draft.groups = groupInput.value;
      markDirty();
    });
    groupField.appendChild(groupInput);
    const saveBtn = el('button', 'modal-btn modal-btn--ok');
    saveBtn.type = 'button';
    setButtonLabel(saveBtn, null, 'Save sources');
    saveBtn.addEventListener('click', async () => {
      if (saveBtn.disabled) return;
      const split = (s) => String(s || '').split(/[,;\s]+/).map((x) => x.trim()).filter(Boolean);
      setButtonBusy(saveBtn, 'saving…', 'Saving sync sources');
      try {
        const data = await apiJSON('/api/sync/sources', {
          method: 'PUT',
          body: JSON.stringify({
            github_orgs: split(orgInput.value),
            gitlab_groups: split(groupInput.value),
          }),
        });
        if (state.payload) {
          state.payload.sync = {
            github_orgs: data.github_orgs || [],
            gitlab_groups: data.gitlab_groups || [],
          };
        }
        state.sourcesDraft = {
          orgs: (data.github_orgs || []).join(', '),
          groups: (data.gitlab_groups || []).join(', '),
        };
        clearDirty();
        setStatus('Sources saved');
      } catch (err) {
        setStatus(err instanceof Error ? err.message : String(err));
      } finally {
        setButtonIdle(saveBtn, { label: 'Save sources', title: '' });
      }
    });
    const actions = el('div', 'manage-actions manage-actions--footer');
    actions.appendChild(saveBtn);
    panel.appendChild(orgField);
    panel.appendChild(groupField);
    panel.appendChild(actions);
  };

  const renderLocal = () => {
    panel.replaceChildren();
    const draft = ensureLocalDraft();
    panel.appendChild(el('p', 'manage-empty', 'Directories scanned for git checkouts (one path per line). Match is via origin remote.'));
    const rootsField = el('div', 'manage-field');
    rootsField.appendChild(el('label', '', 'Scan roots'));
    const rootsInput = document.createElement('textarea');
    rootsInput.rows = 6;
    rootsInput.value = draft.roots;
    rootsInput.addEventListener('input', () => {
      draft.roots = rootsInput.value;
      markDirty();
    });
    rootsField.appendChild(rootsInput);
    const saveBtn = el('button', 'modal-btn modal-btn--ok');
    saveBtn.type = 'button';
    setButtonLabel(saveBtn, null, 'Save roots');
    saveBtn.addEventListener('click', async () => {
      if (saveBtn.disabled) return;
      const roots = String(rootsInput.value || '')
        .split(/\n+/)
        .map((x) => x.trim())
        .filter(Boolean);
      setButtonBusy(saveBtn, 'saving…', 'Saving local scan roots');
      try {
        const data = await apiJSON('/api/local/roots', {
          method: 'PUT',
          body: JSON.stringify({ roots }),
        });
        if (state.payload) {
          state.payload.local = { roots: data.roots || [] };
        }
        state.localDraft = { roots: (data.roots || []).join('\n') };
        clearDirty();
        setStatus(`Saved ${(data.roots || []).length} scan root(s)`);
      } catch (err) {
        setStatus(err instanceof Error ? err.message : String(err));
      } finally {
        setButtonIdle(saveBtn, { label: 'Save roots', title: '' });
      }
    });
    const actions = el('div', 'manage-actions manage-actions--footer');
    actions.appendChild(saveBtn);
    panel.appendChild(rootsField);
    panel.appendChild(actions);
  };

  const renderPanel = () => {
    if (state.tab === 'tracked') renderTracked();
    else if (state.tab === 'discover') renderDiscover();
    else if (state.tab === 'views') renderViews();
    else if (state.tab === 'local') renderLocal();
    else renderSources();
  };

  renderTabs();
  try {
    await refreshState();
  } catch (err) {
    setStatus(err instanceof Error ? err.message : String(err));
    renderPanel();
  }

  await infoDialog({
    title: 'Manage projects and views',
    bodyNode: root,
    manage: true,
    confirmLabel: 'Done',
    beforeClose: () => {
      if (!state.dirty) return true;
      return window.confirm('You have unsaved changes. Discard them and close?');
    },
  });
  viewPayloadCache.clear();
  await loadDashboard({ fresh: true });
}

document.getElementById('manage-sync')?.addEventListener('click', () => {
  void openManageSync();
});

document.getElementById('refresh').addEventListener('click', () => {
  const btn = document.getElementById('refresh');
  setButtonBusy(btn, 'refreshing…', 'Bypass upstream cache and reload');
  void loadDashboard({ fresh: true }).finally(() => {
    setButtonIdle(btn, {
      label: 'Refresh',
      title: 'Bypass upstream cache and reload',
    });
  });
});
document.addEventListener('visibilitychange', () => {
  if (document.hidden) return;
  if (pendingPulls.size > 0 || pullFlushScheduled) return;
  void loadDashboard({ quiet: true, fresh: pullBoardDirty });
});

function findProjectById(id) {
  const want = String(id || '');
  if (!want) return null;
  return allProjects.find((p) => p.id === want) || null;
}

function findBranchOnProject(project, branchName) {
  const name = String(branchName || '');
  if (!project || !name) return null;
  const list = Array.isArray(project.branches) ? project.branches : [];
  return list.find((b) => b.name === name) || { name, run_id: '', ci_status: 'failed' };
}

function bindBoardActions() {
  const tbody = document.getElementById('rows');
  if (!tbody || tbody.dataset.boundActions === '1') return;
  tbody.dataset.boundActions = '1';
  tbody.addEventListener('click', (ev) => {
    const target = ev.target;
    if (!(target instanceof Element)) return;
    const btn = target.closest('[data-action]');
    if (!btn || !tbody.contains(btn)) return;
    const action = btn.getAttribute('data-action');
    if (!action) return;

    if (action === 'copy-path') {
      ev.preventDefault();
      ev.stopPropagation();
      void copyText(btn.getAttribute('data-path') || '', btn);
      return;
    }

    if (action === 'toggle-appearance') {
      ev.preventDefault();
      const row = btn.closest('.appearance-row');
      if (!row) return;
      const detail = row.querySelector(':scope > .appearance-detail');
      const open = btn.getAttribute('aria-expanded') === 'true';
      btn.setAttribute('aria-expanded', open ? 'false' : 'true');
      if (detail) detail.hidden = open;
      row.classList.toggle('is-open', !open);
      return;
    }

    if (action === 'pull') {
      if (btn.disabled || btn.classList.contains('is-busy') || btn.classList.contains('is-blocked')) return;
      ev.preventDefault();
      ev.stopPropagation();
      const behind = Number(btn.getAttribute('data-behind') || '0');
      void pullFFCheckout({
        project_id: btn.getAttribute('data-project-id') || '',
        branch: btn.getAttribute('data-branch') || '',
        repo_path: btn.getAttribute('data-repo-path') || '',
        behind,
        button: btn,
      });
      return;
    }

    if (action === 'inspect-sync') {
      if (btn.disabled || btn.classList.contains('is-busy')) return;
      ev.preventDefault();
      ev.stopPropagation();
      void investigateSync({
        project_id: btn.getAttribute('data-project-id') || '',
        branch: btn.getAttribute('data-branch') || '',
        repo_path: btn.getAttribute('data-repo-path') || '',
      }, btn);
      return;
    }

    if (action === 'prune-safe') {
      if (btn.disabled || btn.classList.contains('is-busy')) return;
      ev.preventDefault();
      ev.stopPropagation();
      void pruneSafeCheckout({
        project_id: btn.getAttribute('data-project-id') || '',
        branch: btn.getAttribute('data-branch') || '',
        worktree_path: btn.getAttribute('data-worktree-path') || '',
        button: btn,
      });
      return;
    }

    if (action === 'investigate-prune') {
      if (btn.disabled || btn.classList.contains('is-busy')) return;
      ev.preventDefault();
      ev.stopPropagation();
      void investigatePrune({
        project_id: btn.getAttribute('data-project-id') || '',
        branch: btn.getAttribute('data-branch') || '',
        worktree_path: btn.getAttribute('data-worktree-path') || '',
        default_branch: btn.getAttribute('data-default-branch') || 'main',
      }, btn);
      return;
    }

    if (action === 'triage') {
      ev.preventDefault();
      ev.stopPropagation();
      const project = findProjectById(btn.getAttribute('data-project-id') || '');
      if (!project) return;
      const branch = findBranchOnProject(project, btn.getAttribute('data-branch') || '');
      if (branch && btn.getAttribute('data-run-id')) {
        branch.run_id = btn.getAttribute('data-run-id');
      }
      void showFailures(project, branch);
    }
  });
}

bindFilters();
bindModal();
bindBoardActions();
updatePollLabel();
schedulePoll();

/**
 * Paint the view tab strip from /api/meta (cheap), then load the active view.
 * Keeps the board chrome visible while Collect runs.
 */
async function bootstrapBoard() {
  const status = document.getElementById('status');
  try {
    const res = await fetch('/api/meta', { cache: 'no-store' });
    if (res.ok) {
      const meta = await res.json();
      if (typeof meta.poll_interval_seconds === 'number') {
        const next = meta.poll_interval_seconds;
        if (next !== pollSeconds) {
          pollSeconds = next;
          updatePollLabel();
          schedulePoll();
        }
      }
      if (meta.ui) renderConfigInfo(meta.ui);
      const views = Array.isArray(meta.views) ? meta.views : [];
      if (views.length) {
        let active = activeViewId;
        if (!active || !views.some((v) => v.id === active)) {
          active = String(views[0].id || '').trim();
          if (active) persistActiveView(active);
        }
        loadingViewId = active;
        renderViewSwitcher(views, active);
        if (status) status.textContent = 'Loading…';
      }
    }
  } catch (_) {
    /* Dashboard fetch still runs below. */
  }
  await loadDashboard({ quiet: false, fresh: false, viewBusy: Boolean(loadingViewId) });
}

void bootstrapBoard();
