/**
 * Formatting utilities for consistent data display
 */

/**
 * Formats bytes into human-readable size
 * @example formatBytes(1536) // "1.5 KB"
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/**
 * Truncates ID to first 8 characters for display
 * @example truncateId("abcd1234-5678-90ef") // "abcd1234"
 */
export function truncateId(id: string, length: number = 8): string {
  return id.substring(0, length);
}

/**
 * Formats JSON with proper indentation
 * Falls back to string representation if JSON.stringify fails
 */
export function formatJSON(obj: any): string {
  try {
    return JSON.stringify(obj, null, 2);
  } catch {
    return String(obj);
  }
}

/**
 * Formats body content - handles both string and object bodies
 */
export function formatBody(body: any): string {
  if (typeof body === "string") {
    return body;
  }
  return formatJSON(body);
}
