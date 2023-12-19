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
  CommandStaticItem,
  CommandStaticSeparator,
} from "~/design/Command";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import Button from "~/design/Button";
import NamespaceCreate from "../NamespaceEdit";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useTranslation } from "react-i18next";

export type ButtonProps = {
  variant?: "onlySelectable";
  onValueChange?: (value: string) => void;
};

const NamespaceSelector: FC<ButtonProps> = ({ variant, onValueChange }) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const {
    data: availableNamespaces,
    isLoading,
    isSuccess,
  } = useListNamespaces();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [open, setOpen] = useState(false);
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  if (!namespace) return null;

  const hasResults = isSuccess && availableNamespaces?.results.length > 0;

  const onNameSpaceChange = (namespace: string) => {
    setNamespace(namespace);
    navigate(pages.explorer.createHref({ namespace }));
  };

  return (
    <div>
      {!variant && (
        <BreadcrumbLink noArrow>
          <Link
            to={pages.explorer.createHref({ namespace })}
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
                  <CommandInput
                    placeholder={t("components.breadcrumb.searchPlaceholder")}
                  />
                  {hasResults && (
                    <CommandList className="max-h-[278px]">
                      <CommandEmpty>
                        {t("components.breadcrumb.notFound")}
                      </CommandEmpty>
                      <CommandGroup>
                        {availableNamespaces?.results.map((ns) => (
                          <CommandItem
                            key={ns.name}
                            value={ns.name}
                            onSelect={(currentValue: string) => {
                              onNameSpaceChange(currentValue);
                              setOpen(false);
                            }}
                          >
                            <Circle
                              className={twMergeClsx(
                                "mr-2 h-2 w-2 fill-current",
                                namespace === ns.name
                                  ? "opacity-100"
                                  : "opacity-0"
                              )}
                            />
                            <span>{ns.name}</span>
                          </CommandItem>
                        ))}
                      </CommandGroup>
                    </CommandList>
                  )}
                  {isLoading && (
                    <CommandStaticItem>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {t("components.breadcrumb.loading")}
                    </CommandStaticItem>
                  )}
                  <CommandStaticSeparator />
                  <DialogTrigger data-testid="new-namespace">
                    <CommandStaticItem>
                      <>
                        <PlusCircle className="mr-2 h-4 w-4" />
                        <span>{t("components.breadcrumb.createBtn")}</span>
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
      )}
      {variant === "onlySelectable" && (
        <Select onValueChange={onValueChange}>
          <SelectTrigger variant="outline">
            <SelectValue
              placeholder={t("components.breadcrumb.selectPlaceholder")}
            />
          </SelectTrigger>
          {isLoading && (
            <SelectContent>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              {t("components.breadcrumb.loading")}
            </SelectContent>
          )}
          {hasResults && (
            <SelectContent>
              {availableNamespaces?.results.map((ns) => (
                <SelectItem key={ns.name} value={ns.name}>
                  <span>{ns.name}</span>
                </SelectItem>
              ))}
            </SelectContent>
          )}
        </Select>
      )}
    </div>
  );
};
export default NamespaceSelector;
