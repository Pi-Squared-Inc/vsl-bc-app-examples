import { TableCell } from "@/components/ui/table";

interface CreatedAtCellProps {
  createdAt: string;
}

export function CreatedAtCell({ createdAt }: CreatedAtCellProps) {
  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return timestamp;
    }
  };

  return (
    <TableCell className='px-4 py-2 text-center align-middle text-sm'>
      {formatTimestamp(createdAt)}
    </TableCell>
  );
}
