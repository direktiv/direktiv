import {
  ContextVariables,
  Variable,
  useVariablesContext,
} from "./VariableContext";

import { LocalVariableNamespace } from "../../../schema/primitives/variable";

type LocalVariables = Record<LocalVariableNamespace, Variable>;
type LocalAndContextVariables = ContextVariables & LocalVariables;

export type LocalVariablesContent = LocalAndContextVariables["this"];

/**
 * returns a function that accepts local variables, and returns
 * a merged object of context variables and local variables.
 */
export const useMergeLocalWithContextVariables = () => {
  const contextVariables = useVariablesContext();
  return (
    localVariables?: LocalVariablesContent
  ): LocalAndContextVariables => ({
    ...contextVariables,
    this: { ...(localVariables ?? {}) },
  });
};
