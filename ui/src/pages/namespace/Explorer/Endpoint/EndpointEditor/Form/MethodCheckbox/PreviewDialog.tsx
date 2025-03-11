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

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { EndpointFormSchemaType } from "../../schema";
import { OpenAPIDocsEditor } from "../OpenAPIDocsEditor";
import { useTranslation } from "react-i18next";

interface PreviewDialogProps {
  method: RouteMethod;
  field: ControllerRenderProps<EndpointFormSchemaType>;
  form: UseFormReturn<EndpointFormSchemaType>;
  open: boolean;
  setOpen: (open: boolean) => void;
}

const defaultMethodValue = {
  responses: { "200": { description: "" } },
} as const;

const isDefaultValue = (value: unknown) =>
  JSON.stringify(value) === JSON.stringify(defaultMethodValue);

export const PreviewDialog: React.FC<PreviewDialogProps> = ({
  method,
  field,
  form,
  open,
  setOpen,
}) => {
  const { t } = useTranslation();

  const currentValue = form.watch(method);
  const otherMethods = Array.from(routeMethods).filter((m) => m !== method);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
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
              {t("pages.explorer.endpoint.editor.form.methods.modal.cancelBtn")}
            </Button>
          </DialogClose>
          <ButtonBar>
            <Button
              type="button"
              onClick={() => {
                field.onChange(undefined);
                setOpen(false);
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
                        setOpen(false);
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
  );
};
