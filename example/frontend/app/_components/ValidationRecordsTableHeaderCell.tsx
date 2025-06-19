import { TableHead } from "@/components/ui/table";
import { cn } from "@/lib/utils";
import React from "react";

interface ValidationRecordsTableHeaderCellProps
  extends React.HTMLAttributes<HTMLTableCellElement> {
  children: React.ReactNode;
}

export function ValidationRecordsTableHeaderCell({
  children,
  className,
  ...props
}: ValidationRecordsTableHeaderCellProps) {
  const baseClasses =
    "px-4 text-pi2-accent-white font-semibold text-base whitespace-nowrap";

  return (
    <TableHead className={cn(baseClasses, className)} {...props}>
      {children}
    </TableHead>
  );
}
