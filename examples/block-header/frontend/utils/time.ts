/**
 * Time formatting utility functions
 */

/**
 * Format validation time from microseconds to readable format
 */
export const formatValidationTime = (microseconds: number | null): string => {
  if (!microseconds) return "—";

  const ms = microseconds / 1000;
  if (ms < 1000) {
    return `${ms.toFixed(2)} ms`;
  }

  const seconds = ms / 1000;
  if (seconds < 60) {
    return `${seconds.toFixed(2)} s`;
  }

  const minutes = seconds / 60;
  return `${minutes.toFixed(2)} min`;
};

/**
 * Format date string with relative time
 * For times less than an hour ago, returns a relative format like "5 minutes ago"
 * For older times, returns an absolute format like "YYYY-MM-DD HH:MM:SS"
 */
export const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();

  // Calculate time difference
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);

  // Format absolute time: YYYY-MM-DD HH:MM:SS
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  const seconds = String(date.getSeconds()).padStart(2, "0");

  const formattedDate = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;

  // If less than one hour, show relative time
  if (diffHour < 1) {
    if (diffMin > 0) {
      return `${diffMin} minute${diffMin > 1 ? "s" : ""} ago`;
    } else {
      return `${diffSec} second${diffSec !== 1 ? "s" : ""} ago`;
    }
  }
  // Otherwise show absolute time
  else {
    return formattedDate;
  }
};
