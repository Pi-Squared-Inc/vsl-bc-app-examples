import { TableCell } from "@/components/ui/table";
import { formatValidationTime } from "@/utils";
import { CheckCircle2 } from "lucide-react";

interface ValidationRecordValidationTimeCellProps {
  validationError?: string | null;
  validationTime?: number | null;
}

export function ValidationRecordValidationTimeCell({
  validationError,
  validationTime,
}: ValidationRecordValidationTimeCellProps) {
  return (
    <TableCell className="px-4 align-middle">
      {validationError ? (
        <span>—</span>
      ) : validationTime !== undefined && validationTime !== null ? (
        <div className="flex items-center text-pi2-green-success">
          <CheckCircle2 className="h-4 w-4 mr-1 flex-shrink-0" />
          <span>{formatValidationTime(validationTime)}</span>
        </div>
      ) : (
        <span>—</span>
      )}
    </TableCell>
  );
}
