import { create } from "zustand";

interface NamespaceState {
  namespace: string | null;
  actions: {
    setNamespace: (namespace: NamespaceState["namespace"]) => void;
  };
}

const useNamespaceState = create<NamespaceState>((set) => ({
  namespace: null,
  actions: {
    setNamespace: (newNamespace) => set(() => ({ namespace: newNamespace })),
  },
}));

export const useNamespace = () => useNamespaceState((state) => state.namespace);
export const useNamespaceActions = () =>
  useNamespaceState((state) => state.actions);
