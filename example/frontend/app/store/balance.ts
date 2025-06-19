import { getBalance } from "@/app/actions/vsl";
import { atom } from "jotai";

/**
 * Primitive atom for the balance value.
 */
export const balanceAtom = atom<number | null>(null);

/**
 * Primitive atom for the fetching state.
 */
export const isFetchingBalanceAtom = atom(false);

/**
 * A write-only atom to fetch the balance and update the state atoms.
 * When called, it also returns the newly fetched balance.
 */
export const fetchBalanceAtom = atom(
  null, // first argument is read, we don't need it so it's null.
  async (get, set, address: `0x${string}`): Promise<number | null> => {
    set(isFetchingBalanceAtom, true);
    try {
      const balanceStr = await getBalance(address);
      const newBalance = Number(balanceStr);
      set(balanceAtom, newBalance);
      return newBalance;
    } catch (error) {
      console.error("Failed to fetch balance", error);
      set(balanceAtom, null); // Set to null on error
      return null;
    } finally {
      set(isFetchingBalanceAtom, false);
    }
  }
);
