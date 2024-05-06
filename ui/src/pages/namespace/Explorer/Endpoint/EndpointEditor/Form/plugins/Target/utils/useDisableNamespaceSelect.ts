import { useNamespaceDetail } from "~/api/namespaces/query/get";

export const useDisableNamespaceSelect = () => {
  const { data: namespaces } = useNamespaceDetail();
  const disableNamespaceSelect = !namespaces?.isSystemNamespace;
  return disableNamespaceSelect;
};
