import { PropsWithChildren, createContext, useContext } from "react";

import { VariableNamespace } from "../../../schema/primitives/variable";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;

export type Variables = {
  [keys in VariableNamespace]: Record<VariableId, DefinedValue>;
};

const defaultVariables: Variables = {
  loop: {},
  form: {},
  query: {},
};

const VariableContext = createContext<Variables | null>(null);

type VariableContextProviderProps = PropsWithChildren<{ variables: Variables }>;

const VariableContextProvider = ({
  children,
  variables,
}: VariableContextProviderProps) => (
  <VariableContext.Provider value={variables}>
    {children}
  </VariableContext.Provider>
);

const useVariableContext = () => {
  const context = useContext(VariableContext);
  return context;
};

const useVariables = () => useVariableContext() ?? defaultVariables;

export { VariableContextProvider, useVariables };
