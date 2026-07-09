---
SectionType: GeneralTopic
Slug: hypha-runtime-api
Title: Hypha JS Module API
Topics:
  - hypha
  - go-go-goja
  - xgoja
---

# Hypha JS Module API

`require("hypha")` exposes the Hypha kernel as a fluent builder API backed by
Go-side opaque objects. Factory functions return Go pointers; goja reflects
their exported methods (PascalCase) and fields into JS.

## Connection

Config (from `xgoja.yaml` `runtime.modules[].config`): `baseUrl`, `pat`,
`handle`. Env fallback: `HYPHA_BASE_URL`, `HYPHA_PAT`, `HYPHA_HANDLE`.

## Builders (terminals throw an aggregated error on misuse)

- `hypha.event(body)` → `EventBuilder`
  - `.Kind(k)`, `.Topic(...t)`, `.To(member)`, `.Ref(id)`, `.Audience(a)`,
    `.Value(amount, unit)`, `.IdemKey(k)`
  - `.Validate()` → `ValidationResult{Valid, Errors}`
  - `.ToSpec()` → plain object (for `JSON.stringify`)
  - `.Post()` → `Event`
- `hypha.feed()` → `FeedQuery`: `.Kind/.Topic/.Actor/.Ref/.Since/.Limit`, `.List()` → `Event[]`
- `hypha.search(q)` → `SearchQuery`: `.Limit`, `.List()` → `Event[]`
- `hypha.pats()` → `PATBuilder`: `.Scopes(...)`, `.Label(...)`, `.Mint()` → `MintPATResult`
- `hypha.webhook(url)` → `WebhookBuilder`: `.Secret(...)`, `.Subscribe()` → `Webhook`
- `hypha.export()` → `ExportQuery`: `.All()` → `Export`
- `hypha.checkpoints()` → `CheckpointsQuery`: `.Limit`, `.List()` → `Checkpoint[]`
- `hypha.verify()` → `VerifyQuery`: `.UpTo(id)`, `.Hash(h)`, `.Run()` → `VerifyResult`

## Terminals (return Go-backed results)

- `hypha.whois(id)` → `Member`, `hypha.whoami()` → `Member`,
  `hypha.members()` → `Member[]`
- `hypha.kudos(target, body)` → `Event`
- `hypha.webhooks()` → `Webhook[]`, `hypha.unsubscribe(id)`
- `hypha.views.balance()` → `BalanceRow[]`
- `hypha.views.trust()` → `TrustResult`
- `hypha.views.topics()` → `TopicRow[]`
- `hypha.views.whoCanHelp(topic)` → `WhoCanHelpRow[]`
- `hypha.views.gratitude(member?)` → `GratitudeResult`

## Result fields (PascalCase, reflected from Go)

`Event.ID`, `.Ts`, `.Actor`, `.Verb`, `.Kind`, `.Target`, `.Ref`,
`.Audience`, `.Topics`, `.Body`, `.ValueAmount`, `.ValueUnit`, `.Redacted`.
`Member.ID`, `.Handle`, `.Name`, `.Trust.Balance`, ...
`Checkpoint.Seq`, `.UpToID`, `.Hash`, ...

## Example

```js
const hypha = require("hypha");
const ev = hypha.event("hello circle").Kind("update").Topic("meta").Post();
console.log(ev.ID, ev.Verb);
const feed = hypha.feed().Topic("meta").Limit(20).List();
```
