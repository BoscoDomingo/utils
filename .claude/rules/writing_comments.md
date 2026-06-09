
When making changes, document at two levels:

1. **Code comments**: Non-obvious logic, shortcuts taken and the reasons for them, or decisions that can't be inferred from the code. Implementation details.
   1. You can also use `MARK:` indicators (and `MARK: -` for separator lines) to clearly delineate sections of code, but these should be used sparingly and for high-level segregation (`MARK: - Test setup` and `MARK: Test execution` are good examples). They show up in IDE minimaps; they are meant for developers to quickly navigate to the relevant code.
2. **Separate doc files**: When the change affects behaviour, usage, setup, architecture, or conventions already documented in `docs/`, `README.md`, or `AGENTS.md`. Anything larger in scope than a couple lines of code should be documented this way.


If uncertain whether documentation is needed, add it but include a note so it is reviewed for removal later.