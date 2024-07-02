import { useNamespaceDetail } from "~/api/namespaces/query/get";

export const useIsSystemNamespace = () => {
  const { data: namespaces } = useNamespaceDetail();
  const isSystemNamespace = !!namespaces?.isSystemNamespace;
  return isSystemNamespace;
};
