# SDK transport coverage

| Surface | Client support |
| --- | --- |
| REST v1 | `rest` package: all 88 resource operations and generated schema models |
| Personal resources | Finite personal/delegated tokens, identity, account, library, reading/search and Lives |
| Organization resources | Scoped service tokens, membership, translations, Lives, Studio, presentations, assets and exports |
| Conditional writes | Exact ETags and explicit caller-supplied idempotency keys; no automatic retries |
| Live events | Bounded single HTTP polls with versioned snapshot validation |
| Registered OAuth | Separate website-origin client for metadata, consent URLs, code exchange, refresh and revocation |
| Browser CLI authentication | Supported code exchange and logout, separate from resource REST |
| SSH command protocol | Builders, decoders and typed domain helpers with an explicit caller-owned SSH executor |

Command-style HTTP transport is removed. Command fixtures describe SSH syntax
and decoding only. See [REST usage](../rest/README.md) and [SSH framing](README.md).
