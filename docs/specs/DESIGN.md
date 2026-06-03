---
version: alpha
name: glass-ricepaper
description: A light, warm reading interface built on a rice-paper (宣纸) base with a single frosted-glass surface language. Every container — nav, file reading card, code block, buttons, content entry cards — is the same sandblasted glass: more transparent than typical glassmorphism, matte and grainy on its face, but with a strong directional light-refraction along its rounded edges so each panel reads as a sheet of frosted glass peeled up off warm paper. Restrained warm-beige glow behind the glass supplies depth; no cold colors, no decorative gradients on chrome. A single Action Blue (#0066cc) carries every interactive element. Adapted from Apple's product-marketing language (SF Pro, tight display tracking, 17px body, pill CTAs, scale(0.95) press) and re-pointed at long-form technical reading.

colors:
  primary: "#0066cc"
  primary-focus: "#0071e3"
  primary-on-dark: "#2997ff"
  ink: "#26221c"
  body: "#3a342b"
  ink-60: "#6b6256"
  ink-40: "#9a9082"
  on-primary: "#ffffff"
  canvas-paper-1: "#f4ecdc"
  canvas-paper-2: "#efe6d3"
  paper-warm-glow: "#f6e4be"
  paper-glow-2: "#f7ead0"
  glass-fill: "rgba(255, 253, 247, 0.38)"
  glass-fill-button: "rgba(255, 253, 247, 0.42)"
  glass-edge-bright: "rgba(255, 255, 255, 1)"
  glass-edge-dim: "rgba(255, 255, 255, 0.02)"
  glass-top-highlight: "rgba(255, 255, 255, 0.95)"
  hairline-warm: "rgba(120, 98, 64, 0.14)"
  code-surface: "rgba(38, 33, 26, 0.80)"
  code-text: "#e3ddd0"
  code-comment: "#7d756a"
  code-keyword: "#e0a3b8"
  code-func: "#8fd9b6"
  code-string: "#e8c07d"
  shadow-warm: "rgba(120, 98, 64, 0.13)"

typography:
  hero-display:
    fontFamily: "SF Pro Display, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 40px
    fontWeight: 600
    lineHeight: 1.1
    letterSpacing: -0.03em
  display-md:
    fontFamily: "SF Pro Display, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 30px
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: -0.02em
  lead:
    fontFamily: "SF Pro Display, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 21px
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: -0.01em
  heading-sm:
    fontFamily: "SF Pro Display, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 18px
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: -0.02em
  body:
    fontFamily: "SF Pro Text, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 17px
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: -0.01em
  body-strong:
    fontFamily: "SF Pro Text, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 17px
    fontWeight: 600
    lineHeight: 1.5
    letterSpacing: -0.01em
  caption:
    fontFamily: "SF Pro Text, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.006em
  meta:
    fontFamily: "SF Pro Text, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: 0
  code:
    fontFamily: "SF Mono, ui-monospace, 'JetBrains Mono', 'Cascadia Code', monospace"
    fontSize: 13.5px
    fontWeight: 400
    lineHeight: 1.7
    letterSpacing: 0
  label:
    fontFamily: "SF Pro Text, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 12px
    fontWeight: 600
    lineHeight: 1.0
    letterSpacing: 0.08em
  fine-print:
    fontFamily: "SF Pro Text, -apple-system, system-ui, Inter, 'PingFang SC', 'Noto Sans SC', sans-serif"
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: -0.01em

rounded:
  none: 0px
  sm: 8px
  md: 14px
  lg: 18px
  xl: 22px
  pill: 9999px

spacing:
  xxs: 4px
  xs: 8px
  sm: 12px
  md: 16px
  lg: 22px
  xl: 32px
  xxl: 46px
  section: 64px

effects:
  glass-blur: "blur(20px) saturate(125%)"
  glass-blur-button: "blur(12px)"
  glass-blur-code: "blur(6px)"
  edge-ring-width: 1.8px
  top-highlight-width: 1.6px
  grain-opacity-surface: 0.10
  grain-opacity-paper: 0.05
  grain-blend-surface: soft-light
  grain-blend-paper: multiply
  press-transform: "scale(0.95)"
  drop-shadow-panel: "0 16px 44px rgba(120,98,64,0.13), 0 3px 10px rgba(120,98,64,0.07)"

components:
  glass-panel:
    backgroundColor: "{colors.glass-fill}"
    backdropFilter: "{effects.glass-blur}"
    rounded: "{rounded.xl}"
  glass-nav:
    backgroundColor: "{colors.glass-fill}"
    backdropFilter: "{effects.glass-blur}"
    rounded: "{rounded.pill}"
    height: 54px
    padding: 0 22px
  file-reading-card:
    backgroundColor: "{colors.glass-fill}"
    backdropFilter: "{effects.glass-blur}"
    rounded: "{rounded.xl}"
    padding: 46px 44px
  content-entry-card:
    backgroundColor: "{colors.glass-fill}"
    backdropFilter: "{effects.glass-blur}"
    rounded: "{rounded.md}"
    padding: 22px
  directory-sidebar:
    backgroundColor: "{colors.glass-fill}"
    backdropFilter: "{effects.glass-blur}"
    rounded: "{rounded.xl}"
    padding: 22px
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.body}"
    rounded: "{rounded.pill}"
    padding: 10px 18px
  glass-pill-button:
    backgroundColor: "{colors.glass-fill-button}"
    backdropFilter: "{effects.glass-blur-button}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.pill}"
    padding: 10px 18px
  keyword-chip:
    backgroundColor: "rgba(0, 102, 204, 0.09)"
    textColor: "{colors.primary}"
    typography: "{typography.label}"
    rounded: "{rounded.pill}"
    padding: 5px 12px
  code-block:
    backgroundColor: "{colors.code-surface}"
    textColor: "{colors.code-text}"
    typography: "{typography.code}"
    backdropFilter: "{effects.glass-blur-code}"
    rounded: "{rounded.md}"
  text-link:
    backgroundColor: transparent
    textColor: "{colors.primary}"
    typography: "{typography.body}"
  comment-card:
    backgroundColor: "{colors.glass-fill}"
    backdropFilter: "{effects.glass-blur}"
    rounded: "{rounded.md}"
    padding: 16px 18px
