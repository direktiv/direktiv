import * as React from "react";

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
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
  CommandStaticItem,
  CommandStaticSeparator,
} from "./index";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "../Dropdown";
import type { Meta, StoryObj } from "@storybook/react";
import { Popover, PopoverContent, PopoverTrigger } from "../Popover";
import Badge from "../Badge";
import Button from "../Button";
import { Card } from "../Card";
import { twMergeClsx } from "~/util/helpers";

const meta = {
  title: "Components/Command",
  component: Command,
} satisfies Meta<typeof Command>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="flex items-center space-x-2">
      <Command>
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Suggestions">
            <CommandItem value="Calendar">Calendar</CommandItem>
            <CommandItem value="Search Emoji">Search Emoji</CommandItem>
            <CommandItem value="Calculator">Calculator</CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Settings">
            <CommandItem value="Profile">Profile</CommandItem>
            <CommandItem value="Billing">Billing</CommandItem>
            <CommandItem value="Settings">Settings</CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    </div>
  ),
};

export const CommandDemo = () => (
  <Command className="rounded-lg border border-gray-4 shadow-md dark:border-gray-dark-4">
    <CommandInput placeholder="Type a command or search..." />
    <CommandList className="max-h-[200px]">
      <CommandEmpty>No results found.</CommandEmpty>
      <CommandGroup heading="Suggestions">
        <CommandItem value="Calendar">
          <Calendar className="mr-2 h-auto w-4" />
          <span>Calendar</span>
        </CommandItem>
        <CommandItem value="Search Emoji">
          <Smile className="mr-2 h-auto w-4" />
          <span>Search Emoji</span>
        </CommandItem>
        <CommandItem value="Calculator">
          <Calculator className="mr-2 h-auto w-4" />
          <span>Calculator</span>
        </CommandItem>
      </CommandGroup>
      <CommandSeparator />
      <CommandGroup heading="Settings">
        <CommandItem value="Profile">
          <User className="mr-2 h-auto w-4" />
          <span>Profile</span>
          <CommandShortcut>⌘P</CommandShortcut>
        </CommandItem>
        <CommandItem>
          <CreditCard className="mr-2 h-auto w-4" />
          <span>Billing</span>
          <CommandShortcut>⌘B</CommandShortcut>
        </CommandItem>
        <CommandItem value="Settings">
          <Settings className="mr-2 h-auto w-4" />
          <span>Settings</span>
          <CommandShortcut>⌘S</CommandShortcut>
        </CommandItem>
      </CommandGroup>
      <CommandStaticSeparator />
      <CommandStaticItem>
        <CreditCard className="mr-2 h-auto w-4" />
        <span>Static Item</span>
        <CommandShortcut>⌘T</CommandShortcut>
      </CommandStaticItem>
    </CommandList>
  </Command>
);

export function CommandDialogDemo() {
  const [open, setOpen] = React.useState(false);

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "j" && e.metaKey) {
        setOpen((open) => !open);
      }
    };

    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  return (
    <>
      <p className="text-sm text-gray-11 dark:text-gray-dark-11">
        Press{" "}
        <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border border-gray-4 bg-gray-1 px-1.5 font-mono text-[10px] font-medium opacity-100 dark:border-gray-dark-4 dark:bg-gray-dark-1">
          <span className="text-xs">⌘</span>J
        </kbd>
      </p>
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Suggestions">
            <CommandItem value="Calendar">
              <Calendar className="mr-2 h-auto w-4" />
              <span>Calendar</span>
            </CommandItem>
            <CommandItem value="Search Emoji">
              <Smile className="mr-2 h-auto w-4" />
              <span>Search Emoji</span>
            </CommandItem>
            <CommandItem value="Calculator">
              <Calculator className="mr-2 h-auto w-4" />
              <span>Calculator</span>
            </CommandItem>
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Settings">
            <CommandItem value="Profile">
              <User className="mr-2 h-auto w-4" />
              <span>Profile</span>
              <CommandShortcut>⌘P</CommandShortcut>
            </CommandItem>
            <CommandItem value="Billing">
              <CreditCard className="mr-2 h-auto w-4" />
              <span>Billing</span>
              <CommandShortcut>⌘B</CommandShortcut>
            </CommandItem>
            <CommandItem value="Settings">
              <Settings className="mr-2 h-auto w-4" />
              <span>Settings</span>
              <CommandShortcut>⌘S</CommandShortcut>
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </>
  );
}

const frameworks = [
  {
    value: "next.js",
    label: "Next.js",
  },
  {
    value: "sveltekit",
    label: "SvelteKit",
  },
  {
    value: "nuxt.js",
    label: "Nuxt.js",
  },
  {
    value: "remix",
    label: "Remix",
  },
  {
    value: "astro",
    label: "Astro",
  },
];

