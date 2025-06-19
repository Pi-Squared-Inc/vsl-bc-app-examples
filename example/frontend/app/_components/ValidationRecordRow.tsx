import { TableCell, TableRow } from "@/components/ui/table";
import { VerificationRecord } from "@/types";
import Link from "next/link";
import { ValidationRecordCreatedAtCell } from "./ValidationRecordCreatedAtCell";
import { ValidationRecordResultCell } from "./ValidationRecordResultCell";

interface ValidationRecordRowProps {
  record: VerificationRecord;
  // onViewData: (title: string, recordId: number, contentType: "claim" | "proof") => void;
  onViewError: (title: string, content: string) => void;
}

export function ValidationRecordRow({
  record,
  //  onViewData,
  onViewError,
}: ValidationRecordRowProps) {
  return (
    <TableRow
      id={`record-${record.id}`}
      suppressHydrationWarning
      className="transition-colors duration-150 hover:bg-white/15 h-[53px]"
    >
      <ValidationRecordCreatedAtCell createdAt={record.created_at} />
      <TableCell className="px-4 whitespace-nowrap align-middle">
        <div className="text-sm">{record.type}</div>
      </TableCell>
      <TableCell className="px-4 whitespace-nowrap align-middle">
        <div className="text-sm">{record.status}</div>
      </TableCell>
      <TableCell className="px-4 whitespace-nowrap align-middle overflow-hidden text-ellipsis">
        {record.claim_id ? (
          <Link
            href={`https://vsl.pi2.network/claim/${record.claim_id}`}
            className="text-sm underline text-purple-500 font-medium"
            target="_blank"
            rel="noopener noreferrer"
          >
            {record.claim_id}
          </Link>
        ) : (
          <span className="text-sm">—</span>
        )}
      </TableCell>
      <ValidationRecordResultCell record={record} onViewError={onViewError} />
    </TableRow>
  );
}