---

## Overview

This is a **light, warm, reading-first** interface. The whole page sits on warm rice-paper (`{colors.canvas-paper-1}` → `{colors.canvas-paper-2}`) with a single soft warm-beige light pool behind the content. Over that paper, **every container is one and the same frosted-glass material** — there is no second surface treatment. Nav, file reading body, code block, content entry cards, buttons, comments: all are sheets of the same sandblasted glass.

The glass has three signature traits, in priority order:

1. **A strong directional refractive edge.** A 1.8px gradient ring runs around each rounded corner — bright white at the top-left and bottom-right, near-zero through the middle — simulating light bending through the thickness of a real glass sheet. This edge is the brand. It is the difference between "frosted glass peeled off the page" and "a flat translucent rectangle."
2. **A matte, grainy face.** The glass surface carries an `feTurbulence` noise grain at `{effects.grain-opacity-surface}` opacity with `soft-light` blend — sandblasted/etched, **not** wet or glossy. There is deliberately **no** specular highlight blob on the face.
3. **High transparency.** Fill is `{colors.glass-fill}` (~38% white), low enough that the warm paper and glow read clearly through every panel, calm `saturate(125%)` so colors behind never get loud.

Density is moderate, not Apple-sparse: this is a place to read 8-minute files, browse directories, and leave comments — not a product gallery. Content max-width is ~760px for comfortable long-form measure. The single Action Blue (`{colors.primary}`) is the only interactive color.

