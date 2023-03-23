import {
  Breadcrumb as BreadcrumbLink,
  BreadcrumbRoot,
} from "../../componentsNext/Breadcump";
import {
  ChevronsUpDown,
  FolderOpen,
  Home,
  Loader2,
  Play,
  PlusCircle,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../componentsNext/Dropdown";
import { Link, useNavigate } from "react-router-dom";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import Button from "../../componentsNext/Button";
import { FC } from "react";
import { analyzePath } from "../../util/router/utils";
import clsx from "clsx";
import { pages } from "../../util/router/pages";
import { useNamespaces } from "../../api/namespaces";
import { useTree } from "../../api/tree";

const BreadcrumbSegment: FC<{ absolute: string; relative: string }> = ({
  absolute,
  relative,
}) => {
  const namespace = useNamespace();

  const { data, isLoading } = useTree({
    path: absolute,
  });

  if (!namespace) return null;

  let Icon = FolderOpen;

  if (data?.node.expandedType === "workflow") {
    Icon = Play;
  }

  const link =
    data?.node.expandedType === "workflow"
      ? pages.workflow.createHref({ namespace, path: absolute })
      : pages.explorer.createHref({ namespace, path: absolute });

  return (
    <BreadcrumbLink>
      <Link to={link} className="gap-2">
        <Icon aria-hidden="true" className={clsx(isLoading && "invisible")} />
        {relative}
      </Link>
    </BreadcrumbLink>
  );
};

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { data: availableNamespaces, isLoading } = useNamespaces();

  const { path: pathParamsExplorer } = pages.explorer.useParams();
  const { path: pathParamsWorkflow } = pages.workflow.useParams();

  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  if (!namespace) return null;

  const path = analyzePath(pathParamsExplorer || pathParamsWorkflow);

  const onNameSpaceChange = (namespace: string) => {
    setNamespace(namespace);
    navigate(pages.explorer.createHref({ namespace }));
  };
  return (
    <BreadcrumbRoot>
      <BreadcrumbLink>
        <Link to={pages.explorer.createHref({ namespace })} className="gap-2">
          <Home />
          {namespace}
        </Link>
        &nbsp;
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button size="sm" variant="ghost" circle>
              <ChevronsUpDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent className="w-56">
            <DropdownMenuLabel>Namespaces</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuRadioGroup
              value={namespace}
              onValueChange={onNameSpaceChange}
            >
              {availableNamespaces?.results.map((ns) => (
                <DropdownMenuRadioItem
                  key={ns.name}
                  value={ns.name}
                  textValue={ns.name}
                >
                  {ns.name}
                </DropdownMenuRadioItem>
              ))}
            </DropdownMenuRadioGroup>
            {isLoading && (
              <DropdownMenuItem disabled>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                loading...
              </DropdownMenuItem>
            )}
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <PlusCircle className="mr-2 h-4 w-4" />
              <span>Create new namespace</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </BreadcrumbLink>

      {path.segments.map((x) => (
        <BreadcrumbSegment
          key={x.absolute}
          absolute={x.absolute}
          relative={x.relative}
        />
      ))}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
