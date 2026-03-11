import AnsiUp from 'ansi-to-html'

const ansiUp = new AnsiUp({
  newline: false,
  escapeXML: true,
  stream: false
})

export function ansiToHtml(ansi: string): string {
  if (!ansi) return ''
  return ansiUp.toHtml(ansi)
}

export function highlightHtml(html: string, keyword: string): string {
  if (!keyword.trim()) return html
  
  const escaped = keyword.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  // Match keyword but not inside HTML tags
  const regex = new RegExp(`(${escaped})(?![^<]*>)`, 'gi')
  return html.replace(regex, '<mark class="bg-yellow-300 text-black">$1</mark>')
}
