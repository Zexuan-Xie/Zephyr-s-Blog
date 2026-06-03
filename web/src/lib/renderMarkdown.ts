import DOMPurify from 'dompurify';
import { marked } from 'marked';

marked.use({
  async: false,
  gfm: true,
});

export function renderSafeMarkdown(markdown: string): string {
  const rendered = marked.parse(markdown, { async: false }) as string;
  return DOMPurify.sanitize(rendered, {
    USE_PROFILES: { html: true },
  });
}

export function sanitizeServerHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
  });
}
