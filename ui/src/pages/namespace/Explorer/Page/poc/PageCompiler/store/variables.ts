import { SynchronousVariableNamespace } from "../../schema/primitives/variable";
import { create } from "zustand";

type VariableNamespace = SynchronousVariableNamespace;
type VariableId = string;
type VariableContent = Array<unknown>;

type Variables = Record<VariableNamespace, Record<VariableId, VariableContent>>;

interface VariableStore {
  variables: Variables;
  actions: {
    setVariable: ({
      namespace,
      id,
      content,
    }: {
      namespace: VariableNamespace;
      id: VariableId;
      content: VariableContent;
    }) => void;
  };
}

const useVariableStore = create<VariableStore>()((set) => ({
  variables: {
    loop: {},
  },
  actions: {
    setVariable: ({ namespace, id, content }) =>
      set((prev) => ({
        ...prev,
        variables: {
          ...prev.variables,
          [namespace]: {
            ...prev.variables[namespace],
            [id]: content,
          },
        },
      })),
  },
}));

export const useVariables = () => useVariableStore((state) => state.variables);

export const useVariableActions = () =>
  useVariableStore((state) => state.actions);
