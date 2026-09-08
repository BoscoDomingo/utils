# Before and After Examples

## Part 1: Public STE examples

These examples illustrate ASD-STE100 principles. They paraphrase public
secondary sources. They are not quotations from the official standard.

| Rule | Before | After | Why |
|---|---|---|---|
| One meaning per word | "Verify the system." / "Check the connections." / "Confirm receipt." | "Make sure the system is correct." / "Make sure the connections are correct." / "Make sure you received the item." | Three near-synonyms can make the reader guess whether the actions are the same. One phrase keeps each action distinct. |
| One part of speech per word | "Oil the valve." | "Apply oil to the valve." | If `oil` is approved only as a noun, using it as a verb breaks the one-word, one-role rule. |
| Precise verb meaning | "Follow the safety instructions." | "Obey the safety instructions." | "Follow" can mean "come after" or "obey". STE selects the unambiguous term. |
| Simple tense only | "We have received the technical reports from HQ." | "We received the technical reports from HQ." | Present perfect can add an unnecessary second interpretation. |

## Part 2: User-supplied text examples

These examples show how to rewrite text that a user supplies for rewriting. They
use tool descriptions and inter-agent instructions as data. They do not direct
the agent to transform live tool calls or internal traffic.

### Example A: Tool description

**Before:**

> This tool will attempt to synchronise state across the various backends that
> have been configured, and if a conflict is detected it may resolve it
> automatically depending on the strategy that has been set, or otherwise it
> will surface the conflict for manual review.

**Violations:**

- Two instructions appear in one sentence.
- Compound tenses and modal verbs add ambiguity.
- The sentence is much longer than the descriptive sentence limit.

**After:**

> The tool attempts to synchronise state across the configured backends. If it
> finds a conflict, it checks the current strategy. If the strategy allows
> automatic resolution, the tool may resolve the conflict. If not, the tool
> reports the conflict for manual review.

### Example B: Error message

**Before:**

> An error may have occurred while processing your request due to a possible
> mismatch in the expected data format, which could be caused by an outdated
> client version.

**Violations:**

- The sentence does not identify the actor.
- Multiple hedges in one sentence add ambiguity.
- One sentence carries two separate claims.

**After:**

> An error may result during request processing when the data format does not
> match the expected format. An outdated client version may cause this mismatch.

The rewrite keeps the original uncertainty and causal link. It uses simple
tenses. It does not state that the request failed or that the client caused the
mismatch.

### Example C: Inter-agent instruction

**Before:**

> Once the upstream job has completed and assuming no errors were raised, the
> downstream agent should proceed to consume the output artifact, though it is
> worth noting that partial artifacts are sometimes produced under timeout
> conditions.

**Violations:**

- Present perfect and nested clauses hide the conditions.
- One sentence carries three separate facts.
- The sentence is longer than the instruction limit.

**After:**

> Wait for the upstream job to finish with no errors. Then consume the output
> artifact. Warning: a timeout can produce a partial artifact.

## How to read these examples

The examples apply one meaning per word, active voice, simple tenses, one
instruction per sentence, explicit conditions, and short sentences. These
principles improve machine-to-machine communication and cross-language clarity.
