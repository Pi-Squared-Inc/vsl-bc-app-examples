"use client";

import BitcoinMirroringTable from "@/app/_components/BitcoinMirroringTable";
import { Button } from "@/components/ui/button";
import { ValidationRecord } from "@/types";
import axios from "axios";
import { RefreshCw } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

export default function Bitcoin() {
  const [validationRecords, setValidationRecords] = useState<ValidationRecord[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isPaging, setIsPaging] = useState(false);
  const [currentPage, setCurrentPage] = useState(0);
  const [targetPage, setTargetPage] = useState(0);
  const [totalPages, setTotalPages] = useState(1);
  const [secondsUntilRefresh, setSecondsUntilRefresh] = useState(10);
  const refreshIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const countdownIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const prevDataRef = useRef<ValidationRecord[]>([]);
  const pageDataCacheRef = useRef<Record<number, ValidationRecord[]>>({});

  const fetchValidationRecords = useCallback(
    async (page: number, loadingType: "initial" | "refresh" | "page" = "initial") => {
      if (loadingType === "initial") setIsLoading(true);
      else if (loadingType === "refresh") setIsRefreshing(true);
      else if (loadingType === "page") {
        setIsPaging(true);
        pageDataCacheRef.current[currentPage] = [...validationRecords];
      }

      setSecondsUntilRefresh(10);

      try {
        const response = await axios.get(`${process.env.NEXT_PUBLIC_API_URL}/block_mirroring_btc_records?page=${page}`);

        const newValidationRecords = response.data.records;
        if (loadingType === "refresh") prevDataRef.current = validationRecords;

        if (JSON.stringify(newValidationRecords) !== JSON.stringify(validationRecords)) {
          setValidationRecords(newValidationRecords);
          pageDataCacheRef.current[page] = newValidationRecords;
        }

        const totalCount = response.data.total;
        const pageSize = 25;
        const calculatedTotalPages = Math.ceil(totalCount / pageSize);
        if (calculatedTotalPages !== totalPages) {
          setTotalPages(calculatedTotalPages);
        }

        if (newValidationRecords?.length === 0 && totalCount > 0 && page > 0) {
          const lastPage = calculatedTotalPages - 1;

          if (loadingType === "initial") setIsLoading(true);
          else if (loadingType === "refresh") setIsRefreshing(true);
          else if (loadingType === "page") {
            setIsPaging(true);
          }

          const lastPageResponse = await axios.get(`${process.env.NEXT_PUBLIC_API_URL || "/api"}/claims?page=${lastPage}`);

          const lastPageClaims = lastPageResponse.data.claims;
          setValidationRecords(lastPageClaims);
          pageDataCacheRef.current[lastPage] = lastPageClaims;

          if (loadingType === "page") {
            setCurrentPage(lastPage);
            setTargetPage(lastPage);
          } else if (loadingType === "initial") {
            setCurrentPage(lastPage);
            setTargetPage(lastPage);
          }

          return;
        }

        if (loadingType === "page") {
          setCurrentPage(page);
        } else if (loadingType === "initial" && page !== currentPage) {
          setTargetPage(page);
        }
      } catch (error) {
        console.error("Failed to fetch claims:", error);
        if (loadingType === "page") {
          setCurrentPage(currentPage);
        }
      } finally {
        setIsLoading(false);
        setIsRefreshing(false);
        setIsPaging(false);
      }
    },
    [validationRecords, currentPage, totalPages]
  );

  useEffect(() => {
    fetchValidationRecords(currentPage, "initial");
  }, []);

  useEffect(() => {
    if (isPaging && targetPage !== currentPage) {
      fetchValidationRecords(targetPage, "page");
    }
  }, [targetPage, isPaging, currentPage, fetchValidationRecords]);

  useEffect(() => {
    refreshIntervalRef.current = setInterval(() => {
      if (!isLoading && !isRefreshing && !isPaging) {
        fetchValidationRecords(currentPage, "refresh");
      }
    }, 10000);

    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
        refreshIntervalRef.current = null;
      }
    };
  }, [currentPage, isLoading, isRefreshing, isPaging, fetchValidationRecords]);

  useEffect(() => {
    countdownIntervalRef.current = setInterval(() => {
      setSecondsUntilRefresh((prev) => {
        if (isRefreshing) return 10;
        return prev > 0 ? prev - 1 : 10;
      });
    }, 1000);

    return () => {
      if (countdownIntervalRef.current) {
        clearInterval(countdownIntervalRef.current);
        countdownIntervalRef.current = null;
      }
    };
  }, [isRefreshing]);

  const handlePageChange = (page: number) => {
    const internalPage = page - 1;
    if (internalPage >= 0 && internalPage < totalPages && internalPage !== currentPage) {
      setIsPaging(true);
      setTargetPage(internalPage);

      if (pageDataCacheRef.current[internalPage]) {
        prevDataRef.current = [...validationRecords];
        setValidationRecords(pageDataCacheRef.current[internalPage]);
      } else {
        prevDataRef.current = [...validationRecords];
      }

      window.scrollTo({ top: 0, behavior: "smooth" });
    }
  };

  const handleRefresh = () => {
    fetchValidationRecords(currentPage, "refresh");
  };

  const displayData = isLoading && !isPaging ? [] : isPaging ? pageDataCacheRef.current[targetPage] || prevDataRef.current : validationRecords;

  const displayPage = (isPaging ? targetPage : currentPage) + 1;
  return (
    <div>
      <div className="flex justify-end py-4">
        <div className="flex items-center gap-4">
          <div className="text-sm text-pi2-purple-50">
            Auto-refresh in: <span className="font-medium">{secondsUntilRefresh}s</span>
          </div>
          <Button onClick={handleRefresh} disabled={isLoading || isRefreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoading || isRefreshing ? "animate-spin" : ""}`} />
            Refresh
          </Button>
        </div>
      </div>
      <BitcoinMirroringTable
        claims={displayData}
        isLoading={isLoading}
        isRefreshing={isRefreshing}
        isPaging={isPaging}
        currentPage={displayPage}
        totalPages={totalPages}
        onPageChange={handlePageChange}
      />
    </div>
  );
}
