# Personal Blog

This context defines the product language for a single-author full-stack personal blog built as a reading-first technical writing and interaction surface.

## Language

**Blog**:
A standalone full-stack writing application for publishing technical files inside a navigable content tree and enabling reader interaction. It is separate from the author's GitHub homepage and does not serve as an academic profile.
_Avoid_: Academic homepage, portfolio site, static homepage

**Author**:
The single person who owns the blog and can publish or manage files. The author is the only admin and is created through deployment configuration, not public registration.
_Avoid_: Writer, owner, maintainer

**Reader**:
A registered user who can comment on and like published files but cannot publish content.
_Avoid_: Visitor, subscriber, customer

**Anonymous Visitor**:
An unauthenticated person who can browse public content but cannot comment or like.
_Avoid_: Guest user, reader

**Content Tree**:
The Unix-like hierarchy that organizes the blog's public content. It contains directories and files and can grow through nested directories.
_Avoid_: Category list, flat post list, article collection

**Directory**:
A named container in the content tree that can contain files and other directories. Directory names may be Chinese or English.
_Avoid_: Category, folder category, section

**File**:
A publishable content item inside the content tree. A file can be written in Chinese, English, or mixed language and can render as either a Markdown article or a full HTML document.
_Avoid_: Post, article row, language version

**Draft**:
A file that the author can edit in the admin area but anonymous visitors and readers cannot access publicly.
_Avoid_: Private post, unpublished page

**Published File**:
A file that is visible in public navigation, file pages, and search results.
_Avoid_: Live article, public draft

**Comment Thread**:
The conversation area attached to a specific published file. It can contain top-level comments and replies.
_Avoid_: Feedback feed, discussion board

**Comment**:
A top-level reader-authored message in a comment thread.
_Avoid_: Reply, feedback item

**Reply**:
A reader-authored message nested under another comment in the same comment thread.
_Avoid_: Child comment, nested row

**Like**:
A logged-in reader's or admin's toggleable positive reaction to a specific file or comment. Anonymous visitors cannot like content.
_Avoid_: Vote, favorite, reaction, anonymous like

**Hybrid Search**:
The blog's search experience that combines lexical full-text matching with semantic vector similarity over published files.
_Avoid_: ILIKE search, vector-only search, keyword-only search

**Render Format**:
The author-selected format for a file body: a Markdown article rendered inside the blog reading surface, or a full HTML document rendered in an isolated container.
_Avoid_: Editor mode, language version

**HTML Document**:
A complete author-created HTML interface stored as a file and rendered separately from the main React application.
_Avoid_: HTML snippet, unsafe DOM injection

**Glass Ricepaper**:
The light-only visual language for the blog: warm rice-paper base, unified frosted-glass surfaces, and a single blue action color.
_Avoid_: Dark mode, generic glassmorphism, multi-theme UI
