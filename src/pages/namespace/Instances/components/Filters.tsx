import {
  Command,
  CommandGroup,
  CommandInput,
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

type FilterField = "as" | "status" | "trigger" | "after" | "before";

type FilterItem = {
  type: string;
  value: string;
};

type Filters = {
  [key in FilterField]?: FilterItem;
};

const ParamSelect = ({
  onSelect,
}: {
  onSelect: (value: FilterField) => void;
}) => (
  <Command>
    <CommandInput placeholder="Type a command or search..." />
    <CommandList>
      <CommandGroup heading="Select filter">
        <CommandItem tabIndex={1} onSelect={() => onSelect("as")}>
          by name
        </CommandItem>
        <CommandItem onSelect={() => onSelect("status")}>by state</CommandItem>
        <CommandItem onSelect={() => onSelect("trigger")}>
          by invoker
        </CommandItem>
        <CommandItem onSelect={() => onSelect("after")}>
          created after
        </CommandItem>
        <CommandItem onSelect={() => onSelect("before")}>
          created before
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
);

const Filters = () => {
  const [selectedField, setSelectedField] = useState<FilterField | undefined>();
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [filters, setFilters] = useState<Filters>({});
  const [inputValue, setInputValue] = useState<string>("");

  const handleOpenChange = (isOpening: boolean) => {
    if (!isOpening) {
      setSelectedField(undefined);
    }
    setIsOpen(isOpening);
  };

  const resetMenu = () => {
    setIsOpen(false);
    setSelectedField(undefined);
  };

  const handleKeyDown = (event: { key: string }) => {
    if (event.key === "Enter" && inputValue) {
      setFilter({
        as: { value: inputValue, type: "contain" },
      });
    }
    if (event.key === "Enter" && !inputValue) {
      clearFilter("as");
    }
  };

  const setFilter = (filterObj: Filters) => {
    setFilters({ ...filters, ...filterObj });
    resetMenu();
  };

  const clearFilter = (field: FilterField) => {
    const newFilters = { ...filters };
    delete newFilters[field];
    setFilters(newFilters);
  };

  const hasFilters = !!Object.keys(filters).length;

  const definedFields = Object.keys(filters) as Array<FilterField>;

  return (
    <div className="m-2 flex flex-row gap-2">
      {definedFields.map((field) => (
        <ButtonBar key={field}>
          <Button variant="outline">{field}</Button>
          <Button variant="outline">{filters[field]?.value}</Button>
          <Button variant="outline" icon>
            <X onClick={() => clearFilter(field)} />
          </Button>
        </ButtonBar>
      ))}

      <Popover open={isOpen} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          {hasFilters ? (
            <Button variant="outline" icon>
              <Plus />
            </Button>
          ) : (
            <Button variant="outline">
              <Plus />
              Filter
            </Button>
          )}
        </PopoverTrigger>
        <PopoverContent align="start">
          {selectedField === undefined && (
            <ParamSelect onSelect={setSelectedField} />
          )}
          {selectedField === "as" && (
            <Command>
              <CommandList>
                <CommandGroup heading="Filter by name">
                  <Input
                    autoFocus
                    placeholder="filename.yaml"
                    value={inputValue}
                    onChange={(event) => setInputValue(event.target.value)}
                    onKeyUp={handleKeyDown}
                  />
                </CommandGroup>
              </CommandList>
            </Command>
          )}
          {selectedField === "status" && (
            <Command>
              <CommandInput
                autoFocus
                placeholder="Type a command or search..."
              />
              <CommandList>
                <CommandGroup heading="Filter by state">
                  <CommandItem
                    value="pending"
                    onSelect={() =>
                      setFilter({
                        status: { value: "pending", type: "match" },
                      })
                    }
                  >
                    Running
                  </CommandItem>
                  <CommandItem
                    value="complete"
                    onSelect={() =>
                      setFilter({
                        status: { value: "complete", type: "match" },
                      })
                    }
                  >
                    Complete
                  </CommandItem>
                  <CommandItem
                    value="cancelled"
                    onSelect={() =>
                      setFilter({
                        status: { value: "cancelled", type: "match" },
                      })
                    }
                  >
                    Cancelled
                  </CommandItem>
                  <CommandItem
                    value="failed"
                    onSelect={() =>
                      setFilter({
                        status: { value: "failed", type: "match" },
                      })
                    }
                  >
                    Failed
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          )}
          {selectedField === "trigger" && (
            <Command>
              <CommandInput
                autoFocus
                placeholder="Type a command or search..."
              />
              <CommandList>
                <CommandGroup heading="Filter by invoker">
                  <CommandItem
                    value="api"
                    onSelect={() =>
                      setFilter({
                        trigger: { value: "api", type: "match" },
                      })
                    }
                  >
                    API
                  </CommandItem>
                  <CommandItem
                    value="cloudevent"
                    onSelect={() =>
                      setFilter({
                        trigger: { value: "cloudevent", type: "match" },
                      })
                    }
                  >
                    Cloud event
                  </CommandItem>
                  <CommandItem
                    value="instance"
                    onSelect={() =>
                      setFilter({
                        trigger: { value: "instance", type: "match" },
                      })
                    }
                  >
                    Instance
                  </CommandItem>
                  <CommandItem
                    value="cron"
                    onSelect={() =>
                      setFilter({
                        trigger: { value: "cron", type: "match" },
                      })
                    }
                  >
                    Cron
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          )}
          {selectedField === "after" && (
            <Command>
              <CommandList className="max-h-[460px]">
                <CommandGroup heading="Filter created after">
                  <Datepicker />
                </CommandGroup>
              </CommandList>
            </Command>
          )}
          {selectedField === "before" && (
            <Command>
              <CommandList className="max-h-[460px]">
                <CommandGroup heading="Filter created before">
                  <Datepicker />
                </CommandGroup>
              </CommandList>
            </Command>
          )}
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default Filters;
