import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot>
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
      {isInstancePage && <InstancesBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
