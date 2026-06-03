# Render full HTML documents in isolation

Files may be full HTML documents rather than only Markdown reading pages. This gives the author room to publish standalone demos and complex interfaces, but those documents must render in a sandboxed iframe instead of being injected into the main React DOM, preserving the blog shell, authentication state, and global styling boundaries.
