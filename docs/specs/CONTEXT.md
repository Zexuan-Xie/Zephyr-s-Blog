# Personal Blog

This context defines the product language for a single-author full-stack personal blog built as a reading-first technical writing and interaction surface.

## Language

**Blog**:
A standalone full-stack writing application for publishing technical files inside a navigable content tree and enabling reader interaction. It is separate from the author's GitHub homepage and does not serve as an academic profile.
_Avoid_: Academic homepage, portfolio site, static homepage

**Author**:
The single person who owns the blog and can publish or manage files. The Author is the only person with admin privileges and is created through deployment configuration, not public registration.
_Avoid_: Writer, owner, maintainer, Admin when referring to the person

**Reader**:
A registered user who can comment on and like published files but cannot publish content.
_Avoid_: Visitor, subscriber, customer

**Anonymous Visitor**:
An unauthenticated person who can browse public content but cannot comment or like.
_Avoid_: Guest user, reader

**Content Tree**:
The Unix-like hierarchy that organizes the blog's public content. It contains directories and files and can grow through nested directories.
_Avoid_: Category list, flat post list, article collection

**Author Workspace**:
The Author-facing creation and management surface for the Content Tree, Files, assets, publication controls, and node settings. The route may be protected by Admin privileges, but product UI should describe the workspace as the Author's workspace rather than an admin console.
_Avoid_: Admin console, Tree Manager, backend panel, node manager

**URL Path**:
The readable address of a Directory or File within the Content Tree, such as `/research/notes`. The Author sees and edits a URL Path; `slug` is an internal implementation term and must not appear in product UI.
_Avoid_: Slug, route key, node URL

**Name**:
The human-readable label shown for a Directory or File. After creation, changing a Name does not change its URL Path.
_Avoid_: Title when referring to a Directory, URL name

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

**Published Content**:
The most recently published Content Version that Readers and Anonymous Visitors can access while its File is published. Later autosaved edits remain private until the Author publishes changes, and unpublishing hides but does not erase this snapshot.
_Avoid_: Current draft, live editor state

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

**Content Version**:
One retained state of a File's body, keywords, and render format. A Content Version does not include the File's tree location, publication state, assets, or reader interactions.
_Avoid_: File backup, node version, site snapshot

**HTML Document**:
A complete author-created HTML interface stored as a file and rendered separately from the main React application.
_Avoid_: HTML snippet, unsafe DOM injection

**Draft Asset**:
A File Asset available only to the Author's Current content and Draft Preview. Uploading it does not make it publicly accessible.
_Avoid_: Private public asset, unpublished download

**Published Asset**:
A File Asset made publicly accessible through an explicit publication of its File changes.
_Avoid_: Uploaded asset, draft attachment

**Glass Ricepaper**:
The light-only visual language for the blog: warm rice-paper base, unified frosted-glass surfaces, and a single blue action color.
_Avoid_: Dark mode, generic glassmorphism, multi-theme UI