**Key characteristics:**
- One frosted-glass material for every container — the surface language is unified, not per-component.
- Strong rounded-edge light refraction + strong inset top highlight = the "peeled glass" signature.
- Matte sandblasted grain on glass faces; **no** gloss/sheen.
- Warm rice-paper base + one restrained warm-beige glow; **no** cold colors, **no** multi-color orbs.
- Single Action Blue (#0066cc) accent, inherited from Apple; warm near-black ink (`#26221c`, not Apple's cold `#1d1d1f`) to sit naturally on paper.
- SF Pro Display tight headlines / SF Pro Text 17px body, with CJK fallback (`PingFang SC` / `Noto Sans SC`) because content is bilingual-leaning Chinese.
- Code blocks are warm charcoal glass (`{colors.code-surface}`), never cold blue-black — they must belong to the same warm world.

## Background & Light

- **Paper base:** vertical gradient `{colors.canvas-paper-1}` → `{colors.canvas-paper-2}`, `background-attachment: fixed`.
- **Warm glow:** soft radial pools of `{colors.paper-warm-glow}` and `{colors.paper-glow-2}` near the top, blurred ~40px, opacity ≤ 0.45. Exactly one dominant pool; keep it calm.
- **Paper grain:** full-page `feTurbulence` noise at `{effects.grain-opacity-paper}` with `multiply` blend, to read as rice-paper fiber.
- **No decorative CSS gradients on chrome, no cold hues, no second light color.** Depth comes from glass over paper, not from color.

## The Frosted-Glass Primitive

Every container uses this recipe. Treat it as one shared class.

```css
.glass {
  position: relative;
  background: rgba(255, 253, 247, 0.38);              /* {colors.glass-fill} */
  backdrop-filter: blur(20px) saturate(125%);          /* {effects.glass-blur} */
  -webkit-backdrop-filter: blur(20px) saturate(125%);
  border-radius: 22px;                                 /* {rounded.xl} */
  box-shadow:
    0 16px 44px rgba(120,98,64,0.13),                  /* warm drop, lifts off paper */
    0 3px 10px rgba(120,98,64,0.07),
    inset 0 1.6px 0 rgba(255,255,255,0.95),            /* STRONG top-edge highlight */
    inset 0 -1px 0 rgba(255,255,255,0.35);
}
/* refractive edge — directional gradient ring via mask-composite */
.glass::before {
  content: ""; position: absolute; inset: 0; border-radius: inherit; padding: 1.8px;
  background: linear-gradient(135deg,
    rgba(255,255,255,1) 0%, rgba(255,255,255,0.55) 18%,
    rgba(255,255,255,0.02) 48%, rgba(255,255,255,0.40) 78%, rgba(255,255,255,1) 100%);
  -webkit-mask: linear-gradient(#000 0 0) content-box, linear-gradient(#000 0 0);
  -webkit-mask-composite: xor; mask-composite: exclude; pointer-events: none;
}
/* matte sandblasted grain on the glass face (NOT a glossy sheen) */
.glass::after {
  content: ""; position: absolute; inset: 0; border-radius: inherit; pointer-events: none;
  background-image: var(--grain); background-size: 180px 180px;
  opacity: 0.10; mix-blend-mode: soft-light;
}
.glass > * { position: relative; z-index: 1; }
```

`--grain` is a shared `feTurbulence` data-URI (`baseFrequency ~0.85`, `numOctaves 2`, `stitchTiles=stitch`).

## Colors

- **Action Blue** (`{colors.primary}` #0066cc): the single interactive color — links, primary CTA fill, focus ring, keyword chip text. No second accent exists.
- **Warm ink** (`{colors.ink}` #26221c): headlines, nav brand, strong text. Chosen over Apple's cold #1d1d1f so text belongs to the warm-paper world.
- **Body** (`{colors.body}` #3a342b): long-form paragraph text — a touch softer than ink for reading comfort.
- **Muted** (`{colors.ink-60}` / `{colors.ink-40}`): secondary copy and meta/date lines.
- **Paper** (`{colors.canvas-paper-1}` / `-2`): the page base, never used as a panel fill (panels are translucent glass, not opaque paper).
- **Code surface** (`{colors.code-surface}` warm charcoal): the only "dark" surface; it is warm, never blue-black.

## Typography

Inherits Apple's voice: **SF Pro Display 600 with negative tracking** for headlines, **SF Pro Text 400 at 17px** for body. Two deltas for a reading blog:

- **Body line-height is 1.6**, not Apple's 1.47 — longer-form reading over glass wants more air.
- **CJK fallback is mandatory** (`PingFang SC`, then `Noto Sans SC`) since content leans Chinese. SF Pro does not render CJK; the fallback carries it. On non-Apple platforms `Inter` covers Latin.

See the `typography:` tokens for the full ladder (hero-display 40 / display-md 30 / lead 21 / heading-sm 18 / body 17 / caption 14 / meta 14 / code 13.5 / label 12 / fine-print 12). Weight ladder is 400 / 600 only; 500 absent.

## Shapes

| Token | Value | Use |
|---|---|---|
| `{rounded.sm}` | 8px | inline thumbnails inside cards |
| `{rounded.md}` | 14px | code blocks, content entry cards, comment cards |
| `{rounded.lg}` | 18px | detail chips |
| `{rounded.xl}` | 22px | the main file reading panel, large glass panels |
| `{rounded.pill}` | 9999px | nav bar, all buttons, keyword chip |

No square-cornered glass — the refractive edge needs a radius to read. Pills carry every "action."

## Elevation & Depth

There is exactly one elevation idea: **glass floating over paper.** Lift comes from the warm drop-shadow + the bright refractive edge + top highlight, never from borders or hard lines. Do not stack glass on glass more than one level deep (a panel on paper, with at most flat content inside) — nested frosted layers muddy the blur. The code block is the one nested dark surface and is allowed.

## Components

- **`glass-nav`** — sticky pill at top, frosted glass, brand left + quiet links right (active link in `{colors.ink}` 600, rest in `{colors.ink-60}`). Includes a **ZH / EN UI language toggle**: render as a small pill segment or text toggle in the right cluster, active locale in `{colors.ink}` 600, inactive in `{colors.ink-40}`. Switching locale changes UI chrome only; Directory/File content remains exactly as authored.
- **`file-reading-card`** — the Markdown file reading panel. Stack: `keyword-chip` keywords (max 3) → `hero-display` title → `meta` path/time line → `lead`/body paragraphs / `code-block` → interaction bar (`button-primary` like + `glass-pill-button` comment/share).
- **`content-entry-card`** — floating glass card for one next-level content tree entry, either Directory or File. Stack: small `label` (`DIRECTORY` / `FILE`, or localized equivalent) → `heading-sm` display `name` → `caption` path/keywords/meta. Directory caption shows child directory/file counts; File caption shows path plus updated time / read time / weak render meta when useful. Markdown vs HTML is not a public category; it only determines the file renderer after opening. Directory cards enter the next directory; File cards open the file. Use the same glass recipe and avoid nested glass inside the card.
- **`directory-sidebar`** — callable directory tree drawer. It overlays the page instead of pushing content. Desktop width ~320px; mobile width `min(88vw, 360px)`. It is a single frosted glass sheet over paper with a warm lightweight scrim. Close via scrim, Esc, or close pill. Use quiet indentation, `caption` text for nested paths, Action Blue only for active/current path and primary controls.
- **`code-block`** — warm charcoal glass `{colors.code-surface}`, a 3-dot title bar tinted to the syntax palette, mono `{typography.code}`. Syntax: keyword `{colors.code-keyword}`, function `{colors.code-func}`, string `{colors.code-string}`, comment `{colors.code-comment}`.
- **`button-primary`** — solid Action Blue pill, white text, `scale(0.95)` on press.
- **`glass-pill-button`** — frosted glass pill for secondary actions; same press transform.
- **`keyword-chip`** — Action Blue on 9% blue tint, `{typography.label}`. Used for up to three public File keywords, not for categories.
- **`comment-card`** — frosted glass `{rounded.md}`, author in `body-strong`, timestamp in `meta`, body in `body`. (Extension — not yet visually prototyped; follows the same glass recipe.)
- **`text-link`** — Action Blue, optional 1px 30%-alpha blue underline.

## Do's and Don'ts

### Do
- Use the one `.glass` recipe for every container — unify the surface.
- Keep the refractive edge **strong** (1.8px ring + 1.6px inset top highlight). It is the signature.
- Keep glass **transparent** (~38%) and the grain **matte** (`soft-light`, ~0.10). Sandblasted, not wet.
- Keep the base warm rice-paper with **one** restrained warm-beige glow.
- Use Action Blue for every interactive element and nothing else.
- Run body at 17px / 1.6, headlines in SF Pro Display 600 with negative tracking.
- Use `scale(0.95)` as the press state on every button.
- Always include the CJK font fallback in every text style.

### Don't
- Don't add a glossy specular sheen to glass — roughness (grain) replaces gloss here.
- Don't make glass opaque or near-white — transparency is the point; you must see paper through it.
- Don't introduce cold colors (blue/pink/green orbs, blue-black code) — the world is warm.
- Don't add a second accent color, or shadows on text/inline elements.
- Don't square the corners of glass, and don't nest frosted glass more than one level (code block excepted).
- Don't drop body below 17px or line-height below 1.6 for Markdown file prose.
- Don't use opaque borders as dividers — let the refractive edge and surface do the separating.

## Responsive Behavior

- Content max-width ~760px; full-width with 20px gutters below ~800px.
- Article card padding tightens from `46px 44px` → ~28px on phones.
- Index grid: 2-col → 1-col below ~640px.
- `hero-display` 40px → 30px below ~640px.
- Touch targets ≥ 44px; pills already clear this.
- `backdrop-filter` fallback: if unsupported, raise `glass-fill` alpha to ~0.85 (near-opaque warm white) so text stays legible — the only graceful degradation.

## Iteration Guide

1. Change one component at a time; reference its YAML key (`{components.code-block}`).
2. Never inline hex — use `{colors.*}` / `{rounded.*}` / `{spacing.*}` / `{effects.*}` refs.
3. The frosted-glass recipe is shared; tune it in one place, not per component.
4. Default and Active(`scale(0.95)`)/Focus states only — do not document hover.
5. The refractive edge + matte grain + warm paper are the non-negotiable identity. When in doubt, strengthen the edge before adding any chrome.

## Known Gaps (to design later)

- **Comment thread** layout is specified as a glass recipe but not yet visually prototyped (nesting, reply indentation, input box).
- **Admin / Markdown editor** UI is undesigned.
- **Form & validation states** (login, register, errors) not yet surfaced.
- **Dark mode** is intentionally **out of scope** — this is a light-only, rice-paper system by decision.
- Empty/loading/skeleton states for the content tree card grid and directory sidebar not yet designed.

> Reference prototype: `docs/design/glass-light-v2.html` (the approved v2 mockup these tokens are derived from — committed to the repo as the visual source of truth).
