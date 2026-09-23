import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { createDashboardLoadArbiter } from './dashboard_load.js';

describe('createDashboardLoadArbiter', () => {
  it('blocks quiet poll while a fresh load is in flight', () => {
    const a = createDashboardLoadArbiter();
    const fresh = a.begin({ fresh: true, reason: 'action' });
    assert.equal(fresh.allowed, true);
    assert.equal(a.hasFreshInFlight(), true);

    const poll = a.begin({ fresh: false, reason: 'poll' });
    assert.equal(poll.allowed, false);
    assert.equal(a.isCurrent(fresh.gen), true);

    a.finish(fresh.gen, fresh.trackingFresh);
    assert.equal(a.hasFreshInFlight(), false);

    const poll2 = a.begin({ fresh: false, reason: 'poll' });
    assert.equal(poll2.allowed, true);
  });

  it('lets view-switch supersede an in-flight fresh load', () => {
    const a = createDashboardLoadArbiter();
    const fresh = a.begin({ fresh: true, reason: 'action' });
    assert.equal(fresh.allowed, true);

    const view = a.begin({ fresh: false, reason: 'view-switch' });
    assert.equal(view.allowed, true);
    assert.equal(a.isCurrent(fresh.gen), false);
    assert.equal(a.isCurrent(view.gen), true);
    assert.equal(fresh.signal.aborted, true);
    assert.equal(view.signal.aborted, false);

    // Stale fresh must still clear ownership so later polls are not stuck.
    a.finish(fresh.gen, fresh.trackingFresh);
    assert.equal(a.hasFreshInFlight(), false);

    a.finish(view.gen, view.trackingFresh);
    const poll = a.begin({ fresh: false, reason: 'poll' });
    assert.equal(poll.allowed, true);
  });

  it('lets view-switch supersede while fresh ownership is still marked', () => {
    const a = createDashboardLoadArbiter();
    const fresh = a.begin({ fresh: true, reason: 'refresh' });
    const cached = a.begin({ fresh: false, reason: 'view-switch' });
    assert.equal(cached.allowed, true);
    assert.equal(a.hasFreshInFlight(), true);

    a.finish(fresh.gen, true);
    assert.equal(a.hasFreshInFlight(), false);
  });

  it('only the current generation may paint', () => {
    const a = createDashboardLoadArbiter();
    const first = a.begin({ fresh: true, reason: 'action' });
    const second = a.begin({ fresh: false, reason: 'view-switch' });

    const painted = [];
    const maybePaint = (lease, label) => {
      if (a.isCurrent(lease.gen)) painted.push(label);
    };
    maybePaint(first, 'A');
    maybePaint(second, 'B');
    assert.deepEqual(painted, ['B']);
  });

  it('streaming chunks from a superseded gen are ignored', () => {
    const a = createDashboardLoadArbiter();
    const streamA = a.begin({ fresh: true, reason: 'action' });
    const streamB = a.begin({ fresh: false, reason: 'view-switch' });

    const events = [];
    const handle = (lease, type) => {
      if (!a.isCurrent(lease.gen)) return;
      events.push(`${lease.gen}:${type}`);
    };
    handle(streamA, 'shell');
    handle(streamA, 'project');
    handle(streamB, 'shell');
    handle(streamA, 'done');
    handle(streamB, 'done');
    assert.deepEqual(events, [`${streamB.gen}:shell`, `${streamB.gen}:done`]);
  });

  it('finish clears fresh ownership even when not current', () => {
    const a = createDashboardLoadArbiter();
    const fresh = a.begin({ fresh: true, reason: 'action' });
    a.begin({ fresh: true, reason: 'refresh' });
    assert.equal(a.hasFreshInFlight(), true);
    // Finishing the older lease must not clear the newer owner's flag.
    a.finish(fresh.gen, true);
    assert.equal(a.hasFreshInFlight(), true);
  });
});
