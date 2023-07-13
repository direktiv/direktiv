import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import FieldSubMenu from "./FieldSubMenu";
import { SelectFieldMenu } from "./SelectFieldMenu";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export type FilterField = "AS" | "STATUS" | "TRIGGER" | "AFTER" | "BEFORE";

type FilterItem = {
  type: "MATCH" | "CONTAINS";
  value: string;
};

export type FiltersObj = {
  [key in FilterField]?: FilterItem;
};

const Filters = ({ onUpdate }: { onUpdate: (filters: FiltersObj) => void }) => {
  const { t } = useTranslation();
  const [selectedField, setSelectedField] = useState<FilterField | undefined>();
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [filters, setFilters] = useState<FiltersObj>({});

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

  const setFilter = (filterObj: FiltersObj) => {
    const newFilters = { ...filters, ...filterObj };
    setFilters(newFilters);
    resetMenu();
    onUpdate(newFilters);
  };

  const clearFilter = (field: FilterField) => {
    const newFilters = { ...filters };
    delete newFilters[field];
    setFilters(newFilters);
    onUpdate(newFilters);
  };

  const hasFilters = !!Object.keys(filters).length;

  const definedFilters = Object.keys(filters) as Array<FilterField>;

  return (
    <div className="m-2 flex flex-row gap-2">
      {definedFilters.map((field) => (
        <ButtonBar key={field}>
          <Popover>
            <Button variant="outline">{field}</Button>
            <PopoverTrigger asChild>
              <Button variant="outline">{filters[field]?.value}</Button>
            </PopoverTrigger>
            <PopoverContent align="start">
              <FieldSubMenu
                field={field}
                value={filters[field]?.value}
                setFilter={setFilter}
                clearFilter={clearFilter}
              />
            </PopoverContent>
            <Button variant="outline" icon>
              <X onClick={() => clearFilter(field)} />
            </Button>
          </Popover>
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
              {t("pages.instances.list.filter.filterButton")}
            </Button>
          )}
        </PopoverTrigger>
        <PopoverContent align="start">
          {selectedField === undefined ? (
            <SelectFieldMenu onSelect={setSelectedField} />
          ) : (
            <FieldSubMenu
              field={selectedField}
              value={filters[selectedField]?.value}
              setFilter={setFilter}
              clearFilter={clearFilter}
            />
          )}
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default Filters;
