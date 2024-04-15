import {
  FC,
  PropsWithChildren,
  createContext,
  useContext,
  useState,
} from "react";

import { StateType } from "./type";

const RunableStateContext = createContext<StateType | null>(null);

const Provider: FC<PropsWithChildren> = ({ children }) => {
  const state = useState(true);
  return (
    <RunableStateContext.Provider value={state}>
      {children}
    </RunableStateContext.Provider>
  );
};

const useReadContextFromProvider = () => {
  const context = useContext(RunableStateContext);
  if (!context) {
    throw new Error("useRunable must be used within a RunableStateContext");
  }
  return context;
};

const useRunable = () => {
  const context = useReadContextFromProvider();
  return context[0];
};

const useSetRunable = () => {
  const context = useReadContextFromProvider();
  return context[1];
};

export { Provider as RunableStateProvider, useRunable, useSetRunable };
