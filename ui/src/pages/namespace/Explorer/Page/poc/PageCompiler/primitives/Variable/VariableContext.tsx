import { PropsWithChildren, createContext, useContext } from "react";

import { ContextVariableNamespace } from "../../../schema/primitives/variable";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;
export type Variable = Record<VariableId, DefinedValue>;

export type ContextVariables = {
  [keys in ContextVariableNamespace]: Variable;
};

const defaultState: ContextVariables = {
  loop: {},
  query: {},
};

const VariableContext = createContext<ContextVariables | null>(null);

type VariableContextProviderProps = PropsWithChildren<{
  variables: ContextVariables;
}>;

export const VariableContextProvider = ({
  children,
  variables,
}: VariableContextProviderProps) => (
  <VariableContext.Provider value={variables}>
    {children}
  </VariableContext.Provider>
);

export const useVariablesContext = () => {
  const context = useContext(VariableContext);
  const variables = context ?? defaultState;
  return variables;
};
