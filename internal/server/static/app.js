function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text != null) node.textContent = text;
  return node;
}

function badgeClass(status) {
  const s = String(status || '').toLowerCase();
  if (['success', 'passed'].includes(s)) return 'badge badge--success';
  if (['failed', 'failure', 'error', 'cancelled', 'canceled'].includes(s)) return 'badge badge--failed';
  if (['running', 'pending', 'in_progress', 'queued'].includes(s)) return 'badge badge--running';
  return 'badge badge--neutral';
}

function renderTooling(tool) {
  const root = document.getElementById('tooling');
  root.innerHTML = '';
  const gh = tool?.github || {};
  const gl = tool?.gitlab || {};
  root.appendChild(el('span', '', `gh: ${gh.installed ? (gh.authed ? 'ready' : 'login required') : 'missing'}${gh.detail ? ` (${gh.detail})` : ''}`));
  root.appendChild(el('span', '', `glab: ${gl.installed ? (gl.authed ? 'ready' : 'login required') : 'missing'}${gl.detail ? ` (${gl.detail})` : ''}`));
}

function renderRows(projects) {
  const tbody = document.getElementById('rows');
  tbody.innerHTML = '';
  for (const row of projects || []) {
    const tr = el('tr');
    const titleCell = el('td');
    const title = el('strong', '', row.label || row.id);
    titleCell.appendChild(title);
    if (row.open_url) {
      const link = el('a', '', ' open');
      link.href = row.open_url;
      link.target = '_blank';
      link.rel = 'noopener';
      titleCell.appendChild(link);
    }
    if (row.error) {
      titleCell.appendChild(el('div', 'error', row.error));
    }
    tr.appendChild(titleCell);
    tr.appendChild(el('td', '', row.host || '—'));

    const ciCell = el('td');
    if (row.ci) {
      ciCell.appendChild(el('span', badgeClass(row.ci.status), row.ci.status || 'unknown'));
      if (row.ci.ref) ciCell.appendChild(el('div', '', row.ci.ref));
      if (row.ci.web_url) {
        const a = el('a', '', 'pipeline');
        a.href = row.ci.web_url;
        a.target = '_blank';
        a.rel = 'noopener';
        ciCell.appendChild(a);
      }
    } else {
      ciCell.textContent = '—';
    }
    tr.appendChild(ciCell);

    const openCell = el('td');
    const open = row.open_items || {};
    const bits = [];
    if (open.merge_requests) bits.push(`${open.merge_requests} MR`);
    if (open.pull_requests) bits.push(`${open.pull_requests} PR`);
    openCell.textContent = bits.length ? bits.join(', ') : '0';
    tr.appendChild(openCell);

    const actions = el('td');
    if (row.ci && ['failed', 'failure'].includes(String(row.ci.status).toLowerCase())) {
      const btn = el('button', '', 'Inspect');
      btn.addEventListener('click', () => showFailures(row));
      actions.appendChild(btn);
    }
    tr.appendChild(actions);
    tbody.appendChild(tr);
  }
}

async function loadDashboard() {
  const status = document.getElementById('status');
  status.textContent = 'Loading…';
  try {
    const res = await fetch('/api/dashboard', { cache: 'no-store' });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    renderTooling(data.tooling);
    renderRows(data.projects);
    status.textContent = `Updated ${data.generated_at || ''}`;
  } catch (err) {
    status.textContent = `Error: ${err instanceof Error ? err.message : String(err)}`;
  }
}

async function showFailures(project) {
  const detail = document.getElementById('detail');
  const jobsRoot = document.getElementById('jobs');
  const triage = document.getElementById('triage');
  detail.hidden = false;
  triage.hidden = true;
  document.getElementById('detail-title').textContent = `${project.label}: failed jobs`;
  jobsRoot.innerHTML = 'Loading jobs…';
  const params = new URLSearchParams({ project: project.id, run_id: project.ci?.run_id || '' });
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
    ai.addEventListener('click', () => runTriage(project, job));
    row.appendChild(ai);
    jobsRoot.appendChild(row);
  }
  if (!(data.jobs || []).length) {
    jobsRoot.textContent = 'No failed jobs returned. Try opening the pipeline link.';
  }
}

async function runTriage(project, job) {
  const triage = document.getElementById('triage');
  triage.hidden = false;
  triage.textContent = 'Running triage…';
  const res = await fetch('/api/triage', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      project_id: project.id,
      run_id: project.ci?.run_id || '',
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
void loadDashboard();
