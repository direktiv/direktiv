import { BaseSyntheticEvent, useState } from "react";
import { Command, CommandGroup, CommandList } from "~/design/Command";
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

type MenuAnchor =
  | "main"
  | "AS"
  | "STATUS"
  | "TRIGGER"
  | "AFTER"
  | "BEFORE"
  | "AFTER.time"
  | "BEFORE.time";

const Filters = ({ value, onUpdate }: FiltersProps) => {
  const { t } = useTranslation();

  // activeMenu controls which popover component is opened (there are
  // separate popovers triggered by the respective buttons)
  const [activeMenu, setActiveMenu] = useState<MenuAnchor | null>(null);

  // selectedField controls which submenu is shown in the main menu
  const [selectedField, setSelectedField] = useState<keyof FiltersObj | null>(
    null
  );

  const handleOpenChange = (isOpening: boolean, menu: MenuAnchor) => {
    if (!isOpening) {
      setSelectedField(null);
    }
    toggleMenu(menu);
  };

  const toggleMenu = (value: MenuAnchor) => {
    if (activeMenu === value) {
      return setActiveMenu(null);
    }
    setActiveMenu(value);
  };

  const resetMenu = () => {
    setActiveMenu(null);
    setSelectedField(null);
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
    const [hr, min, sec] = event.target.value.split(":");
    const newFilters = { ...value };

    newFilters[field]?.value.setHours(hr);
    newFilters[field]?.value.setMinutes(min);
    newFilters[field]?.value.setSeconds(sec);

    onUpdate(newFilters);
    resetMenu();
  };

  const hasFilters = !!Object.keys(value).length;

  const definedFilters = Object.keys(value) as Array<keyof FiltersObj>;

  return (
    <div className="m-2 flex flex-row gap-2">
      {definedFilters.map((field) => (
        <ButtonBar key={field}>
          {(field === "AS" || field === "TRIGGER" || field === "STATUS") && (
            <Popover
              open={activeMenu === field}
              onOpenChange={(state) => handleOpenChange(state, field)}
            >
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
                  setFilter={(value) => setFilter(value)}
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
              <Popover
                open={activeMenu === field}
                onOpenChange={(state) => handleOpenChange(state, field)}
              >
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(value[field]?.value).format("YYYY-MM-DD")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  <FieldSubMenu
                    field={field}
                    date={value[field]?.value}
                    setFilter={(value) => setFilter(value)}
                    clearFilter={clearFilter}
                  />
                </PopoverContent>
              </Popover>
              <Popover
                open={activeMenu === `${field}.time`}
                onOpenChange={(state) =>
                  handleOpenChange(state, `${field}.time`)
                }
              >
                <PopoverTrigger asChild>
                  <Button variant="outline" className="px-2">
                    {moment(value[field]?.value).format("HH:mm:ss")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  <Command>
                    <CommandList className="max-h-[460px]">
                      <CommandGroup
                        heading={t(
                          "pages.instances.list.filter.menuHeading.time"
                        )}
                      >
                        <Input
                          type="time"
                          step={1}
                          onChange={(event) => setTime(event, field)}
                        />
                      </CommandGroup>
                    </CommandList>
                  </Command>
                </PopoverContent>
              </Popover>
              <Button variant="outline" icon>
                <X onClick={() => clearFilter(field)} />
              </Button>
            </>
          )}
        </ButtonBar>
      ))}
      <Popover
        open={activeMenu === "main"}
        onOpenChange={(state) => handleOpenChange(state, "main")}
      >
        <PopoverTrigger asChild>
          {hasFilters ? (
            <Button variant="outline" icon onClick={() => toggleMenu("main")}>
              <Plus />
            </Button>
          ) : (
            <Button variant="outline" onClick={() => toggleMenu("main")}>
              <Plus />
              {t("pages.instances.list.filter.filterButton")}
            </Button>
          )}
        </PopoverTrigger>
        <PopoverContent align="start">
          {selectedField === null ? (
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
