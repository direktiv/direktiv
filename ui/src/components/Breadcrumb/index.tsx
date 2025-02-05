import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import { FileRoutesById } from "~/routeTree.gen";
import GatewayBreadcrumb from "./Gateway";
import HistoryBreadcrumb from "./Events/HistoryBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import JqPlaygroundBreadcrumb from "./JqPlaygroundBreadcrumb";
import ListenerBreadcrumb from "./Events/ListenerBreadcrumb";
import MirrorBreadcrumb from "./MirrorBreadcrumb";
import MonitoringBreadcrumb from "./MonitoringBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import SettingsBreadcrumb from "./SettingsBreadcrumb";
import { useMatches } from "@tanstack/react-router";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const matches = useMatches();

  const matchRouteStart = (routeId: keyof FileRoutesById) =>
    matches.some((match) => match.routeId.startsWith(routeId));

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      <NamespaceSelector />
      {matchRouteStart("/n/$namespace/explorer") && <ExplorerBreadcrumb />}
      {matchRouteStart("/n/$namespace/instances/") && <InstancesBreadcrumb />}
      {matchRouteStart("/n/$namespace/services/") && <ServicesBreadcrumb />}
      {matchRouteStart("/n/$namespace/events/history") && <HistoryBreadcrumb />}
      {matchRouteStart("/n/$namespace/events/listeners") && (
        <ListenerBreadcrumb />
      )}
      {matchRouteStart("/n/$namespace/monitoring") && <MonitoringBreadcrumb />}
      {/* { && <PermissionsBreadcrumb />} */}
      {matchRouteStart("/n/$namespace/settings") && <SettingsBreadcrumb />}
      {matchRouteStart("/n/$namespace/jq") && <JqPlaygroundBreadcrumb />}
      {matchRouteStart("/n/$namespace/mirror/") && <MirrorBreadcrumb />}
      {matchRouteStart("/n/$namespace/gateway/") && <GatewayBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
