import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import { useMemo } from "react";
import { BlockHeaderRecordRow } from "./BlockHeaderRecordRow";
import { BlockHeaderRecordsTableHeaderCell } from "./BlockHeaderRecordsTableHeaderCell";
import { Pagination } from "../Pagination";
import { BlockHeaderRecord } from "@/types/blockHeader";

const SKELETON_ROWS = 25;
const TABLE_COLUMNS = 5;

interface BlockHeaderRecordsTableProps {
  blockHeaderRecords: BlockHeaderRecord[];
  isLoading: boolean;
  isPaging?: boolean;
  isComputing?: boolean;
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function BlockHeaderRecordsTable({
  blockHeaderRecords,
  isLoading,
  isPaging = false,
  isComputing = false,
  currentPage,
  totalPages,
  onPageChange,
}: BlockHeaderRecordsTableProps) {
  const effectiveTotalPages = useMemo(
    () => (totalPages > 0 ? totalPages : 1),
    [totalPages]
  );

  const skeletonRows = useMemo(() => {
    return Array.from({ length: SKELETON_ROWS }).map((_, index) => (
      <TableRow
        key={`skeleton-${index}`}
        className='animate-pulse h-[53px]'
        style={{ animationDelay: `${index * 50}ms` }}
      >
        {[...Array(TABLE_COLUMNS)].map((_, i) => (
          <TableCell key={i} className='px-4 py-2 text-center align-middle'>
            <div className='flex justify-center'>
              <Skeleton className='h-6 w-24' />
            </div>
          </TableCell>
        ))}
      </TableRow>
    ));
  }, []);

  return (
    <section
      aria-label='Block Header Records data'
      className='records-table-container'
    >
      <ScrollArea className='w-full'>
        <Table className='table-fixed w-full min-w-[1400px]'>
          <TableHeader>
            <TableRow className='hover:bg-transparent'>
              <BlockHeaderRecordsTableHeaderCell className='text-center'>
                Block Number
              </BlockHeaderRecordsTableHeaderCell>
              <BlockHeaderRecordsTableHeaderCell className='text-center'>
                Claim ID
              </BlockHeaderRecordsTableHeaderCell>
              <BlockHeaderRecordsTableHeaderCell className='text-center'>
                Claim
              </BlockHeaderRecordsTableHeaderCell>
              <BlockHeaderRecordsTableHeaderCell className='text-center'>
                Created At
              </BlockHeaderRecordsTableHeaderCell>
              <BlockHeaderRecordsTableHeaderCell className='text-center'>
                Verification Time
              </BlockHeaderRecordsTableHeaderCell>
            </TableRow>
          </TableHeader>
          <TableBody
            className={cn(
              "transition-opacity duration-200",
              (isPaging || isComputing) && "opacity-60"
            )}
          >
            {isLoading && !isPaging && !isComputing ? (
              <>{skeletonRows}</>
            ) : blockHeaderRecords.length > 0 ? (
              blockHeaderRecords.map((record) => (
                <BlockHeaderRecordRow
                  key={record.claim_details.claim_id}
                  blockHeaderRecord={record}
                />
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={TABLE_COLUMNS}
                  className='text-center h-24 px-4'
                >
                  No block header records available
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        <ScrollBar orientation='horizontal' />
      </ScrollArea>
      <div className='mt-6'>
        <Pagination
          currentPage={currentPage}
          totalPages={effectiveTotalPages}
          onPageChange={onPageChange}
          disabled={isLoading || isPaging || isComputing}
        />
      </div>
    </section>
  );
}
