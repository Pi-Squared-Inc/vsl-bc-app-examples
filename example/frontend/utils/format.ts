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

export const formatNumberWithCommas = (num: number): string => {
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
};

/**
 * Shortens an Ethereum address for display purposes.
 * @param address The full Ethereum address.
 * @param startChars The number of characters to show at the beginning.
 * @param endChars The number of characters to show at the end.
 * @returns The shortened address string (e.g., "0x1234...5678").
 */
export const shortenAddress = (
  address: string,
  startChars = 6,
  endChars = 4
): string => {
  if (!address) return "";
  const start = address.substring(0, startChars);
  const end = address.substring(address.length - endChars);
  return `${start}...${end}`;
};
