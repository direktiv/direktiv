import {
  GlobalVariableScope,
  Variable,
  useGlobalVariableScope,
} from "./VariableContext";

import { LocalVariableNamespace } from "../../../schema/primitives/variable";

type LocalVariableScope = Record<LocalVariableNamespace, Variable>;
type VariableScope = GlobalVariableScope & LocalVariableScope;

export type LocalVariables = VariableScope["this"];

/**
 * returns a function that allows to inject local
 * variables into the global variables object.
 */
export const useCreateVariableInjection = () => {
  const globalVariables = useGlobalVariableScope();
  return (localVariables?: LocalVariables): VariableScope => ({
    ...globalVariables,
    this: { ...(localVariables ?? {}) },
  });
};
