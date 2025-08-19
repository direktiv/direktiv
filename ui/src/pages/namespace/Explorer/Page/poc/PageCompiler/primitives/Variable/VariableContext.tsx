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

export type FormVariables = Variables["form"];

/**
 * returns a function to get the current variables context.
 * the function allows to inject form variables into the results
 */
const useGetVariables = () => {
  const context = useContext(VariableContext);
  const variables = context ?? defaultVariables;
  return (formVariables?: FormVariables): Variables => {
    if (!formVariables) return variables;
    return {
      ...variables,
      form: {
        ...variables.form,
        ...formVariables,
      },
    };
  };
};

/**
 * returns the current variables of the context.
 */
const useVariables = () => {
  const getVariables = useGetVariables();
  return getVariables();
};

export { VariableContextProvider, useGetVariables, useVariables };
