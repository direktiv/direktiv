import {
  Breadcrumb as BreadcrumbLink,
  BreadcrumbRoot,
} from "../../componentsNext/Breadcump";
import {
  ChevronsUpDown,
  FolderOpen,
  Github,
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
} from "@radix-ui/react-dropdown-menu";
import { Link, useNavigate } from "react-router-dom";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import Button from "../../componentsNext/Button";
import { FC } from "react";
import clsx from "clsx";
import { pages } from "../../util/router/pages";
import { useNamespaces } from "../../api/namespaces";
import { useTree } from "../../api/tree";

const BreadcrumbComponent: FC<{ path: string }> = ({ path }) => {
  // split path string in to chunks, using the last / as the separator
  const segments = path.split("/");
  const namespace = useNamespace();

  const { data, isLoading } = useTree({
    directory: path,
  });

  if (!namespace) return null;

  let Icon = FolderOpen;

  if (data?.node.expandedType === "git") {
    Icon = Github;
  }

  if (data?.node.expandedType === "directory") {
    Icon = FolderOpen;
  }

  if (data?.node.expandedType === "workflow") {
    Icon = Play;
  }

  const link =
    data?.node.expandedType === "workflow"
      ? pages.workflow.createHref({ namespace, file: path })
      : pages.explorer.createHref({ namespace, directory: path });

  return (
    <BreadcrumbLink>
      <Link to={link} className="gap-2">
        <Icon aria-hidden="true" className={clsx(isLoading && "invisible")} />
        {segments.slice(-1)}
      </Link>
    </BreadcrumbLink>
  );
};

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { data: availableNamespaces, isLoading } = useNamespaces();
  const { directory } = pages.explorer.useParams();
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  if (!namespace) return null;

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
      {/* TODO: extract this into a util and write some tests */}
      {directory &&
        directory?.split("/").map((segment, index, srcArr) => {
          const absolutePath = srcArr.slice(0, index + 1).join("/");
          return <BreadcrumbComponent key={absolutePath} path={absolutePath} />;
        })}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
