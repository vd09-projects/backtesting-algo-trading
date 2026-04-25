# GST base = brokerage + exchange charges only (STT, SEBI, stamp exempt)

- **Date:** 2026-04-25
- **Status:** accepted
- **Category:** convention
- **Tags:** commission, GST, NSE, SEBI, STT, stamp-duty, TASK-0038
- **Task:** TASK-0038

## Decision

GST (18%) is levied only on the sum of brokerage and NSE exchange transaction charges. STT, SEBI
charges, and stamp duty are excluded from the GST base.

## Rationale

This is the SEBI / Indian tax regulation:

- **STT (Securities Transaction Tax)** — a central government tax; GST is not applicable on taxes.
- **SEBI charges** — a regulatory levy; exempt from GST (similar to STT treatment).
- **Stamp duty** — a state-level duty; not subject to GST.
- **Brokerage** — a service fee; GST applies.
- **Exchange transaction charges** — exchange fees for trade processing; classified as a service;
  GST applies.

The Zerodha brokerage calculator (brokerage.zerodha.com) confirms this treatment and was used to
hand-verify the ₹88.24 round-trip expected value in the golden test.

## Consequences

The golden test for a ₹30,000 round-trip (buy + sell at same notional, default ₹20 brokerage cap)
verifies the exact expected value. Any change to which components are included in the GST base must
also update the golden test and this decision record.
