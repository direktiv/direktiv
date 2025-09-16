import { PropsWithChildren, createContext, useContext } from "react";
import {
  Variable,
  contextVariableNamespaces,
} from "../../../schema/primitives/variable";

import z from "zod";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;
export type Variable = Record<VariableId, DefinedValue>;

export const ContextVariablesSchema = z.record(
  z.enum(contextVariableNamespaces),
  z.record(z.string(), z.unknown())
);

export type ContextVariables = z.infer<typeof ContextVariablesSchema>;

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
  const variables = context ?? {};
  return variables;
};
