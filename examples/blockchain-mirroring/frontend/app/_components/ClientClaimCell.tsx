import { TableCell } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { ValidationRecord } from "@/types";
import Link from "next/link";

interface ClientClaimCellProps {
  details: ValidationRecord["claim_details"];
  executionClient: string;
  className?: string;
}

export function ClientClaimCell({
  details,
  executionClient,
  className,
}: ClientClaimCellProps) {
  const clientClaimId = details?.[executionClient]?.claim_id;

  return (
    <TableCell
      className={cn(
        "px-4 whitespace-nowrap align-middle overflow-ellipsis overflow-hidden",
        className
      )}
    >
      {clientClaimId ? (
        <Link
          href={`${process.env.NEXT_PUBLIC_EXPLORER_URL}/claim/${clientClaimId}`}
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline text-pi2-purple-300"
        >
          {clientClaimId}
        </Link>
      ) : (
        "—"
      )}
    </TableCell>
  );
}
