import {
  Breadcrumb as BreadcrumbLink,
  BreadcrumbRoot,
} from "../../design/Breadcrumbs";
import {
  ChevronsUpDown,
  Circle,
  Home,
  Loader2,
  PlusCircle,
} from "lucide-react";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandStaticItem,
} from "../../design/Command";
import { Dialog, DialogContent, DialogTrigger } from "../../design/Dialog";
import { Link, useNavigate } from "react-router-dom";
import { Popover, PopoverContent, PopoverTrigger } from "../../design/Popover";
import React, { useState } from "react";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";

import BreadcrumbSegment from "./BreadcrumbSegment";
import Button from "../../design/Button";
import NamespaceCreate from "../NamespaceCreate";
import { analyzePath } from "../../util/router/utils";
import clsx from "clsx";
import { pages } from "../../util/router/pages";
import { useListNamespaces } from "../../api/namespaces/query/get";
import { useTranslation } from "react-i18next";

const Breadcrumb = () => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const { data: availableNamespaces, isLoading } = useListNamespaces();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [open, setOpen] = React.useState(false);

  const { path: pathParams } = pages.explorer.useParams();

  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  if (!namespace) return null;

  const path = analyzePath(pathParams);

  const onNameSpaceChange = (namespace: string) => {
    setNamespace(namespace);
    navigate(pages.explorer.createHref({ namespace }));
  };
  return (
    <BreadcrumbRoot>
      <BreadcrumbLink noArrow>
        <Link to={pages.explorer.createHref({ namespace })}>
          <Home />
          {namespace}
        </Link>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
              <Button size="sm" variant="ghost" circle>
                <ChevronsUpDown />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-56 p-0">
              <Command>
                <CommandInput
                  placeholder={t(
                    "pages.explorer.breadcrumbs.searchPlaceholder"
                  )}
                />
                <CommandList>
                  <CommandEmpty>
                    {t("pages.explorer.breadcrumbs.notFound")}
                  </CommandEmpty>
                  <CommandGroup>
                    {availableNamespaces?.results.map((ns) => (
                      <CommandItem
                        key={ns.name}
                        onSelect={(currentValue: string) => {
                          onNameSpaceChange(currentValue);
                          setOpen(false);
                        }}
                      >
                        <Circle
                          className={clsx(
                            "mr-2 h-2 w-2 fill-current",
                            namespace === ns.name ? "opacity-100" : "opacity-0"
                          )}
                        />
                        <span>{ns.name}</span>
                      </CommandItem>
                    ))}
                  </CommandGroup>
                </CommandList>
                <CommandSeparator />
                {isLoading && (
                  <CommandItem disabled>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    {t("pages.explorer.breadcrumbs.loading")}
                  </CommandItem>
                )}
                <CommandSeparator />
                <DialogTrigger>
                  <CommandStaticItem>
                    <>
                      <PlusCircle className="mr-2 h-4 w-4" />
                      <span>
                        {t("pages.explorer.breadcrumbs.createButton")}
                      </span>
                    </>
                  </CommandStaticItem>
                </DialogTrigger>
              </Command>
            </PopoverContent>
          </Popover>

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
