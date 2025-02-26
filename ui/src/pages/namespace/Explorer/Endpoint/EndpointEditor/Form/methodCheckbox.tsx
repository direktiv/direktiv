import Badge from "~/design/Badge";
import { Checkbox } from "~/design/Checkbox";
import { ControllerRenderProps } from "react-hook-form";
import { EndpointFormSchemaType } from "../schema";
import { RouteMethod } from "~/api/gateway/schema";
import { useState } from "react";

interface MethodCheckboxProps {
  method: RouteMethod;
  field: ControllerRenderProps<EndpointFormSchemaType>;
}

export const MethodCheckbox: React.FC<MethodCheckboxProps> = ({
  method,
  field,
}) => {
  const [isChecked, setIsChecked] = useState(!!field.value);

  return (
    <label className="flex items-center gap-2 text-sm" htmlFor={method}>
      <Checkbox
        id={method}
        checked={isChecked}
        onCheckedChange={(checked) => {
          if (checked === "indeterminate") return;
          setIsChecked(checked);
          if (checked) {
            field.onChange({ responses: { "200": { description: "" } } });
          } else {
            field.onChange(undefined);
          }
        }}
      />
      <Badge variant={isChecked ? undefined : "secondary"}>{method}</Badge>
    </label>
  );
};
