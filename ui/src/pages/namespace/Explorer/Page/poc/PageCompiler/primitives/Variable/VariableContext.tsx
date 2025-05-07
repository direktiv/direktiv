import { FC, PropsWithChildren, createContext, useContext } from "react";

import { VariableNamespace } from "../../../schema/primitives/variable";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;

type State = {
  [keys in VariableNamespace]: Record<VariableId, DefinedValue>;
};

const defaultState: State = {
  loop: {},
  form: {},
  query: {},
};

const VariableContext = createContext<State | null>(null);

type VariableContextProviderProps = PropsWithChildren<{ value: State }>;

const VariableContextProvider: FC<VariableContextProviderProps> = ({
  children,
  value,
}) => (
  <VariableContext.Provider value={value}>{children}</VariableContext.Provider>
);

const useVariableContext = () => {
  const context = useContext(VariableContext);
  return context;
};

const useVariables = () => useVariableContext() ?? defaultState;

export { VariableContextProvider, useVariables };
