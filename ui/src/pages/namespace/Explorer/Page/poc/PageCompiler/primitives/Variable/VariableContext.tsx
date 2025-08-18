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

export type InjectedVariables = Partial<Variables>;

/**
 * returns a function to get the current variables context.
 * the function allows to inject variables into the results
 */
const useGetVariables = () => {
  const context = useContext(VariableContext);
  const variables = context ?? defaultVariables;
  return (injectedVariables?: InjectedVariables): Variables => {
    if (!injectedVariables) return variables;
    return {
      ...variables,
      form: {
        ...variables.form,
        ...injectedVariables.form,
      },
      loop: {
        ...variables.loop,
        ...injectedVariables.loop,
      },
      query: {
        ...variables.query,
        ...injectedVariables.query,
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
