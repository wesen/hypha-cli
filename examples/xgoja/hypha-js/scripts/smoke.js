// scripts/smoke.js — smoke test for the generated hypha-js binary.
// Run: dist/hypha-js run examples/xgoja/hypha-js/scripts/smoke.js
// Env: HYPHA_BASE_URL, HYPHA_PAT (or set in xgoja.yaml runtime.modules[].config).
const hypha = require("hypha");

const ev = hypha.event("hypha-js smoke (will redact)")
  .Kind("acme.smoke")
  .Topic("meta")
  .Post();
console.log("posted:", ev.ID);

hypha.checkpoints().Limit(1).List().forEach((cp) => {
  console.log("checkpoint:", cp.Seq, cp.Hash);
});

hypha.feed().Limit(3).List().forEach((e) => {
  console.log("feed:", e.ID, e.Verb);
});
