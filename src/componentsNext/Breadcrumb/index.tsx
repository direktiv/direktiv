import {
  ArrowUpCircle,
  Calculator,
  Calendar,
  Check,
  CheckCircle2,
  ChevronsUpDown,
  Circle,
  CreditCard,
  HelpCircle,
  Home,
  Loader2,
  MoreHorizontal,
  PlusCircle,
  Settings,
  Smile,
  Tags,
  Trash,
  User,
  XCircle,
} from "lucide-react";
import {
  Breadcrumb as BreadcrumbLink,
  BreadcrumbRoot,
} from "../../design/Breadcrumbs";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandSeparator,
  CommandShortcut,
} from "../../design/Command";
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

const Breadcrumb = () => {
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
        {/* <Dialog open={dialogOpen} onOpenChange={setDialogOpen}> */}
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger asChild>
            <Button size="sm" variant="ghost" circle>
              <ChevronsUpDown />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-56 p-0">
            <Command>
              <CommandInput placeholder="Search namespace..." />
              <CommandEmpty>No framework found.</CommandEmpty>
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
                {/* <CommandItem>
                  <Calendar className="mr-2 h-4 w-4" />
                  <span>Calendar</span>
                </CommandItem>
                <CommandItem>
                  <Smile className="mr-2 h-4 w-4" />
                  <span>Search Emoji</span>
                </CommandItem>
                <CommandItem>
                  <Calculator className="mr-2 h-4 w-4" />
                  <span>Calculator</span>
                </CommandItem> */}
              </CommandGroup>
              <CommandSeparator />
              <CommandGroup heading="Settings">
                <CommandItem>
                  <User className="mr-2 h-4 w-4" />
                  <span>Profile</span>
                  <CommandShortcut>⌘P</CommandShortcut>
                </CommandItem>
                <CommandItem>
                  <CreditCard className="mr-2 h-4 w-4" />
                  <span>Billing</span>
                  <CommandShortcut>⌘B</CommandShortcut>
                </CommandItem>
                <CommandItem>
                  <Settings className="mr-2 h-4 w-4" />
                  <span>Settings</span>
                  <CommandShortcut>⌘S</CommandShortcut>
                </CommandItem>
              </CommandGroup>
              {isLoading && (
                <CommandItem disabled>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  loading...
                </CommandItem>
              )}
              <CommandSeparator />
              {/* <DialogTrigger> */}
              <CommandGroup>
                <CommandItem>
                  <>
                    <PlusCircle className="mr-2 h-4 w-4" />
                    <span>Create new namespace</span>
                  </>
                </CommandItem>
              </CommandGroup>
              {/* </DialogTrigger> */}
            </Command>
          </PopoverContent>
        </Popover>

        {/* <DialogContent>
          <NamespaceCreate close={() => setDialogOpen(false)} />
        </DialogContent> */}
        {/* </Dialog> */}
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
