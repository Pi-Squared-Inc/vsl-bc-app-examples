import { TableHead } from "@/components/ui/table";

interface BlockHeaderRecordsTableHeaderCellProps {
  children: React.ReactNode;
  className?: string;
}

export function BlockHeaderRecordsTableHeaderCell({
  children,
  className = "",
}: BlockHeaderRecordsTableHeaderCellProps) {
  return (
    <TableHead className={`font-medium ${className}`}>{children}</TableHead>
  );
}
