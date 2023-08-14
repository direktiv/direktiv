import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import MonitoringBreadcrumb from "./MonitoringBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();
  const { isMonitoringPage } = pages.monitoring.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
      {isInstancePage && <InstancesBreadcrumb />}
      {isMonitoringPage && <MonitoringBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
