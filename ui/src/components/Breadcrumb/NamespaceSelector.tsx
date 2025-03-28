import { ChevronsUpDown, Home, PlusCircle } from "lucide-react";
import {
  Command,
  CommandStaticItem,
  CommandStaticSeparator,
} from "~/design/Command";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Link, useNavigate } from "@tanstack/react-router";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import Button from "~/design/Button";
import NamespaceEdit from "../NamespaceEdit";
import { NamespaceSelectorList } from "../NamespaceSelectorList";
import { useNamespace } from "~/util/store/namespace";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const NamespaceSelector = () => {
  const { t } = useTranslation();
  const namespace = useNamespace();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();

  if (!namespace) return null;

  const onNameSpaceChange = (namespace: string) => {
    setOpen(false);
    navigate({ to: "/n/$namespace/explorer", params: { namespace } });
  };

  return (
    <BreadcrumbLink noArrow>
      <Link
        to="/n/$namespace/explorer"
        params={{ namespace }}
        data-testid="breadcrumb-namespace"
      >
        <Home />
        {namespace}
      </Link>
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button
              size="sm"
              variant="ghost"
              circle
              data-testid="dropdown-trg-namespace"
            >
              <ChevronsUpDown />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-56 p-0">
            <Command>
              <NamespaceSelectorList onSelectNamespace={onNameSpaceChange} />
              <CommandStaticSeparator />
              <DialogTrigger data-testid="new-namespace">
                <CommandStaticItem>
                  <>
                    <PlusCircle className="mr-2 size-4" />
                    <span>{t("components.breadcrumb.createBtn")}</span>
                  </>
                </CommandStaticItem>
              </DialogTrigger>
            </Command>
          </PopoverContent>
        </Popover>
        <DialogContent>
          <NamespaceEdit close={() => setDialogOpen(false)} />
        </DialogContent>
      </Dialog>
    </BreadcrumbLink>
  );
};

export default NamespaceSelector;
