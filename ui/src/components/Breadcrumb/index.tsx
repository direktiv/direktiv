import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import { useMatch } from "@tanstack/react-router";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const isExplorerPage = useMatch({
    from: "/n/$namespace/explorer",
    shouldThrow: false,
  });
  const isServicePage = useMatch({
    from: "/n/$namespace/services",
    shouldThrow: false,
  });

  // const { isInstancePage } = pages.instances.useParams();
  // const { isServicePage } = pages.services.useParams();
  // const { isEventsHistoryPage, isEventsListenersPage } =
  //   pages.events.useParams();
  // const { isMonitoringPage } = pages.monitoring.useParams();
  // const { isPermissionsPage } = pages.permissions?.useParams() ?? {};
  // const { isSettingsPage } = pages.settings.useParams();
  // const { isJqPlaygroundPage } = pages.jqPlayground.useParams();
  // const { isMirrorPage } = pages.mirror.useParams();
  // const { isGatewayPage } = pages.gateway.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      {/* <NamespaceSelector /> */}
      {isExplorerPage && <ExplorerBreadcrumb />}
      {/* {isInstancePage && <InstancesBreadcrumb />} */}
      {isServicePage && <ServicesBreadcrumb />}
      {/* {isEventsHistoryPage && <EventHistoryBreadcrumb />}
      {isEventsListenersPage && <EventListenerBreadcrumb />}
      {isMonitoringPage && <MonitoringBreadcrumb />}
      {isPermissionsPage && <PermissionsBreadcrumb />}
      {isSettingsPage && <SettingsBreadcrumb />}
      {isJqPlaygroundPage && <JqPlaygroundBreadcrumb />}
      {isMirrorPage && <MirrorBreadcrumb />}
      {isGatewayPage && <GatewayBreadcrumb />} */}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
