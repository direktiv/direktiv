import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();
  const { isServicePage } = pages.services.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
      {isInstancePage && <InstancesBreadcrumb />}
      {isServicePage && <ServicesBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
