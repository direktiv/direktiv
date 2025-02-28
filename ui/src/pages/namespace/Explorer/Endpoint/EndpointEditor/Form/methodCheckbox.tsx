import { ControllerRenderProps, UseFormReturn } from "react-hook-form";

import Badge from "~/design/Badge";
import { Checkbox } from "~/design/Checkbox";
import { EndpointFormSchemaType } from "../schema";
import { RouteMethod } from "~/api/gateway/schema";

interface MethodCheckboxProps {
  isChecked: boolean;
  method: RouteMethod;
  field: ControllerRenderProps<EndpointFormSchemaType>;
  form: UseFormReturn<EndpointFormSchemaType>;
}

export const MethodCheckbox: React.FC<MethodCheckboxProps> = ({
  method,
  field,
  isChecked,
}) => (
  <label className="flex items-center gap-2 text-sm" htmlFor={method}>
    <Checkbox
      id={method}
      checked={isChecked}
      onCheckedChange={(checked) => {
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
