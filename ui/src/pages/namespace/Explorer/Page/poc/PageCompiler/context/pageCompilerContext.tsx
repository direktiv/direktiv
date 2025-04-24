import {
  Dispatch,
  FC,
  PropsWithChildren,
  SetStateAction,
  createContext,
  useContext,
  useState,
} from "react";

import { DirektivPagesType } from "../../schema";

type State = {
  mode: "preview" | "live";
  page: DirektivPagesType;
  actions: {
    setMode: Dispatch<SetStateAction<State["mode"]>>;
    setPage: Dispatch<SetStateAction<State["page"]>>;
  };
};

const PageCompilerContext = createContext<State | null>(null);

type PageCompilerContextProviderProps = PropsWithChildren<
  Omit<State, "actions">
>;

const PageCompilerContextProvider: FC<PageCompilerContextProviderProps> = ({
  mode: defaultMode,
  page: defaultPage,
  children,
}) => {
  const [mode, setMode] = useState<"preview" | "live">(defaultMode);
  const [page, setPage] = useState<DirektivPagesType>(defaultPage);

  const value: State = {
    mode,
    page,
    actions: { setMode, setPage },
  };

  return (
    <PageCompilerContext.Provider value={value}>
      {children}
    </PageCompilerContext.Provider>
  );
};

const usePageStateContext = () => {
  const context = useContext(PageCompilerContext);
  if (!context) {
    throw new Error(
      "usePageStateContext must be used within a PageCompilerContext"
    );
  }
  return context;
};

const useMode = () => {
  const { mode } = usePageStateContext();
  return mode;
};

const usePage = () => {
  const { page } = usePageStateContext();
  return page;
};

const useActions = () => {
  const { actions } = usePageStateContext();
  return actions;
};

export { PageCompilerContextProvider, useMode, usePage, useActions };
