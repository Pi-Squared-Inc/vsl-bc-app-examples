/**
 * Formatting utility functions
 */

/**
 * Format bytes to human-readable string with appropriate units
 *
 * @param bytes - Number of bytes to format
 * @param decimals - Number of decimal places to display
 * @returns Formatted string with appropriate unit (Bytes, KB, MB, etc.)
 */
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
}
