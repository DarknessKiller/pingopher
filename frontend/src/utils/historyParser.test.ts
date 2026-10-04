import { test } from "node:test";
import assert from "node:assert/strict";

import type { Host, Result } from "../api";
import { processHistoryResults } from "./historyParser.ts";

const host = {
  protocol: "https",
  acceptedStatusCodes: ["200"],
  pingInterval: 60,
} as unknown as Host;

const result = (dns: string, timestamp: string, down: boolean): Result => ({
  dns,
  statusCode: down ? 503 : 200,
  latency: "10 ms",
  timestamp,
  errorMsg: down ? "boom" : undefined,
});

test("one failing resolver marks the round down", () => {
  const { uptimePercent } = processHistoryResults(
    [
      result("a", "2026-01-01T00:00:00Z", true),
      result("b", "2026-01-01T00:00:00Z", false),
      result("a", "2026-01-01T00:01:00Z", false),
      result("b", "2026-01-01T00:01:00Z", false),
    ],
    host,
  );
  assert.equal(uptimePercent, 50);
});

test("a second resolver failing in the same round is not counted twice", () => {
  const oneFailed = processHistoryResults(
    [
      result("a", "2026-01-01T00:00:00Z", true),
      result("b", "2026-01-01T00:00:00Z", false),
      result("a", "2026-01-01T00:01:00Z", false),
      result("b", "2026-01-01T00:01:00Z", false),
    ],
    host,
  );
  const twoFailed = processHistoryResults(
    [
      result("a", "2026-01-01T00:00:00Z", true),
      result("b", "2026-01-01T00:00:00Z", true),
      result("a", "2026-01-01T00:01:00Z", false),
      result("b", "2026-01-01T00:01:00Z", false),
    ],
    host,
  );
  assert.equal(twoFailed.uptimePercent, oneFailed.uptimePercent);
});

test("a round with every resolver up stays up", () => {
  const { uptimePercent } = processHistoryResults(
    [
      result("a", "2026-01-01T00:00:00Z", false),
      result("b", "2026-01-01T00:00:00Z", false),
      result("a", "2026-01-01T00:01:00Z", true),
      result("b", "2026-01-01T00:01:00Z", true),
      result("a", "2026-01-01T00:02:00Z", false),
      result("b", "2026-01-01T00:02:00Z", false),
    ],
    host,
  );
  assert.equal(uptimePercent, 66.67);
});
