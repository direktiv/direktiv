import { useRef } from "react";
import { useVariableActions } from "../../store/variables";

/**
 * this hook initializes the loop variable in the store to makes it available
 * in the global zustand store.
 */
export function useInitLoopVariable(
  id: string,
  variableContent: unknown[] | undefined
) {
  const isInitialized = useRef(false);
  const variableActions = useVariableActions();
  if (variableContent && !isInitialized.current) {
    isInitialized.current = true;
    variableActions.setVariable({
      namespace: "loop",
      id,
      content: variableContent,
    });
  }
}