export function CommandCombobox() {
  const [open, setOpen] = React.useState(false);
  const [value, setValue] = React.useState("");

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between"
        >
          {value
            ? frameworks.find((framework) => framework.value === value)?.label
            : "Select framework..."}
          <ChevronsUpDown className="h-auto w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px]">
        <Command>
          <CommandInput placeholder="Search framework..." />
          <CommandEmpty>No framework found.</CommandEmpty>
          <CommandGroup>
            {frameworks.map((framework) => (
              <CommandItem
                value={framework.value}
                key={framework.value}
                onSelect={(currentValue) => {
                  setValue(currentValue === value ? "" : currentValue);
                  setOpen(false);
                }}
              >
                <Check
                  className={twMergeClsx(
                    "mr-2 h-auto w-4",
                    value === framework.value ? "opacity-100" : "opacity-0"
                  )}
                />
                {framework.label}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

type Status = {
  value: string;
  label: string;
  icon: typeof HelpCircle;
};

const statuses: Status[] = [
  {
    value: "backlog",
    label: "Backlog",
    icon: HelpCircle,
  },
  {
    value: "todo",
    label: "Todo",
    icon: Circle,
  },
  {
    value: "in progress",
    label: "In Progress",
    icon: ArrowUpCircle,
  },
  {
    value: "done",
    label: "Done",
    icon: CheckCircle2,
  },
  {
    value: "canceled",
    label: "Canceled",
    icon: XCircle,
  },
];

export function CommandPopover() {
  const [open, setOpen] = React.useState(false);
  const [selectedStatus, setSelectedStatus] = React.useState<Status | null>(
    null
  );

  return (
    <div className="flex items-center space-x-4">
      <p className="text-sm">Status</p>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button variant="outline" className="justify-start">
            {selectedStatus ? (
              <>
                <selectedStatus.icon />
                {selectedStatus.label}
              </>
            ) : (
              <>
                <PlusCircle /> Set status
              </>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="p-0" side="bottom" align="start">
          <Command>
            <CommandInput placeholder="Change status..." />
            <CommandList>
              <CommandEmpty>No results found.</CommandEmpty>
              <CommandGroup>
                {statuses.map((status) => (
                  <CommandItem
                    value={status.value}
                    key={status.value}
                    onSelect={(value) => {
                      setSelectedStatus(
                        statuses.find((priority) => priority.value === value) ||
                          null
                      );
                      setOpen(false);
                    }}
                  >
                    <status.icon className={twMergeClsx("mr-2 h-auto w-4")} />
                    <span>{status.label}</span>
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}

const labels = [
  "feature",
  "bug",
  "enhancement",
  "documentation",
  "design",
  "question",
  "maintenance",
];

export function CommandDropdownMenu() {
  const [label, setLabel] = React.useState("feature");
  const [open, setOpen] = React.useState(false);

  return (
    <Card
      noShadow
      className="flex w-full flex-col items-start justify-between p-3 sm:flex-row sm:items-center"
    >
      <div className="text-sm">
        <Badge>{label}</Badge> Create a new project
      </div>
      <DropdownMenu open={open} onOpenChange={setOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="sm">
            <MoreHorizontal />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-[200px]">
          <DropdownMenuLabel>Actions</DropdownMenuLabel>
          <DropdownMenuGroup>
            <DropdownMenuItem>
              <User className="mr-2 h-auto w-4" />
              Assign to...
            </DropdownMenuItem>
            <DropdownMenuItem>
              <Calendar className="mr-2 h-auto w-4" />
              Set due date...
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuSub>
              <DropdownMenuSubTrigger>
                <Tags className="mr-2 h-auto w-4" />
                Apply label
              </DropdownMenuSubTrigger>
              <DropdownMenuSubContent className="p-0">
                <Command>
                  <CommandInput
                    placeholder="Filter label..."
                    autoFocus={true}
                  />
                  <CommandList>
                    <CommandEmpty>No label found.</CommandEmpty>
                    <CommandGroup>
                      {labels.map((label) => (
                        <CommandItem
                          value={label}
                          key={label}
                          onSelect={(value) => {
                            setLabel(value);
                            setOpen(false);
                          }}
                        >
                          {label}
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </DropdownMenuSubContent>
            </DropdownMenuSub>
            <DropdownMenuSeparator />
            <DropdownMenuItem className="text-danger-9 dark:text-danger-dark-9">
              <Trash className="mr-2 h-auto w-4" />
              Delete
              <DropdownMenuShortcut>⌘⌫</DropdownMenuShortcut>
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </Card>
  );
}
