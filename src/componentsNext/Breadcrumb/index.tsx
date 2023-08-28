import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import EventHistoryBreadcrumb from "./Events/HistoryBreadcrumb";
import EventListenerBreadcrumb from "./Events/ListenerBreadcrumb";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import MonitoringBreadcrumb from "./MonitoringBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import PermissionsBreadcrumb from "./PermissionsBreadcrumb";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();
  const { isEventsHistoryPage, isEventsListenersPage } =
    pages.events.useParams();
  const { isMonitoringPage } = pages.monitoring.useParams();
  const { isPermissionsPage } = pages.permissions?.useParams() ?? {};

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
      {isInstancePage && <InstancesBreadcrumb />}
      {isEventsHistoryPage && <EventHistoryBreadcrumb />}
      {isEventsListenersPage && <EventListenerBreadcrumb />}
      {isMonitoringPage && <MonitoringBreadcrumb />}
      {isPermissionsPage && <PermissionsBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
