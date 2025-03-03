import { ChevronDown, ClipboardPaste } from "lucide-react";
import { ControllerRenderProps, UseFormReturn } from "react-hook-form";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { RouteMethod, routeMethods } from "~/api/gateway/schema";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Checkbox } from "~/design/Checkbox";
import { CheckedState } from "@radix-ui/react-checkbox";
import { EndpointFormSchemaType } from "../schema";
import { OpenAPIDocsEditor } from "./OpenAPIDocsEditor";
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

export const MethodCheckbox: React.FC<MethodCheckboxProps> = ({
  method,
  field,
  isChecked,
  form,
}) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const previousValue = form.watch(method);
  const onCheckedChange = (checked: CheckedState) => {
    if (checked) {
      field.onChange(defaultMethodValue);
    } else {
      const isCustomValue =
        JSON.stringify(previousValue) !== JSON.stringify(defaultMethodValue);
      if (isCustomValue) {
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
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Are you sure?</DialogTitle>
          </DialogHeader>
          You are about to delete the documentation for the {method}.
          <OpenAPIDocsEditor
            defaultValue={{
              connect: form.getValues("connect"),
              delete: form.getValues("delete"),
              get: form.getValues("get"),
              head: form.getValues("head"),
              options: form.getValues("options"),
              patch: form.getValues("patch"),
              post: form.getValues("post"),
              put: form.getValues("put"),
              trace: form.getValues("trace"),
            }}
            readOnly
          />
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="ghost">
                Cancel
              </Button>
            </DialogClose>
            <ButtonBar>
              <Button
                type="button"
                onClick={() => {
                  form.setValue("patch", previousValue);
                  field.onChange(undefined);
                  setDialogOpen(false);
                }}
              >
                Delete Documentation
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button icon>
                    <ChevronDown />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-40">
                  {Array.from(routeMethods).map((method) => (
                    <DropdownMenuItem
                      key={method}
                      onClick={() => {
                        form.setValue(method, previousValue);
                        field.onChange(undefined);
                        setDialogOpen(false);
                      }}
                    >
                      <ClipboardPaste className="mr-2 size-4" /> copy to{" "}
                      {method}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </ButtonBar>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </label>
  );
};
