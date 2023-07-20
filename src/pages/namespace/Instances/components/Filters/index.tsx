import { BaseSyntheticEvent, useState } from "react";
import { Command, CommandGroup, CommandList } from "~/design/Command";
import { Plus, X } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import DatePicker from "./DatePicker";
import { FiltersObj } from "~/api/instances/query/get";
import Input from "~/design/Input";
import Options from "./Options";
import { SelectFieldMenu } from "./SelectFieldMenu";
import TextInput from "./TextInput";
import moment from "moment";
import { useTranslation } from "react-i18next";

type FiltersProps = {
  filters: FiltersObj;
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

const Filters = ({ filters, onUpdate }: FiltersProps) => {
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

  const setFilter = (newFilter: FiltersObj) => {
    const newFilters = { ...filters, ...newFilter };
    onUpdate(newFilters);
    resetMenu();
  };

  const clearFilter = (field: keyof FiltersObj) => {
    const newFilters = { ...filters };
    delete newFilters[field];
    onUpdate(newFilters);
  };

  const setTimeOnDate = (
    event: BaseSyntheticEvent,
    field: "AFTER" | "BEFORE",
    date: Date
  ) => {
    const [hr, min, sec] = event.target.value.split(":");
    date.setHours(hr);
    date.setMinutes(min);
    date.setSeconds(sec);
    setFilter({
      [field]: { type: "MATCH", value: date },
    });
  };

  const hasFilters = !!Object.keys(filters).length;

  const definedFilters = Object.keys(filters) as Array<keyof FiltersObj>;

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
                <Button variant="outline">{filters[field]?.value}</Button>
              </PopoverTrigger>
              <PopoverContent align="start">
                {field === "AS" && (
                  <TextInput
                    field={field}
                    setFilter={setFilter}
                    clearFilter={clearFilter}
                    value={filters[field]?.value}
                  />
                )}
                {field === "STATUS" && (
                  <Options
                    field={field}
                    value={filters[field]?.value}
                    setFilter={setFilter}
                  />
                )}
                {field === "TRIGGER" && (
                  <Options
                    field={field}
                    value={filters[field]?.value}
                    setFilter={setFilter}
                  />
                )}
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
                    {moment(filters[field]?.value).format("YYYY-MM-DD")}
                  </Button>
                </PopoverTrigger>
                <PopoverContent align="start">
                  {(field === "AFTER" || field === "BEFORE") && (
                    <DatePicker
                      field={field}
                      date={filters[field]?.value}
                      setFilter={setFilter}
                    />
                  )}
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
                    {moment(filters[field]?.value).format("HH:mm:ss")}
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
                          onChange={(event) => {
                            const date = filters[field]?.value;
                            if (!date) {
                              return;
                            }
                            setTimeOnDate(event, field, date);
                          }}
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
          {(selectedField === null && (
            <SelectFieldMenu onSelect={setSelectedField} />
          )) ||
            (selectedField === "AS" && (
              <TextInput
                field={selectedField}
                setFilter={setFilter}
                clearFilter={clearFilter}
                value={filters[selectedField]?.value}
              />
            )) ||
            (selectedField === "STATUS" && (
              <Options
                field={selectedField}
                value={filters[selectedField]?.value}
                setFilter={setFilter}
              />
            )) ||
            (selectedField === "TRIGGER" && (
              <Options
                field={selectedField}
                value={filters[selectedField]?.value}
                setFilter={setFilter}
              />
            )) ||
            ((selectedField === "AFTER" || selectedField === "BEFORE") && (
              <DatePicker
                field={selectedField}
                date={filters[selectedField]?.value}
                setFilter={setFilter}
              />
            ))}
        </PopoverContent>
      </Popover>
    </div>
  );
};

export default Filters;
