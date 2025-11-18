import { useCallback, useRef } from "react";

import { LocalVariables } from "../../primitives/Variable/VariableContext";
import { extractFormKeys } from "../formPrimitives/utils";

export const useRegisterForm = (
  register?: (localVariables: LocalVariables) => void
) => {
  const observerRef = useRef<MutationObserver | null>(null);
  const lastVariablesRef = useRef<string | null>(null);

  const registerForm = useCallback(
    (form: HTMLFormElement | null): void => {
      if (!register) {
        return;
      }

      if (!form) {
        observerRef.current?.disconnect();
        observerRef.current = null;
        return;
      }

      const updateVariables = () => {
        const localVariables = extractFormKeys(form.elements);
        const serialized = JSON.stringify(localVariables);
        if (lastVariablesRef.current !== serialized) {
          lastVariablesRef.current = serialized;
          register({ this: localVariables });
        }
      };

      updateVariables();

      const observer = new MutationObserver(() => updateVariables());
      observer.observe(form, { childList: true, subtree: true });
      observerRef.current = observer;
    },
    [register]
  );

  return registerForm;
};
