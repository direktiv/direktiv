import { useNamespace } from "~/util/store/namespace";

const gatewayNamespaceName = "gateway_namespace";

/**
 * this hook is used to check if the the current namespace
 * is the special gatway namespace. It returns undefined
 * if the namespace is not known (yet)
 */

const useIsGatewayAvailable = () => {
  const namespace = useNamespace();

  if (!namespace) {
    return undefined;
  }

  return namespace === gatewayNamespaceName;
};

export default useIsGatewayAvailable;
