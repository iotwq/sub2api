# Source attribution

Ported from `ranxi2001/sub2api` at commit
`6b0c0ddbd1649caad5d980e92368059b1a5d1158` (2026-09-25 review).
Source: https://github.com/ranxi2001/sub2api/tree/6b0c0ddbd1649caad5d980e92368059b1a5d1158/backend/internal/service/basispoints
The Responses protocol bridge and its tests are included here. On 2026-09-26,
the optional image relay, its tests and admin settings were also ported from
v2.8.12 (ae4bc4dcbb5a3014ed3269e9f726ad70111912bf), including 460b02d3b,
eb069778a and 9e4c33c7e. Admission is local to opted-in Basispoints forwarding,
not the reference global gateway middleware. Account pools and deployment
code are not imported.

On 2026-09-26, Basispoints-related changes from v2.8.14
(e39898c680ecd69381e549ae54c97011107f1757) were ported: FUNCTION_CODE
transport (#87), actual-body image admission and configurable capacity (#83,
#91), opt-in HTTP 403 protocol fallback (#84), tool roundtrip tests (#86),
and usage snapshots / HTTP 429 cooldown (#89). The local account endpoint
setting, scoped image admission, and bounded retry/failover behavior are
retained. Unrelated capture, usage UI and deployment changes are not imported.

On 2026-09-26, standalone Basispoints changes through v2.8.17 (26b324b),
compared with v2.8.14, were ported: #96/#97/#99/#100 tool protocol and
bounded corrections; #98 isolated throttling and replay expiry; #101
compaction threshold; #102 native attachments; #105 opt-in 403 group actions;
#107 image limits; and #110 authentication handling. Local endpoint selection,
credential scoping, accounting and account concurrency remain in force.
The v2.8.14 HTTP 429 cooldown described above is superseded by isolation.
Mihomo managed proxies, subscriptions and session pools are not imported.

This package is adapted from hloolx/codex2api, whose README declares MIT License.
Original author of the Basispoints changes: hloolx. Source: https://github.com/hloolx/codex2api

Requested commits and required intervening fixes, in order:

- 9d02d3f5e5d69632ebb9590082a833c0a0916356 — initial Basispoints routing.
- c125e560eefb5fd15c995943eb1e111795b0635f — tool-loop identity and replay.
- 20ff3e860d9a149e2df731e37ba1d9b56ae053fc — envelope formatting and model access errors.
- d39f7e3697aab342e303bf4be0142e39b6a58515 — tool catalog and complete terminal items.
- 4dea83ec53b7668419edd2a9a9dd40fb55fdaacd — HTTPS image references.

Sub2API supplies its own account setting,
OAuth credential lifecycle, proxy transport, usage recording and frontend.
The repositories have different layouts; this is a source port, not a Git merge
of the other application's deployment, database or account pool implementation.

Additional review reference: JaxsonWang/cpa-plugin-oai-basispoints at
05b2d97efa1bd117da6bd4d362d6e88f8e483680. Its tool/args envelope examples
were compared with this package. We retain scoped caches, incremental text
streaming and multiple-terminal-tool handling rather than its global call-ID
cache and single-transport extraction. No CPA plugin ABI is imported.

On 2026-09-27, standalone BPS changes from v2.8.18 (c1008182bd1bb8f50ff95133fa486fb9d4676811)
and v2.8.19 (08356987417f205cd704bf79180def20fcbbf445) were ported relative to v2.8.17:
#118 encrypted reasoning recovery and route-based agent capabilities; #120/#131 403 markers;
#121/#151 image omission; #122/#139/#156 native screenshots and tool/agent image history;
#124 validated tool-batch preservation; #125 scheduled test isolation; #133 isolated H2 policy;
#134/#140 BPS route policy; #145 safe validation locations; #148 history attribution;
#149 cancellation/duration accounting; and #154 opt-in encrypted message omission.
The local endpoint selector, credential/proxy/concurrency scopes and accounting remain authoritative.
The final #140 policy supersedes #134: BPS-selected requests stay on BPS; unsupported hosted
tool declarations are omitted with an unavailable-capability notice, while forced selections fail.
The unused omission switch is not imported. Probe request counts remain bounded.
Mihomo, managed proxy pools, warm workers and unrelated features are not imported.
After explicit approval on 2026-09-27, the three obsolete replay fields were removed from
the shared error-log INSERT and its arguments. Single/batch writes and cancellation metrics
pass against the currently migrated PostgreSQL schema. No schema workaround, billing change
or production data change is applied; see docs/OPENAI_OAUTH_BASISPOINTS.md.

On 2026-09-28, the standalone BPS 429 fix from ranxi2001/sub2api v2.8.20
(`dc01c71b758e8c24bd76ea5f7ddb089fc76481d3`) was adapted. Main Responses requests,
native attachment uploads and pre-output encrypted-content recovery now fail over through
the existing account switch path on HTTP 429. Post-output tool corrections only cool the
current account's BPS route and are never replayed. The cooldown accepts a validated
Retry-After seconds value or HTTP date, falls back to the existing 429 avoidance setting,
and is isolated from Codex quota, account health, billing and global cooldown state. The
ordinary, load-aware and refreshed account selectors skip BPS-cooled accounts; exhaustion
returns a sanitized 429. The local names use `Basispoints`/`Excel BPS` for compatibility
with existing settings and tests.

The v2.9.0 commits `b3e494dbd` and `211202d4a` add quality-rule automatic BPS enable/disable.
They were reviewed but not copied because this repository has no compatible
`account_quality_bps` repository/service, Pelican quality-plan scheduler, migrations or
admin UI. Mihomo and managed proxy dependencies remain outside this port. Existing manual
Basispoints account selection remains authoritative.

On 2026-09-29, applicable fixes from ranxi2001/sub2api v2.9.1 through
v2.9.4 (release commit 7dd10bfe4b635f226f0ddfa52cc65797697272d8) were adapted:

- f6666ab43: validated native attachment references use only type/file_id.
- e0ba95a48: visible streaming output prevents unknown-tool regeneration.
- b6617bf62: safe in-band failure classification, cancellation terminals, and
  typed failures from tool correction readers.

The failure classification commit cites JaxsonWang/cpa-plugin-oai-basispoints
v0.2.7 (8960a41dacec8ceca2ae2550ce8c9515997cb7a6) as its behavioral reference
for safe codes and explicit status precedence. No CPA ABI or WS transport is
imported here. This port reuses the local Responses/Chat shared parsers,
preserves raw accounting usage and the no-usage/no-record policy, and keeps
actual HTTP status separate from the semantic failure status.

The BPS-only pool partition / initial admission retry (a6e51622a), managed
proxy acquisition health fix (0c74d2d7a), and automatic recovery / quality
operations fixes depend on architecture absent here. Local mixed-pool tests
cover existing capacity fallback without introducing BPS priority semantics.
New image generation, compaction policies, capacity ceilings, default-account
configuration, account-plan restrictions, and managed operational services
are not part of this bug-fix port.
