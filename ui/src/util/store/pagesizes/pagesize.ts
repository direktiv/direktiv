import { create } from "zustand";
import { persist } from "zustand/middleware";
import { z } from "zod";

export const pageSizeValue = ["10", "20", "30", "50"] as const;
export const PageSizeValueSchema = z.enum(pageSizeValue);
export type PageSizeValueType = z.infer<typeof PageSizeValueSchema>;

const defaultPageSize: PageSizeValueType = "10";

interface PageSizeState {
  pageSize: PageSizeValueType;
  actions: {
    setPageSize: (pageSize: PageSizeValueType) => void;
  };
}

type StoreSlices = {
  [key: string]: PageSizeState;
};

const usePageSizeState = create<StoreSlices>()(
  persist(() => ({}), {
    name: "direktiv-store-page-size",
    partialize: (state) => state,
  })
);

export const initializePageSizeSlice = (storeName: PageSizeNameType) => {
  usePageSizeState.setState((state) => {
    if (!state[storeName]) {
      return {
        ...state,
        [storeName]: {
          pageSize: defaultPageSize,
          actions: {
            setPageSize: (newPageSize: PageSizeValueType) =>
              usePageSizeState.setState((currentState) => ({
                ...currentState,
                [storeName]: {
                  ...currentState[storeName],
                  pageSize: newPageSize,
                },
              })),
          },
        },
      };
    }
    return state;
  });
};

const pageSizeNames = ["events", "eventlisteners"] as const;
export const PageSizeNameSchema = z.enum(pageSizeNames);
export type PageSizeNameType = z.infer<typeof PageSizeNameSchema>;

export const usePageSize = (storeName: PageSizeNameType) => {
  const slice = usePageSizeState((state) => state[storeName]);
  if (!slice) {
    throw new Error(`Store slice "${storeName}" has not been initialized.`);
  }
  return slice.pageSize;
};

export const usePageSizeActions = (storeName: PageSizeNameType) => {
  const slice = usePageSizeState((state) => state[storeName]);
  if (!slice) {
    throw new Error(`Store slice "${storeName}" has not been initialized.`);
  }
  return slice.actions;
};
