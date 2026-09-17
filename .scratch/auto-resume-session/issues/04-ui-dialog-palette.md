# 04: UI dialog, command palette entry, and status message

**What to build:** Add an "Auto Resume Session" entry to the command palette (Ctrl+P) with Ctrl+A keyboard shortcut. When selected, opens a config dialog with two controls: Resume Model (toggle between "Small" and "Large", default Large) and Context Threshold (integer input/slider 0–100, default 100). Dialog changes write immediately to global config via the ConfigStore typed mutators (no "Save" button — changes apply on close). When the coordinator triggers auto-compression before a message, the UI shows a brief single-line status message "Auto-resuming…" that replaces itself once the prompt goes through.

**Blocked by:** 03 (coordinator pre-send check — the status message needs the trigger to exist)

**Status:** ready-for-agent

- [ ] "Auto Resume Session" entry added to command palette
- [ ] Ctrl+A shortcut bound to open the dialog
- [ ] Config dialog with Resume Model selector (small/large toggle)
- [ ] Config dialog with Context Threshold input/slider (0–100)
- [ ] Dialog changes persist via ConfigStore typed mutators (global only)
- [ ] Changes take effect immediately on dialog close
- [ ] "Auto-resuming…" status message shown during compression
- [ ] Status message disappears once the prompt is sent
- [ ] Dialog opens correctly from palette entry