import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import EventsBreadcrumb from "./EventsBreadcrumb";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import InstancesBreadcrumb from "./InstancesBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();
  const { isInstancePage } = pages.instances.useParams();
  const { isEventsPage } = pages.events.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot className="group">
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
      {isInstancePage && <InstancesBreadcrumb />}
      {isEventsPage && <EventsBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
