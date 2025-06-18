import { TableCell } from "@/components/ui/table";
import { ValidationRecord } from "@/types"; // Assuming Claim is the type for claim_details
import { formatDate } from "@/utils";

interface ClientCreatedAtCellProps {
  details: ValidationRecord["claim_details"];
  executionClient: string;
  fallbackCreatedAt?: string; // Optional top-level fallback
}

export function ClientCreatedAtCell({
  details,
  executionClient,
  fallbackCreatedAt,
}: ClientCreatedAtCellProps) {
  const clientSpecificCreatedAt = details?.[executionClient]?.created_at;

  const displayDate = clientSpecificCreatedAt
    ? formatDate(clientSpecificCreatedAt)
    : fallbackCreatedAt
    ? formatDate(fallbackCreatedAt) // Use formatted fallback
    : "—"; // Final fallback

  return (
    <TableCell className="px-4 whitespace-nowrap align-middle">
      <div className="text-sm">{displayDate}</div>
    </TableCell>
  );
}
