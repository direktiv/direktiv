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
import { useTranslation } from "react-i18next";

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
  const { t } = useTranslation();
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

  const otherMethods = Array.from(routeMethods).filter((m) => m !== method);

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
            <DialogTitle>
              {t("pages.explorer.endpoint.editor.form.methods.modal.headline")}
            </DialogTitle>
          </DialogHeader>
          {t("pages.explorer.endpoint.editor.form.methods.modal.description", {
            method,
          })}
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
                {t(
                  "pages.explorer.endpoint.editor.form.methods.modal.cancelBtn"
                )}
              </Button>
            </DialogClose>
            <ButtonBar>
              <Button
                type="button"
                onClick={() => {
                  field.onChange(undefined);
                  setDialogOpen(false);
                }}
              >
                {t(
                  "pages.explorer.endpoint.editor.form.methods.modal.confirmBtn"
                )}
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button icon>
                    <ChevronDown />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-64">
                  {otherMethods.map((method) => {
                    const overwrite =
                      form.watch(method) && !isDefaultValue(form.watch(method));

                    return (
                      <DialogClose
                        key={method}
                        onClick={() => {
                          form.setValue(method, currentValue);
                          field.onChange(undefined);
                          setDialogOpen(false);
                        }}
                        className="w-full"
                      >
                        <DropdownMenuItem>
                          <ClipboardPaste className="mr-2 size-4" />
                          {t(
                            overwrite
                              ? "pages.explorer.endpoint.editor.form.methods.modal.overwrite"
                              : "pages.explorer.endpoint.editor.form.methods.modal.copy",
                            { method }
                          )}
                        </DropdownMenuItem>
                      </DialogClose>
                    );
                  })}
                </DropdownMenuContent>
              </DropdownMenu>
            </ButtonBar>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </label>
  );
};
