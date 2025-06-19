import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { formatBytes } from "@/utils";

interface ValidationRecordClaimCellProps {
  recordId: number;
  claimSize?: number | null;
  onViewData: (
    title: string,
    recordId: number,
    type: "claim" | "proof"
  ) => void;
}

export function ValidationRecordClaimCell({
  recordId,
  claimSize,
  onViewData,
}: ValidationRecordClaimCellProps) {
  const effectiveSize = claimSize;
  const displaySize =
    effectiveSize !== undefined && effectiveSize !== null
      ? formatBytes(effectiveSize)
      : "—";
  const canView = effectiveSize !== undefined && effectiveSize !== null;

  return (
    <TableCell className="px-4 align-middle">
      <div className="flex items-center gap-3">
        <Button
          size="sm"
          onClick={() => onViewData("Claim", recordId, "claim")}
          className="flex-shrink-0 bg-pi2-accent-white text-pi2-accent-black hover:bg-pi2-accent-white/90"
          disabled={!canView}
        >
          View
        </Button>
        <div className="text-sm text-pi2-accent-white font-normal">
          {displaySize}
        </div>
      </div>
    </TableCell>
  );
}
