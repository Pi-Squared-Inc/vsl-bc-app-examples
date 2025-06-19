"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { createBlockHeaderRecordsAPI } from "../actions/blockHeaderRecord";
import { BlockHeaderRecordsTable } from "../_components/blockHeaderSettlement/BlockHeaderRecordsTable";
import { BlockHeaderRecord } from "@/types/blockHeader";

const PAGE_SIZE = 25;

interface BlockHeaderSettlementProps {
  apiBaseUrl?: string;
  defaultChain?: string;
  className?: string;
}

export function BlockHeaderSettlement({
  apiBaseUrl = "",
  defaultChain = "",
  className = "",
}: BlockHeaderSettlementProps) {
  const [blockHeaderRecords, setBlockHeaderRecords] = useState<
    BlockHeaderRecord[]
  >([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isPaging, setIsPaging] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [selectedChain, setSelectedChain] = useState<string>(defaultChain);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const fetchControllerRef = useRef(false);
  const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const blockHeaderAPI = createBlockHeaderRecordsAPI(apiBaseUrl);

  const fetchBlockHeaderRecordsData = useCallback(
    async (
      page: number,
      chain: string = "",
      loadingType: "initial" | "page" | "silent" = "initial"
    ) => {
      if (fetchControllerRef.current) return;
      fetchControllerRef.current = true;

      if (loadingType === "initial") {
        setIsLoading(true);
      } else if (loadingType === "page") {
        setIsPaging(true);
      }

      try {
        const response = await blockHeaderAPI.getBlockHeaderRecords({
          page: page - 1,
          per_page: PAGE_SIZE,
          chain: chain || undefined,
        });

        if (response) {
          const { block_header_records: records = [], total: totalCount } =
            response;
          const calculatedTotalPages = Math.ceil(totalCount / PAGE_SIZE);

          setBlockHeaderRecords(records);
          setTotal(totalCount);
          setTotalPages(calculatedTotalPages);

          if (loadingType === "page") {
            setCurrentPage(page);
          }
        } else {
          throw new Error("Failed to fetch block header records");
        }
      } catch (error) {
        console.error("Failed to fetch block header records:", error);
        if (loadingType !== "silent") {
          setBlockHeaderRecords([]);
          setTotal(0);
          setTotalPages(0);
        }
      } finally {
        setIsLoading(false);
        setIsPaging(false);
        fetchControllerRef.current = false;
      }
    },
    [blockHeaderAPI]
  );

  useEffect(() => {
    if (defaultChain !== selectedChain) {
      setSelectedChain(defaultChain);
    }
  }, [defaultChain]);

  useEffect(() => {
    setCurrentPage(1);
    setTotalPages(0);
    fetchBlockHeaderRecordsData(1, selectedChain, "initial");
  }, [selectedChain]);

  useEffect(() => {
    if (autoRefresh && currentPage === 1) {
      refreshIntervalRef.current = setInterval(() => {
        if (!fetchControllerRef.current) {
          fetchBlockHeaderRecordsData(1, selectedChain, "silent");
        }
      }, 10000);
    } else {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
        refreshIntervalRef.current = null;
      }
    }

    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
      }
    };
  }, [autoRefresh, currentPage, selectedChain, fetchBlockHeaderRecordsData]);

  const handlePageChange = (page: number) => {
    if (
      page >= 1 &&
      page <= totalPages &&
      page !== currentPage &&
      !fetchControllerRef.current
    ) {
      fetchBlockHeaderRecordsData(page, selectedChain, "page");
    }
  };

  return (
    <div className={`w-full ${className}`}>
      <div className='flex items-center justify-between mb-4'>
        <div className='flex items-center gap-4'>
          <div className='flex items-center gap-2'></div>
        </div>
        <div className='flex items-center gap-2'>
          <label className='flex items-center gap-2 text-sm text-muted-foreground cursor-pointer'>
            <input
              type='checkbox'
              checked={autoRefresh}
              onChange={(e) => setAutoRefresh(e.target.checked)}
              disabled={isLoading || isPaging}
              className='w-4 h-4 text-pi2-purple-500 border-gray-300 rounded focus:ring-pi2-purple-500'
            />
            Auto-refresh (10s)
          </label>
          {autoRefresh && currentPage === 1 && (
            <span className='w-2 h-2 bg-green-500 rounded-full animate-pulse'></span>
          )}
        </div>
      </div>

      <BlockHeaderRecordsTable
        blockHeaderRecords={blockHeaderRecords}
        isLoading={isLoading}
        isPaging={isPaging}
        isComputing={false}
        currentPage={currentPage}
        totalPages={totalPages}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
