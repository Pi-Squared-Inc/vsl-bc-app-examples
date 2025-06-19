import { TableCell } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { AlertCircle, Loader2 } from "lucide-react";
import { formatValidationTime } from "@/utils";

interface VerificationTimeCellProps {
  verificationTime?: number | null;
}

export function VerificationTimeCell({
  verificationTime,
}: VerificationTimeCellProps) {
  if (verificationTime === null || verificationTime === undefined) {
    return (
      <TableCell className='px-4 py-2 text-center align-middle'>
        <div className='flex items-center justify-center gap-2'>
          <Loader2 className='h-4 w-4 animate-spin text-pi2-purple-500' />
          <span className='text-sm text-muted-foreground'>Pending</span>
        </div>
      </TableCell>
    );
  }

  return (
    <TableCell className='px-4 py-2 text-center align-middle'>
      <span className='text-sm text-green-600 font-medium'>
        {formatValidationTime(verificationTime)}
      </span>
    </TableCell>
  );
}
