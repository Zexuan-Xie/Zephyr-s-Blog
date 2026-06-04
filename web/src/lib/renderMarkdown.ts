import DOMPurify from 'dompurify';
import type { Config } from 'dompurify';
import { marked } from 'marked';

marked.use({
  async: false,
  gfm: true,
});

const markdownSanitizeConfig = {
  USE_PROFILES: { html: true },
  FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form'],
} satisfies Config;

export function renderSafeMarkdown(markdown: string): string {
  const rendered = marked.parse(markdown, { async: false }) as string;
  return DOMPurify.sanitize(rendered, markdownSanitizeConfig);
}

export function sanitizeServerHtml(html: string): string {
  return DOMPurify.sanitize(html, markdownSanitizeConfig);
}
