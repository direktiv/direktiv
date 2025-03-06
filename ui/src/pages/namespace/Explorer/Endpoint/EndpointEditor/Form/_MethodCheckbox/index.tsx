import { ControllerRenderProps, UseFormReturn } from "react-hook-form";

import Badge from "~/design/Badge";
import { Checkbox } from "~/design/Checkbox";
import { CheckedState } from "@radix-ui/react-checkbox";
import { EndpointFormSchemaType } from "../../schema";
import { PreviewDialog } from "./PreviewDialog";
import { RouteMethod } from "~/api/gateway/schema";
import { useState } from "react";

interface MethodCheckboxProps {
  isChecked: boolean;
  method: RouteMethod;
  field: ControllerRenderProps<EndpointFormSchemaType>;
  form: UseFormReturn<EndpointFormSchemaType>;
}

const defaultMethodValue = {
  responses: { "200": { description: "" } },
} as const;

const isDefaultValue = (value: unknown) =>
  JSON.stringify(value) === JSON.stringify(defaultMethodValue);

export const MethodCheckbox: React.FC<MethodCheckboxProps> = ({
  method,
  field,
  isChecked,
  form,
}) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const currentValue = form.watch(method);
  const onCheckedChange = (checked: CheckedState) => {
    if (checked) {
      field.onChange(defaultMethodValue);
    } else {
      if (!isDefaultValue(currentValue)) {
        setDialogOpen(true);
      } else {
        field.onChange(undefined);
      }
    }
  };

  return (
    <label className="flex items-center gap-2 text-sm" htmlFor={method}>
      <Checkbox
        id={method}
        checked={isChecked}
        onCheckedChange={onCheckedChange}
      />
      <Badge variant={isChecked ? undefined : "secondary"}>{method}</Badge>
      <PreviewDialog
        field={field}
        form={form}
        method={method}
        open={dialogOpen}
        setOpen={setDialogOpen}
      />
    </label>
  );
};
