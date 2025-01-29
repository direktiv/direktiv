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
import ServicesBreadcrumb from "./ServicesBreadcrumb";
import SettingsBreadcrumb from "./SettingsBreadcrumb";
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
      {match("/n/$namespace/instances/") && <InstancesBreadcrumb />}
      {match("/n/$namespace/services/") && <ServicesBreadcrumb />}
      {match("/n/$namespace/events/history") && <HistoryBreadcrumb />}
      {match("/n/$namespace/events/listeners") && <ListenerBreadcrumb />}
      {match("/n/$namespace/monitoring") && <MonitoringBreadcrumb />}
      {/* { && <PermissionsBreadcrumb />} */}
      {match("/n/$namespace/settings") && <SettingsBreadcrumb />}
      {match("/n/$namespace/jq") && <JqPlaygroundBreadcrumb />}
      {match("/n/$namespace/mirror/") && <MirrorBreadcrumb />}
      {match("/n/$namespace/gateway/routes") && <GatewayBreadcrumb />}
      {match("/n/$namespace/gateway/consumers") && <GatewayBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
