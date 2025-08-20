import { PropsWithChildren, createContext, useContext } from "react";

import { GlobalVariableNamespace } from "../../../schema/primitives/variable";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;
export type Variable = Record<VariableId, DefinedValue>;

export type GlobalVariableScope = {
  [keys in GlobalVariableNamespace]: Variable;
};

const defaultState: GlobalVariableScope = {
  loop: {},
  query: {},
};

const VariableContext = createContext<GlobalVariableScope | null>(null);

type VariableContextProviderProps = PropsWithChildren<{
  variables: GlobalVariableScope;
}>;

export const VariableContextProvider = ({
  children,
  variables,
}: VariableContextProviderProps) => (
  <VariableContext.Provider value={variables}>
    {children}
  </VariableContext.Provider>
);

export const useGlobalVariableScope = () => {
  const context = useContext(VariableContext);
  const variables = context ?? defaultState;
  return variables;
};
