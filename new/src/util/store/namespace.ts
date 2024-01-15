import { create } from "zustand";
import { persist } from "zustand/middleware";

interface NamespaceState {
  namespace: string | null;
  actions: {
    setNamespace: (namespace: NamespaceState["namespace"]) => void;
  };
}

const useNamespaceState = create<NamespaceState>()(
  persist(
    (set) => ({
      namespace: null,
      actions: {
        setNamespace: (newNamespace) =>
          set(() => ({ namespace: newNamespace })),
      },
    }),
    {
      name: "direktiv-store-namespace",
      partialize: (state) => ({
        namespace: state.namespace, // pick all fields to be persistent and don't persist actions
      }),
    }
  )
);

export const useNamespace = () => useNamespaceState((state) => state.namespace);
export const useNamespaceActions = () =>
  useNamespaceState((state) => state.actions);
