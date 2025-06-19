"use client";

import { Button } from "@/components/ui/button";
import { TableCell } from "@/components/ui/table";
import { VerificationRecord } from "@/types";

interface ValidationRecordResultCellProps {
  record: VerificationRecord;
  onViewError: (title: string, content: string) => void;
}

export function ValidationRecordResultCell({ record, onViewError }: ValidationRecordResultCellProps) {
  const hasError = !!record.status && record.status.toLowerCase() === "error";

  const handleShowError = () => {
    if (hasError) {
      onViewError("Validation Error", record.error ?? "No error data available");
    }
  };

  const handleShowResultData = () => {
    const resultData = record.result ?? "No result data available";
    onViewError("Result", resultData);
  };

  return (
    <TableCell className="px-4 align-middle">
      {hasError ? (
        <Button variant="destructive" size="sm" onClick={handleShowError}>
          View Error
        </Button>
      ) : (
        <Button size="sm" onClick={handleShowResultData} className="flex-shrink-0 bg-pi2-accent-white text-pi2-accent-black hover:bg-pi2-accent-white/90">
          View
        </Button>
      )}
    </TableCell>
  );
}
