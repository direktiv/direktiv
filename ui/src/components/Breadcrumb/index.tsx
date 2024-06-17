import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import EventHistoryBreadcrumb from "./Events/HistoryBreadcrumb";
import EventListenerBreadcrumb from "./Events/ListenerBreadcrumb";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import GatewayBreadcrumb from "./Gateway";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import JqPlaygroundBreadcrumb from "./JqPlaygroundBreadcrumb";
import MirrorBreadcrumb from "./MirrorBreadcrumb";
import MonitoringBreadcrumb from "./MonitoringBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import PermissionsBreadcrumb from "./Permissions";
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import SettingsBreadcrumb from "./SettingsBreadcrumb";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";

const Breadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();
  const { isServicePage } = pages.services.useParams();
  const { isEventsHistoryPage, isEventsListenersPage } =
    pages.events.useParams();
  const { isMonitoringPage } = pages.monitoring.useParams();
  const { isPermissionsPage } = pages.permissions?.useParams() ?? {};
  const { isSettingsPage } = pages.settings.useParams();
  const { isJqPlaygroundPage } = pages.jqPlayground.useParams();
  const { isMirrorPage } = pages.mirror.useParams();
  const { isGatewayPage } = pages.gateway.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
      {isInstancePage && <InstancesBreadcrumb />}
      {isServicePage && <ServicesBreadcrumb />}
      {isEventsHistoryPage && <EventHistoryBreadcrumb />}
      {isEventsListenersPage && <EventListenerBreadcrumb />}
      {isMonitoringPage && <MonitoringBreadcrumb />}
      {isPermissionsPage && <PermissionsBreadcrumb />}
      {isSettingsPage && <SettingsBreadcrumb />}
      {isJqPlaygroundPage && <JqPlaygroundBreadcrumb />}
      {isMirrorPage && <MirrorBreadcrumb />}
      {isGatewayPage && <GatewayBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
