import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import EventHistoryBreadcrumb from "./Events/HistoryBreadcrumb";
import EventListenerBreadcrumb from "./Events/ListenerBreadcrumb";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import JqPlaygroundBreadcrumb from "./JqPlaygroundBreadcrumb";
import MonitoringBreadcrumb from "./MonitoringBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import SettingsBreadcrumb from "./SettingsBreadcrumb";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();
  const { isServicePage } = pages.services.useParams();
  const { isEventsHistoryPage, isEventsListenersPage } =
    pages.events.useParams();
  const { isMonitoringPage } = pages.monitoring.useParams();
  const { isSettingsPage } = pages.settings.useParams();
  const { isJqPlaygroundPage } = pages.jqPlayground.useParams();

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
      {isSettingsPage && <SettingsBreadcrumb />}
      {isJqPlaygroundPage && <JqPlaygroundBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
