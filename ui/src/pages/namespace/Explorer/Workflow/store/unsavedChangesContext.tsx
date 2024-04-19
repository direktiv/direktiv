import {
  FC,
  PropsWithChildren,
  createContext,
  useContext,
  useState,
} from "react";

import { StateType } from "./type";

const UnsavedChangesStateContext = createContext<StateType | null>(null);

const Provider: FC<PropsWithChildren> = ({ children }) => {
  const state = useState(false);
  return (
    <UnsavedChangesStateContext.Provider value={state}>
      {children}
    </UnsavedChangesStateContext.Provider>
  );
};

const useReadContextFromProvider = () => {
  const context = useContext(UnsavedChangesStateContext);
  if (!context) {
    throw new Error(
      "useUnsavedChanges must be used within a UnsavedChangesStateContext"
    );
  }
  return context;
};

const useUnsavedChanges = () => {
  const context = useReadContextFromProvider();
  return context[0];
};

const useSetUnsavedChanges = () => {
  const context = useReadContextFromProvider();
  return context[1];
};

export {
  Provider as UnsavedChangesStateProvider,
  useUnsavedChanges,
  useSetUnsavedChanges,
};
