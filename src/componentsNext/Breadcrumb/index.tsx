import { BreadcrumbRoot } from "~/design/Breadcrumbs";
import ExplorerBreadcrumb from "./ExplorerBreadcrumb";
import NamespaceSelector from "./NamespaceSelector";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { isExplorerPage } = pages.explorer.useParams();

  if (!namespace) return null;

  return (
    <BreadcrumbRoot>
      <NamespaceSelector />
      {isExplorerPage && <ExplorerBreadcrumb />}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
