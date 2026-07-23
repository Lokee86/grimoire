## WriteFreely post creation trace rubric

Fixture commit: `8f942b2aed5951aba717268f0f7a597bd487e8e5`

Each trial is scored out of 10 points:

1. **Routes (2 points)** — identifies `routes.go:143` and `routes.go:161` as POST entry points using `handler.All(newPost)`.
2. **Handler flow (2 points)** — traces `posts.go:newPost`, authentication/session selection, silencing/login checks, JSON/form parsing, publishable-content validation, font validation, and collection ownership where applicable.
3. **Datastore call chain (2 points)** — distinguishes `CreateOwnedPost` from direct `CreatePost` and explains how both converge on `database.go:CreatePost`.
4. **Persistence boundary (2 points)** — identifies friendly ID/slug/timestamp preparation, the `INSERT INTO posts (...)`, duplicate-key slug retry, and error propagation.
5. **Response and side effects (2 points)** — identifies response enrichment/HTTP 201 plus conditional federation and email-job scheduling after persistence.

A point is awarded only when the claim is materially correct and supported by a usable file:line citation. Discovery-call counts are the agent's self-reported source search/list/read calls after its required initial context-package read.
