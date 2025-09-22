import {
  LocalVariableNamespace,
  Variable,
  contextVariableNamespaces,
} from "../../../schema/primitives/variable";
import { PropsWithChildren, createContext, useContext } from "react";

import z from "zod";

type VariableId = string;
type DefinedValue = Exclude<unknown, undefined>;
type Variable = Record<VariableId, DefinedValue>;

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

export type LocalVariables = Record<LocalVariableNamespace, Variable>;
type LocalAndContextVariables = ContextVariables & LocalVariables;

export type LocalVariablesContent =
  LocalAndContextVariables[LocalVariableNamespace];
