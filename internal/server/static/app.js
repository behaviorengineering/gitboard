function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text != null) node.textContent = text;
  return node;
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
};

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
  const wrap = el('span', 'project-host');
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

/** Prefer primary appearance worktree, else main, else first. */
function pickLocalWorktree(wts, local) {
  const list = Array.isArray(wts) ? wts : (wts ? [wts] : []);
  if (!list.length) return null;
  const primaryPath = String(local?.path || '');
  const primary = list.find((w) => primaryPath && w.appearance_path === primaryPath);
  if (primary) return primary;
  const main = list.find((w) => w.main);
  return main || list[0];
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
  btn.insertAdjacentHTML('beforeend', ICONS.clipboard);
  btn.addEventListener('click', (ev) => {
    ev.preventDefault();
    ev.stopPropagation();
    void copyText(path, btn);
  });
  return btn;
}

function localColumn(wts, local, branchName) {
  const cell = el('div', 'branch-local');
  if (!local) {
    cell.appendChild(el('span', 'branch-local-empty', '—'));
    return cell;
  }
  if (!local.mapped) {
    cell.appendChild(el('span', 'branch-local-empty', ''));
    cell.title = 'Local checkout not mapped';
    return cell;
  }
  const list = Array.isArray(wts) ? wts : (wts ? [wts] : []);
  if (!list.length) {
    if (local.error) {
      cell.appendChild(el('span', 'branch-local-empty', '!'));
      cell.title = local.error;
      return cell;
    }
    cell.appendChild(el('span', 'branch-local-empty', '—'));
    cell.title = 'No local checkout on this branch';
    return cell;
  }

  const marks = el('span', 'branch-local-marks');
  const primaryPath = String(local.path || '');
  for (const wt of list) {
    const label = wt.appearance_label || wt.appearance_path || wt.path || 'local checkout';
    const isPrimary = primaryPath && wt.appearance_path === primaryPath;
    marks.appendChild(iconMark(
      ICONS.laptop,
      isPrimary || wt.main ? 'mark--default' : 'mark--muted',
      label,
    ));
    if (wt.dirty) {
      marks.appendChild(iconMark(ICONS.alert, 'mark--bad', `dirty: ${label}`));
    }
  }
  const preferred = pickLocalWorktree(list, local);
  if (preferred?.path) {
    const copyBtn = copyPathButton(preferred.path);
    if (copyBtn) marks.appendChild(copyBtn);
  }
  cell.appendChild(marks);

  const bits = [];
  if (preferred?.ahead) bits.push(`↑${preferred.ahead}`);
  if (preferred?.behind) bits.push(`↓${preferred.behind}`);
  if (branchName && local.default_branch === branchName) {
    if (local.default_behind) bits.push(`def↓${local.default_behind}`);
    if (local.default_ahead) bits.push(`def↑${local.default_ahead}`);
  }
  if (bits.length) {
    cell.appendChild(el('span', 'branch-local-sync', bits.join(' ')));
  }
  return cell;
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
  head.appendChild(el('span', 'branch-head-label', 'local'));
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
  const localWt = pickLocalWorktree(localWts, local);
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

  const marks = el('span', 'branch-marks');
  if (b.open_review) marks.appendChild(iconMark(ICONS.pr, 'mark--review', `open ${reviewKind}`));
  if (b.draft) marks.appendChild(iconMark(ICONS.draft, 'mark--muted', 'draft'));
  if (b.conflict) marks.appendChild(iconMark(ICONS.alert, 'mark--bad', 'merge conflicts'));
  if (b.stale) marks.appendChild(iconMark(ICONS.clock, 'mark--stale', 'stale (>14 days)'));
  item.appendChild(marks);

  const body = el('div', 'branch-body');
  const title = el('div', 'branch-title');
  if (b.default) title.appendChild(iconMark(ICONS.lock, 'mark--default', 'default branch'));
  const nameHref = b.web_url || localWt?.merged_url || '';
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
    const chip = el('span', 'branch-chip branch-chip--ok', 'safe to remove');
    const bits = [];
    if (localWt.merged_id) bits.push(`merged ${reviewKind} #${localWt.merged_id}`);
    const mergedWhen = relativeTime(localWt.merged_at);
    if (mergedWhen) bits.push(mergedWhen);
    if (pruneCmd) bits.push(pruneCmd);
    chip.title = bits.join(' · ');
    meta.appendChild(chip);
  } else if (pruneHint === 'likely') {
    const chip = el('span', 'branch-chip branch-chip--warn', 'likely removable');
    chip.title = pruneCmd
      ? `Remote head gone · check then: ${pruneCmd}`
      : 'Remote head gone';
    meta.appendChild(chip);
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

  item.appendChild(localColumn(localWts, local, b.name));

  const actions = el('div', 'branch-actions');
  const ci = ciMark(b.ci_status);
  if (ci) actions.appendChild(ci);
  else actions.appendChild(el('span', 'branch-ci-slot'));

  if (b.ci_url) {
    const go = el('a', 'branch-goto');
    go.href = b.ci_url;
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
  } else {
    actions.appendChild(el('span', 'branch-ci-slot'));
  }

  if (failed && b.run_id) {
    const btn = el('button', 'branch-inspect', 'Inspect');
    btn.type = 'button';
    btn.title = 'Inspect failed jobs';
    btn.addEventListener('click', () => showFailures(project, b));
    actions.appendChild(btn);
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

function appearanceStatusMarks(app) {
  const marks = el('span', 'appearance-marks');
  if (app.error) {
    marks.appendChild(iconMark(ICONS.alert, 'mark--bad', app.error));
    return marks;
  }
  if (app.branch) {
    const br = el('span', 'appearance-branch', app.detached ? `detached ${app.branch}` : app.branch);
    marks.appendChild(br);
  }
  if (app.dirty) marks.appendChild(iconMark(ICONS.alert, 'mark--bad', 'dirty working tree'));
  const sync = [];
  if (app.ahead) sync.push(`↑${app.ahead}`);
  if (app.behind) sync.push(`↓${app.behind}`);
  if (sync.length) marks.appendChild(el('span', 'appearance-sync', sync.join(' ')));
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
    if (wt.ahead) bits.push(`↑${wt.ahead}`);
    if (wt.behind) bits.push(`↓${wt.behind}`);
    li.appendChild(el('span', '', bits.join(' · ')));
    if (wt.path && wt.path !== app.path) {
      li.appendChild(el('span', 'appearance-wt-path', wt.path));
    }
    list.appendChild(li);
  }
  detail.appendChild(list);
  return detail;
}

function renderAppearances(local) {
  if (!local?.mapped) return null;
  const apps = Array.isArray(local.appearances) && local.appearances.length
    ? local.appearances
    : (local.path ? [{
      path: local.path,
      display_id: local.path,
      primary: true,
      branch: local.branch,
      dirty: local.dirty,
      ahead: local.ahead,
      behind: local.behind,
      detached: local.detached,
      error: local.error,
      worktrees: local.worktrees,
    }] : []);
  if (!apps.length) return null;

  const wrap = el('div', 'project-appearances');
  for (const app of apps) {
    const row = el('div', 'appearance-row' + (app.primary ? ' appearance-row--primary' : ''));
    const line = el('div', 'appearance-line');
    const copyBtn = copyPathButton(app.path);
    if (copyBtn) line.appendChild(copyBtn);
    const head = el('button', 'appearance-head');
    head.type = 'button';
    head.setAttribute('aria-expanded', 'false');
    const id = el('span', 'project-local-path appearance-id', app.display_id || app.path || '');
    id.title = app.path || '';
    head.appendChild(id);
    head.appendChild(appearanceStatusMarks(app));
    const detail = renderAppearanceDetail(app);
    detail.hidden = true;
    head.addEventListener('click', (ev) => {
      ev.preventDefault();
      const open = head.getAttribute('aria-expanded') === 'true';
      head.setAttribute('aria-expanded', open ? 'false' : 'true');
      detail.hidden = open;
      row.classList.toggle('is-open', !open);
    });
    line.appendChild(head);
    row.appendChild(line);
    row.appendChild(detail);
    wrap.appendChild(row);
  }
  return wrap;
}

function renderTooling(tool) {
  const root = document.getElementById('tooling');
  root.innerHTML = '';
  const gh = tool?.github || {};
  const gl = tool?.gitlab || {};
  const ghRow = el('span', 'tool-row');
  ghRow.insertAdjacentHTML('beforeend', ICONS.github);
  ghRow.appendChild(document.createTextNode(` gh: ${gh.installed ? (gh.authed ? 'ready' : 'login required') : 'missing'}${gh.detail ? ` (${gh.detail})` : ''}`));
  const glRow = el('span', 'tool-row');
  glRow.insertAdjacentHTML('beforeend', ICONS.gitlab);
  glRow.appendChild(document.createTextNode(` glab: ${gl.installed ? (gl.authed ? 'ready' : 'login required') : 'missing'}${gl.detail ? ` (${gl.detail})` : ''}`));
  root.appendChild(ghRow);
  root.appendChild(glRow);
}

const CI_FAILED = new Set(['failed', 'failure', 'error', 'cancelled', 'canceled']);

let allProjects = [];
const filters = {
  q: '',
  host: '',
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

function haystack(row) {
  const parts = [
    row.label,
    row.id,
    row.path,
    row.org,
    row.host,
  ];
  for (const b of row.branches || []) {
    if (b.name) parts.push(b.name);
  }
  const local = row.local;
  if (local?.path) parts.push(local.path);
  if (local?.branch) parts.push(local.branch);
  for (const app of local?.appearances || []) {
    if (app.display_id) parts.push(app.display_id);
    if (app.path) parts.push(app.path);
    if (app.branch) parts.push(app.branch);
    if (app.parent_label) parts.push(app.parent_label);
  }
  for (const wt of local?.worktrees || []) {
    if (wt.branch) parts.push(wt.branch);
    if (wt.path) parts.push(wt.path);
  }
  return normalizeQ(parts.filter(Boolean).join(' '));
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

function projectMatches(row, state) {
  if (state.host && String(row.host || '').toLowerCase() !== state.host) return false;
  if (state.chips.ci_failed) {
    const hit = (row.branches || []).some((b) => CI_FAILED.has(String(b.ci_status || '').toLowerCase()));
    if (!hit) return false;
  }
  if (state.chips.open_review) {
    const openCount = (row.open_items?.pull_requests || 0) + (row.open_items?.merge_requests || 0);
    const hit = openCount > 0 || (row.branches || []).some((b) => b.open_review);
    if (!hit) return false;
  }
  if (state.chips.dirty && !projectDirty(row)) return false;
  if (state.chips.prune && !projectPrune(row)) return false;
  if (state.chips.conflicts) {
    const hit = (row.branches || []).some((b) => b.conflict);
    if (!hit) return false;
  }
  if (state.q) {
    const hay = haystack(row);
    const tokens = state.q.split(/\s+/).filter(Boolean);
    if (!tokens.every((tok) => hay.includes(tok))) return false;
  }
  return true;
}

function filtersActive(state) {
  if (state.q || state.host) return true;
  return Object.values(state.chips).some(Boolean);
}

function filteredProjects() {
  return (allProjects || []).filter((row) => projectMatches(row, filters));
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

function applyBoard() {
  const rows = filteredProjects();
  renderRows(rows);
  updateFilterChrome(rows.length);
}

function readFiltersFromDom() {
  const qInput = document.getElementById('filter-q');
  const hostSel = document.getElementById('filter-host');
  filters.q = normalizeQ(qInput?.value);
  filters.host = String(hostSel?.value || '').toLowerCase();
  for (const btn of document.querySelectorAll('.filter-chip[data-filter]')) {
    const key = btn.getAttribute('data-filter');
    if (key && key in filters.chips) {
      filters.chips[key] = btn.getAttribute('aria-pressed') === 'true';
    }
  }
}

function clearFilters() {
  const qInput = document.getElementById('filter-q');
  const hostSel = document.getElementById('filter-host');
  if (qInput) qInput.value = '';
  if (hostSel) hostSel.value = '';
  for (const btn of document.querySelectorAll('.filter-chip[data-filter]')) {
    btn.setAttribute('aria-pressed', 'false');
  }
  filters.q = '';
  filters.host = '';
  for (const key of Object.keys(filters.chips)) filters.chips[key] = false;
  applyBoard();
}

function bindFilters() {
  const qInput = document.getElementById('filter-q');
  const hostSel = document.getElementById('filter-host');
  const clearBtn = document.getElementById('filter-clear');
  const onChange = () => {
    readFiltersFromDom();
    applyBoard();
  };
  qInput?.addEventListener('input', onChange);
  hostSel?.addEventListener('change', onChange);
  clearBtn?.addEventListener('click', () => clearFilters());
  for (const btn of document.querySelectorAll('.filter-chip[data-filter]')) {
    btn.addEventListener('click', () => {
      const pressed = btn.getAttribute('aria-pressed') === 'true';
      btn.setAttribute('aria-pressed', pressed ? 'false' : 'true');
      onChange();
    });
  }
}

function renderRows(projects) {
  const tbody = document.getElementById('rows');
  tbody.innerHTML = '';
  for (const row of projects || []) {
    const tr = el('tr');
    const titleCell = el('td', 'project-cell');
    if (row.org) {
      titleCell.appendChild(el('div', 'project-org', row.org));
    }
    const titleRow = el('div', 'project-title');
    titleRow.appendChild(hostPrefix(row.host, row.path));
    titleRow.appendChild(el('strong', '', row.label || row.id));
    if (row.open_url) {
      const link = el('a', 'project-open');
      link.href = row.open_url;
      link.target = '_blank';
      link.rel = 'noopener';
      link.title = 'Open repository';
      link.setAttribute('aria-label', `Open ${row.label || row.id}`);
      link.insertAdjacentHTML('beforeend', ICONS.external);
      titleRow.appendChild(link);
    }
    titleCell.appendChild(titleRow);
    const local = row.local;
    const appearanceBlock = renderAppearances(local);
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

    const actions = el('td', 'actions-cell');
    const failedBranch = (row.branches || []).find((b) =>
      ['failed', 'failure', 'error'].includes(String(b.ci_status || '').toLowerCase()) && b.run_id);
    if (failedBranch) {
      const btn = el('button', '', 'Inspect');
      btn.type = 'button';
      btn.addEventListener('click', () => showFailures(row, failedBranch));
      actions.appendChild(btn);
    }
    tr.appendChild(actions);
    tbody.appendChild(tr);
  }
}

let pollSeconds = 30;
let pollTimer = null;
let loading = false;

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
    if (document.hidden || loading) return;
    void loadDashboard({ quiet: true });
  }, pollSeconds * 1000);
}

async function loadDashboard({ quiet = false, fresh = false } = {}) {
  if (loading) return;
  loading = true;
  const status = document.getElementById('status');
  if (!quiet) status.textContent = fresh ? 'Force refreshing…' : 'Loading…';
  try {
    const url = fresh ? '/api/dashboard?fresh=1' : '/api/dashboard';
    const res = await fetch(url, { cache: 'no-store' });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    if (typeof data.poll_interval_seconds === 'number') {
      const next = data.poll_interval_seconds;
      if (next !== pollSeconds) {
        pollSeconds = next;
        updatePollLabel();
        schedulePoll();
      }
    }
    renderTooling(data.tooling);
    allProjects = Array.isArray(data.projects) ? data.projects : [];
    applyBoard();
    status.textContent = `Updated ${data.generated_at || ''}`;
  } catch (err) {
    status.textContent = `Error: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    loading = false;
  }
}

async function showFailures(project, branch) {
  const detail = document.getElementById('detail');
  const jobsRoot = document.getElementById('jobs');
  const triage = document.getElementById('triage');
  detail.hidden = false;
  triage.hidden = true;
  const branchName = branch?.name ? ` / ${branch.name}` : '';
  document.getElementById('detail-title').textContent = `${project.label}${branchName}: failed jobs`;
  jobsRoot.innerHTML = 'Loading jobs…';
  const runID = branch?.run_id || project.ci?.run_id || '';
  const params = new URLSearchParams({ project: project.id, run_id: runID });
  const res = await fetch(`/api/failures?${params}`);
  if (!res.ok) {
    jobsRoot.textContent = await res.text();
    return;
  }
  const data = await res.json();
  jobsRoot.innerHTML = '';
  for (const job of data.jobs || []) {
    const row = el('div', 'job-row');
    row.appendChild(el('span', '', job.name || job.id));
    if (job.web_url) {
      const a = el('a', '', 'log');
      a.href = job.web_url;
      a.target = '_blank';
      a.rel = 'noopener';
      row.appendChild(a);
    }
    const ai = el('button', '', 'AI triage');
    ai.addEventListener('click', () => runTriage(project, job, runID));
    row.appendChild(ai);
    jobsRoot.appendChild(row);
  }
  if (!(data.jobs || []).length) {
    jobsRoot.textContent = 'No failed jobs returned. Try opening the pipeline link.';
  }
}

async function runTriage(project, job, runID) {
  const triage = document.getElementById('triage');
  triage.hidden = false;
  triage.textContent = 'Running triage…';
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
  if (!res.ok) {
    triage.textContent = text;
    return;
  }
  const data = JSON.parse(text);
  if (data.unavailable) {
    triage.textContent = data.unavailable;
    return;
  }
  const lines = [
    data.summary && `Summary: ${data.summary}`,
    data.root_cause && `Root cause: ${data.root_cause}`,
    data.fix_steps?.length && `Fix steps:\n${data.fix_steps.map((s, i) => `${i + 1}. ${s}`).join('\n')}`,
    data.confidence && `Confidence: ${data.confidence}`,
    data.model && `Model: ${data.model}`,
  ].filter(Boolean);
  triage.textContent = lines.join('\n\n');
}

document.getElementById('refresh').addEventListener('click', () => { void loadDashboard(); });
document.getElementById('force-refresh').addEventListener('click', () => { void loadDashboard({ fresh: true }); });
document.addEventListener('visibilitychange', () => {
  if (!document.hidden) void loadDashboard({ quiet: true });
});
bindFilters();
updatePollLabel();
schedulePoll();
void loadDashboard();
