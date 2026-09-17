# 02 — Skill attachment: per-turn loaded-skill injection

**What to build:** A user message can carry a skill attachment as a first-class content part. When such a message is converted for the model, each attached skill arrives as an explicit loaded-skill block (name, description, location, full instructions) grouped under a single system note stating the user attached these skills to this message and the model should follow their instructions; skill blocks are grouped separately from attached-file blocks. Because prompt assembly runs over the whole history each turn, skills attached in earlier turns keep being injected for the life of the conversation. The attachment survives serialize/deserialize and the client/backend message boundary, and sessions saved before this change still load.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] The attachment and the persisted content part carry an optional skill descriptor (name, description, location, instructions); absent by default so old sessions load unchanged
- [ ] The raw skill file content is still carried as the attachment's data for display; the descriptor drives prompt injection
- [ ] A user message with one or more skill attachments produces loaded-skill blocks plus exactly one system note, grouped apart from file-attachment blocks, in the prompt text
- [ ] Skill attachments are treated as text for prompt assembly and are excluded from the binary file-part conversion used for images
- [ ] A message with skill attachments survives the content-parts JSON round-trip with the skill field intact
- [ ] The skill field crosses the client/server message boundary in both directions
- [ ] Tests at the message-layer seam assert the exact injected prompt text and both round-trips
