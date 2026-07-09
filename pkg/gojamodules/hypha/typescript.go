package hyphamod

import (
	"github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"
)

// TypeScriptDeclarer makes xgoja gen-dts emit a hypha.d.ts describing the
// PascalCase fluent-builder + result API that goja reflects from the Go
// methods/fields. This is the JS contract; it must track the real Go method
// names in event_builder.go / builders.go.
func (module) TypeScriptModule() *spec.Module {
	return &spec.Module{
		Name:        "hypha",
		Description: "Hypha kernel fluent client (Go-backed opaque builders + results)",
		RawDTS: []string{
			"export interface ValidationResult { Valid: boolean; Errors: string[] }",

			// Result objects (reflected Go fields — PascalCase).
			"export interface Value { Amount: number; Unit: string }",
			"export interface Event {",
			"  ID: string; Ts: number; Actor: string; Verb: string;",
			"  Kind: string | null; Target: string | null; Ref: string | null;",
			"  Audience: string; Topics: string[]; Body: string | null;",
			"  ValueAmount: number | null; ValueUnit: string | null; Redacted: number;",
			"}",
			"export interface Member {",
			"  ID: string; Handle: string; Name: string; Bio: string; Status: string;",
			"  Trust?: TrustSummary; Profile?: Profile;",
			"}",
			"export interface TrustSummary {",
			"  Balance: number; HoursGiven: number; HoursReceived: number;",
			"  InvitesGiven: number; Connections: number;",
			"}",
			"export interface Profile {",
			"  BioLong: string; Links: string[]; Skills: string[];",
			"  OpenToWork: string | null; LocationNote: string; LastActive: number;",
			"}",
			"export interface Webhook { ID: string; URL: string; CreatedAt: number }",
			"export interface MintPATResult { Token: string; ID: string }",
			"export interface Checkpoint {",
			"  Seq: number; UpToID: string; EventCount: number; Hash: string;",
			"  PrevHash: string | null; CreatedAt: number;",
			"}",
			"export interface VerifyResult { OK: boolean; ComputedHash: string; EventCount: number }",

			// Ask (ISO board) results via MCP.
			"export interface Ask {",
			"  AskID: string; Flavor: string; Author: string; Status: string;",
			"  Topics: string[]; ReplyCount: number; LastActivity: number; CreatedAt: number;",
			"  Body?: string;",
			"}",
			"export interface AskReply {",
			"  ID: string; Ts: number; Actor: string; Verb: string; Body: string;",
			"}",
			"export interface AskList { Asks: Ask[] }",
			"export interface AskThread extends Ask { Thread: AskReply[] }",

			// Builders (terminal calls throw an aggregated error on misuse).
			"export interface EventBuilder {",
			"  Kind(k: string): EventBuilder;",
			"  Topic(...t: string[]): EventBuilder;",
			"  To(member: string): EventBuilder;",
			"  Ref(id: string): EventBuilder;",
			"  Audience(a: string): EventBuilder;",
			"  Value(amount: number, unit: string): EventBuilder;",
			"  IdemKey(k: string): EventBuilder;",
			"  Validate(): ValidationResult;",
			"  ToSpec(): Record<string, unknown>;",
			"  Post(): Event;",
			"}",
			"export interface FeedQuery {",
			"  Kind(k: string): FeedQuery; Topic(...t: string[]): FeedQuery;",
			"  Actor(a: string): FeedQuery; Ref(r: string): FeedQuery;",
			"  Since(s: string): FeedQuery; Limit(n: number): FeedQuery;",
			"  List(): Event[];",
			"}",
			"export interface SearchQuery { Limit(n: number): SearchQuery; List(): Event[] }",
			"export interface PATBuilder { Scopes(...s: string[]): PATBuilder; Label(l: string): PATBuilder; Mint(): MintPATResult }",
			"export interface WebhookBuilder { Secret(s: string): WebhookBuilder; Subscribe(): Webhook }",
			"export interface ExportQuery { All(): Export }",
			"export interface Export { Member: Member | null; Events: Event[] }",
			"export interface CheckpointsQuery { Limit(n: number): CheckpointsQuery; List(): Checkpoint[] }",
			"export interface VerifyQuery { UpTo(id: string): VerifyQuery; Hash(h: string): VerifyQuery; Run(): VerifyResult }",

			// The module exports (require("hypha")).
			"export interface Hypha {",
			"  event(body: string): EventBuilder;",
			"  feed(): FeedQuery;",
			"  search(q: string): SearchQuery;",
			"  whois(id: string): Member;",
			"  whoami(): Member;",
			"  members(): Member[];",
			"  kudos(target: string, body: string): Event;",
			"  webhook(url: string): WebhookBuilder;",
			"  webhooks(): Webhook[];",
			"  unsubscribe(id: string): void;",
			"  pats(): PATBuilder;",
			"  export(): ExportQuery;",
			"  checkpoints(): CheckpointsQuery;",
			"  verify(): VerifyQuery;",
			"  views: {",
			"    balance(): any[]; trust(): any; topics(): any[];",
			"    whoCanHelp(topic: string): any[]; gratitude(member?: string): any;",
			"  };",
			"  postISO(body: string, topics: string[]): Ask;",
			"  replyAsk(id: string, body: string): AskReply;",
			"  closeAsk(id: string, helped: string, note: string): void;",
			"  listAsks(flavor: string, status: string): AskList;",
			"  getAsk(id: string): AskThread;",
			"}",
			"const _default: Hypha;",
			"export default _default;",
		},
	}
}
