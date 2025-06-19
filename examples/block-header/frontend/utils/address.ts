/**
 * Truncates an Ethereum address to a more readable format
 * Example: 0x1234...5678
 */
export function truncateEthAddress(address: string): string {
  if (!address) return "";
  return `${address.slice(0, 6)}...${address.slice(-4)}`;
}
