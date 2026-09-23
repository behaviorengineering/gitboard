/**
 * Dashboard load request arbitration (DOM-free).
 *
 * Quiet non-fresh loads must not abort an in-flight ?fresh=1 paint.
 * View switches always supersede, including an in-flight fresh load, so the
 * selected tab cannot jump back when the older response arrives.
 */

/**
 * @typedef {{ allowed: false }} DashboardLoadDenied
 * @typedef {{
 *   allowed: true,
 *   gen: number,
 *   signal: AbortSignal,
 *   abortController: AbortController,
 *   trackingFresh: boolean,
 * }} DashboardLoadLease
 */

/**
 * @returns {{
 *   begin: (opts?: { fresh?: boolean, reason?: string }) => DashboardLoadDenied | DashboardLoadLease,
 *   isCurrent: (gen: number) => boolean,
 *   finish: (gen: number, trackingFresh: boolean) => void,
 *   releaseAbort: (gen: number, ac: AbortController) => void,
 *   currentGen: () => number,
 *   hasFreshInFlight: () => boolean,
 * }}
 */
export function createDashboardLoadArbiter() {
  let gen = 0;
  /** Generation that owns dashboardFreshInFlight; 0 means none. */
  let freshOwnerGen = 0;
  /** @type {AbortController | null} */
  let abort = null;

  return {
    /**
     * @param {{ fresh?: boolean, reason?: string }} [opts]
     * reason "view-switch" always supersedes; other non-fresh loads defer while fresh is in flight.
     */
    begin({ fresh = false, reason = '' } = {}) {
      const isViewSwitch = reason === 'view-switch';
      const wantFresh = Boolean(fresh);

      if (!wantFresh && !isViewSwitch && freshOwnerGen !== 0) {
        return { allowed: false };
      }

      const nextGen = ++gen;
      if (abort) {
        abort.abort();
      }
      const ac = new AbortController();
      abort = ac;

      const trackingFresh = wantFresh;
      if (trackingFresh) {
        freshOwnerGen = nextGen;
      } else if (isViewSwitch && freshOwnerGen !== 0) {
        // Superseded fresh must not keep blocking later quiet loads; the aborted
        // lease still calls finish() and clears only if it still owns the flag.
        // Bumping ownership away here would strand the flag if finish uses ===.
        // Keep freshOwnerGen on the old lease; finish clears it when that lease ends.
      }

      return {
        allowed: true,
        gen: nextGen,
        signal: ac.signal,
        abortController: ac,
        trackingFresh,
      };
    },

    isCurrent(leaseGen) {
      return leaseGen === gen;
    },

    /**
     * Release fresh ownership when this lease was the owner, even if a newer
     * generation already superseded the paint (so the flag cannot stick).
     * @param {number} leaseGen
     * @param {boolean} trackingFresh
     */
    finish(leaseGen, trackingFresh) {
      if (trackingFresh && freshOwnerGen === leaseGen) {
        freshOwnerGen = 0;
      }
    },

    /**
     * @param {number} leaseGen
     * @param {AbortController} ac
     */
    releaseAbort(leaseGen, ac) {
      if (leaseGen === gen && abort === ac) {
        abort = null;
      }
    },

    currentGen() {
      return gen;
    },

    hasFreshInFlight() {
      return freshOwnerGen !== 0;
    },
  };
}
