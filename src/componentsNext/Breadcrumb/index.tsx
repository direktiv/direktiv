import {
  Breadcrumb as BreadcrumbLink,
  BreadcrumbRoot,
} from "../../design/Breadcrumbs";
import { ChevronsUpDown, Home, Loader2, PlusCircle } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "../../design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../design/Dropdown";
import { Link, useNavigate } from "react-router-dom";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import BreadcrumbSegment from "./BreadcrumbSegment";
import Button from "../../design/Button";
import NamespaceCreate from "../NamespaceCreate";
import { analyzePath } from "../../util/router/utils";
import { pages } from "../../util/router/pages";
import { useListNamespaces } from "../../api/namespaces/query/get";
import { useState } from "react";

const Breadcrumb = () => {
  const namespace = useNamespace();
  const { data: availableNamespaces, isLoading } = useListNamespaces();
  const [dialogOpen, setDialogOpen] = useState(false);

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
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
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
              <DialogTrigger>
                <DropdownMenuItem>
                  <PlusCircle className="mr-2 h-4 w-4" />
                  <span>Create new namespace</span>
                </DropdownMenuItem>
              </DialogTrigger>
            </DropdownMenuContent>
          </DropdownMenu>
          <DialogContent>
            <NamespaceCreate close={() => setDialogOpen(false)} />
          </DialogContent>
        </Dialog>
      </BreadcrumbLink>
      {path.segments.map((x, i) => (
        <BreadcrumbSegment
          key={x.absolute}
          absolute={x.absolute}
          relative={x.relative}
          isLast={i === path.segments.length - 1}
        />
      ))}
    </BreadcrumbRoot>
  );
};

export default Breadcrumb;
