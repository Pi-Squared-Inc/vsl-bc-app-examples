import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { ValidationRecord } from "@/types";
import { formatValidationTime } from "@/utils";
import { CheckCircle2 } from "lucide-react";

interface ClientValidationCellProps {
  details: ValidationRecord["claim_details"];
  executionClient: string;
  onViewError: (title: string, content: string, executionClient: string) => void;
}

export function ClientValidationCell({ details, executionClient, onViewError }: ClientValidationCellProps) {
  const clientError = details?.[executionClient]?.error;
  const clientTime = details?.[executionClient]?.verification_time;

  const validationError = clientError;
  const validationTime = clientTime;

  return (
    <TableCell className="px-4 align-middle">
      {validationError ? (
        <Button variant="destructive" size="sm" onClick={() => onViewError("Verification Error", validationError || "", executionClient)}>
          View Error
        </Button>
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
