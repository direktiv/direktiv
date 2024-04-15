import {
  FC,
  PropsWithChildren,
  createContext,
  useContext,
  useState,
} from "react";

import { StateType } from "./type";

const DisabledStateContext = createContext<StateType | null>(null);

const Provider: FC<PropsWithChildren> = ({ children }) => {
  const state = useState(false);
  return (
    <DisabledStateContext.Provider value={state}>
      {children}
    </DisabledStateContext.Provider>
  );
};

const useReadContextFromProvider = () => {
  const context = useContext(DisabledStateContext);
  if (!context) {
    throw new Error("useDisabled must be used within a DisabledStateContext");
  }
  return context;
};

const useDisabled = () => {
  const context = useReadContextFromProvider();
  return context[0];
};

const useSetDisabled = () => {
  const context = useReadContextFromProvider();
  return context[1];
};

export { Provider as DisabledStateProvider, useDisabled, useSetDisabled };
