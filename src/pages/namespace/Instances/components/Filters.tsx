import {
  Command,
  CommandGroup,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Datepicker } from "~/design/Datepicker";
import Input from "~/design/Input";
import { useState } from "react";

const hasFilters = false;

type FilterParam = "name" | "state" | "invoker" | "after" | "before";

const ParamSelect = ({
  onSelect,
}: {
  onSelect: (value: FilterParam) => void;
}) => (
  <CommandList>
    <CommandGroup heading="Select filter">
      <CommandItem onSelect={() => onSelect("name")}>by name</CommandItem>
      <CommandItem onSelect={() => onSelect("state")}>by state</CommandItem>
      <CommandItem onSelect={() => onSelect("invoker")}>by invoker</CommandItem>
      <CommandItem onSelect={() => onSelect("after")}>
        created after
      </CommandItem>
      <CommandItem onSelect={() => onSelect("before")}>
        created before
      </CommandItem>
    </CommandGroup>
  </CommandList>
);

// Mockup
const ExistingFilters = () => (
  <>
    <ButtonBar>
      <Button variant="outline">Type</Button>
      <Button variant="outline">noop.yaml</Button>
      <Button variant="outline" icon>
        <X />
      </Button>
    </ButtonBar>
    <ButtonBar>
      <Button variant="outline">Started after</Button>
      <Button variant="outline">01-Feb-2022</Button>
      <Button variant="outline" icon>
        <X />
      </Button>
    </ButtonBar>
    <ButtonBar>
      <Button variant="outline">Started before</Button>
      <Button variant="outline">01-Feb-2023</Button>
      <Button variant="outline" icon>
        <X />
      </Button>
    </ButtonBar>
  </>
);

const Filters = () => {
  const [param, setParam] = useState<FilterParam | undefined>();
  const [isOpen, setIsOpen] = useState<boolean>(false);

  const handleOpenChange = (isOpening: boolean) => {
    if (!isOpening) {
      setParam(undefined);
    }
    setIsOpen(isOpening);
  };

  return (
    <div className="m-2 flex flex-row gap-2">
      {hasFilters ? (
        <>
          <ExistingFilters />
          <Button variant="outline" icon>
            <Plus />
          </Button>
        </>
      ) : (
        <Popover open={isOpen} onOpenChange={handleOpenChange}>
          <PopoverTrigger asChild>
            <Button variant="outline">
              <Plus />
              Filter
            </Button>
          </PopoverTrigger>
          <PopoverContent align="start">
            <Command>
              {param === undefined && <ParamSelect onSelect={setParam} />}
              {param === "name" && (
                <CommandList>
                  <CommandGroup heading="Filter by name">
                    <Input placeholder="filename.yaml" />
                  </CommandGroup>
                </CommandList>
              )}
              {param === "state" && (
                <CommandList>
                  <CommandGroup heading="Filter by state">
                    <CommandItem value="running">Running</CommandItem>
                    <CommandItem value="complete">Complete</CommandItem>
                    <CommandItem value="cancelled">Cancelled</CommandItem>
                    <CommandItem value="failed">Failed</CommandItem>
                  </CommandGroup>
                </CommandList>
              )}
              {param === "invoker" && (
                <CommandList>
                  <CommandGroup heading="Filter by invoker">
                    <CommandItem value="running">API</CommandItem>
                    <CommandItem value="complete">Cloud event</CommandItem>
                    <CommandItem value="cancelled">Instance</CommandItem>
                    <CommandItem value="failed">Cron</CommandItem>
                  </CommandGroup>
                </CommandList>
              )}
              {param === "after" && (
                <CommandList className="max-h-[460px]">
                  <CommandGroup heading="Filter created after">
                    <Datepicker />
                  </CommandGroup>
                </CommandList>
              )}
              {param === "before" && (
                <CommandList className="max-h-[460px]">
                  <CommandGroup heading="Filter created before">
                    <Datepicker />
                  </CommandGroup>
                </CommandList>
              )}
            </Command>
          </PopoverContent>
        </Popover>
      )}
    </div>
  );
};

export default Filters;
