Secret Hitler Server
===

Secret Hitler Server is a simple go application that manages the game state for secret hitler.

Introduction
---


Technical Details
---

The server is broken down into a few components, which can all be independently tested.
The idea was to create a log of events that could be applied to the game state.
The event log could be replayed at any time to discover the game state at any point in time.
Subscribers to the log could all independently keep their own state machine.

### Validate

Each player initiated action is first validated against the current game state.
The validation ensures that all incoming player events are allowed according to the rules.

### Apply

Apply takes the validated events, and blindly applies them to the current game state.
It also ensures the correct order of events, assigning each event an incrmenting event identifier.

### Engine

The engine is just another subscriber to events.
It will take the incoming event, and then produce additional events to advance the game state.

### Filter

Before any event, or the game state is sent to players it is filtered.
This is done to ensure that information is guarded while the game is in progress.
