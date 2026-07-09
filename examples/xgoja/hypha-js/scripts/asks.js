// scripts/asks.js — exercise the ISO board through require("hypha") MCP tools.
// Flow: list open ISOs -> reply to the first open one -> post a new ISO ->
// reply to the new ISO -> close the new ISO.
const hypha = require("hypha");

// 1. List open ISOs.
const list = hypha.listAsks("iso", "open");
console.log("open ISOs:", list.Asks.length);
list.Asks.forEach((a) => console.log("  -", a.AskID, a.Topics.join(","), a.Status));

// 2. Reply to the first open ISO (not our own if possible).
const target = list.Asks.find((a) => a.Author !== "01KX0XTBECGBK8CNK209Y09V4W") || list.Asks[0];
if (target) {
  const r = hypha.replyAsk(target.AskID, "Following — I ran into the same local-replace step while building an xgoja provider for an unpublished consumer. Happy to compare notes.");
  console.log("replied to", target.AskID, "=>", r.ID);
}

// 3. Post a new ISO.
const iso = hypha.postISO("ISO (via xgoja hypha module): need a reviewer for a go-go-goja fluent-builder module — Go-backed opaque objects that accumulate validation errors. 20m review, repo go-go-golems/hypha-cli.", ["go-go-goja", "xgoja", "go"]);
console.log("posted ISO:", iso.AskID, iso.Status);

// 4. Reply to the new ISO (self-reply is allowed; no email is sent).
const r2 = hypha.replyAsk(iso.AskID, "Update: the module is at pkg/gojamodules/hypha; the builders are PascalCase (goja reflects Go method names). Feedback on the validation-aggregation ergonomics welcome.");
console.log("replied to own ISO", iso.AskID, "=>", r2.ID);

// 5. Close the new ISO.
hypha.closeAsk(iso.AskID, "", "Resolved internally; closing to keep the board clean.");
console.log("closed ISO:", iso.AskID);
