import { BaseSyntheticEvent, useState } from "react";
import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import FieldSubMenu from "./FieldSubMenu";
import { FiltersObj } from "~/api/instances/query/get";
import Input from "~/design/Input";
import { SelectFieldMenu } from "./SelectFieldMenu";
import moment from "moment";
import { useTranslation } from "react-i18next";

type FiltersProps = {
  value: FiltersObj;
  onUpdate: (filters: FiltersObj) => void;
};

const Filters = ({ value, onUpdate }: FiltersProps) => {
  const { t } = useTranslation();
  const [selectedField, setSelectedField] = useState<
    keyof FiltersObj | undefined
  >();
  const [isOpen, setIsOpen] = useState<boolean>(false);

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
    const newFilters = { ...value, ...filterObj };
    onUpdate(newFilters);
    resetMenu();
  };

  const clearFilter = (field: keyof FiltersObj) => {
    const newFilters = { ...value };
    delete newFilters[field];
    onUpdate(newFilters);
  };

  const setTime = (event: BaseSyntheticEvent, field: "AFTER" | "BEFORE") => {
    const [hr, min] = event.target.value.split(":");
    const newFilters = { ...value };

    newFilters[field]?.value.setHours(hr);
    newFilters[field]?.value.setMinutes(min);

    onUpdate(newFilters);
  };

  const hasFilters = !!Object.keys(value).length;

  const definedFilters = Object.keys(value) as Array<keyof FiltersObj>;

  return (
    <div className="m-2 flex flex-row gap-2">
      {definedFilters.map((field) => (
        <ButtonBar key={field}>
          {(field === "AS" || field === "TRIGGER" || field === "STATUS") && (
            <Popover>
              <Button variant="outline">
                {t([`pages.instances.list.filter.field.${field}`])}
              </Button>
              <PopoverTrigger asChild>
                <Button variant="outline">{value[field]?.value}</Button>
              </PopoverTrigger>
              <PopoverContent align="start">
                <FieldSubMenu
                  field={field}
                  value={value[field]?.value}
                  setFilter={setFilter}
                  clearFilter={clearFilter}
                />
              </PopoverContent>
              <Button variant="outline" icon>
                <X onClick={() => clearFilter(field)} />
              </Button>
            </Popover>
          )}
          {(field === "BEFORE" || field === "AFTER") && (
            <>
              <Button variant="outline">
                {t([`pages.instances.list.filter.field.${field}`])}
              </Button>
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(value[field]?.value).format("YYYY-MM-DD")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  <FieldSubMenu
                    field={field}
                    date={value[field]?.value}
                    setFilter={setFilter}
                    clearFilter={clearFilter}
                  />
                </PopoverContent>
              </Popover>
              <Popover>
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(value[field]?.value).format("HH:mm:ss")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  <Input
                    type="time"
                    step={1}
                    onChange={(event) => setTime(event, field)}
                  />
                </PopoverContent>
              </Popover>
              <Button variant="outline" icon>
                <X onClick={() => clearFilter(field)} />
              </Button>
            </>
          )}
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
            ((selectedField === "AS" ||
              selectedField === "TRIGGER" ||
              selectedField === "STATUS") && (
              <FieldSubMenu
                field={selectedField}
                value={value[selectedField]?.value}
                setFilter={setFilter}
                clearFilter={clearFilter}
              />
            )) ||
            ((selectedField === "BEFORE" || selectedField === "AFTER") && (
              <FieldSubMenu
                field={selectedField}
                date={value[selectedField]?.value}
                setFilter={setFilter}
                clearFilter={clearFilter}
              />
            ))
          )}
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default Filters;
