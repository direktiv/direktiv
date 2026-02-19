import {
  Dispatch,
  FC,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
} from "react";

import { DirektivPagesType } from "../../schema";

export type PageCompilerMode = "edit" | "live";

type PageCompilerContextProps = {
  mode: PageCompilerMode;
  page: DirektivPagesType;
  setPage: (page: DirektivPagesType) => void;
  scrollPos: number;
  setScrollPos: Dispatch<SetStateAction<number>>;
};

const PageCompilerContext = createContext<PageCompilerContextProps | null>(
  null
);

type PageCompilerContextProviderProps =
  PropsWithChildren<PageCompilerContextProps>;

export const PageCompilerContextProvider: FC<
  PageCompilerContextProviderProps
> = ({ children, ...value }) => (
  <PageCompilerContext.Provider value={{ ...value }}>
    {children}
  </PageCompilerContext.Provider>
);

export const usePageStateContext = () => {
  const context = useContext(PageCompilerContext);
  if (!context) {
    throw new Error(
      "usePageStateContext must be used within a PageCompilerContext"
    );
  }
  return context;
};

export const usePage = () => {
  const { page } = usePageStateContext();
  return page;
};
