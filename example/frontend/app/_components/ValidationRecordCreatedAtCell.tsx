import { TableCell } from "@/components/ui/table";
import { formatDate } from "@/utils";

interface ValidationRecordCreatedAtCellProps {
  createdAt?: string;
}

export function ValidationRecordCreatedAtCell({
  createdAt,
}: ValidationRecordCreatedAtCellProps) {
  const displayDate = createdAt ? formatDate(createdAt) : "—";

  return (
    <TableCell className="px-4 whitespace-nowrap align-middle">
      <div className="text-sm">{displayDate}</div>
    </TableCell>
  );
}
