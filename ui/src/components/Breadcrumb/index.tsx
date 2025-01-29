import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import { FileRoutesById } from "~/routeTree.gen";
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import { useMatches } from "@tanstack/react-router";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const matches = useMatches();

  const match = (routeId: keyof FileRoutesById) =>
    matches.some((match) => match.routeId.startsWith(routeId));

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      {/* <NamespaceSelector /> */}
      {match("/n/$namespace/explorer/") && <ExplorerBreadcrumb />}
      {/* {isInstancePage && <InstancesBreadcrumb />} */}
      {match("/n/$namespace/services/") && <ServicesBreadcrumb />}
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
